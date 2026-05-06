# 提现审核中间层设计（Withdraw Review Layer）

- 日期：2026-05-06
- 状态：Draft → 待 Review
- 作者：Sovereign Engineering

## 背景与问题

当前用户提现路径（`POST /api/v1/wallets/withdraw`）直接调用 Cobo MPC 钱包发起转账：

1. 校验白名单 / 2FA / 冷却 / 限额
2. 从 `available` 扣除金额，创建 `transactions(type=withdraw, status=pending)`
3. 调用 Cobo `CreateTransferTransaction`
4. 失败则回滚余额、tx 标记 `failed`；成功则 `status=processing`，等待 webhook

**问题：** Cobo MPC 钱包的源地址（`withdraw_addresses[network]`）若余额不足（error_code 12007）或 Cobo 接口任何短暂异常，用户提现就直接失败。最近一次线上事故：Cobo 热钱包仅余 ~0.97 USDT，用户申请 50 USDT 提现，返回 500，用户体验和资金安全感差。

**目标：** 在用户与 Cobo 之间引入审核层。用户提现请求落库为待审核工单；管理员审核通过后再代为调用 Cobo；Cobo 调用过程中出现的非"明确拒绝"异常（如热钱包余额不足、网络错误），不应导致用户提现失败，而是进入可重试状态。

## 范围

**包含：**
- 提现状态机扩展（review_status 新维度）
- 用户侧：提交后转入 frozen，可在 pending_review 阶段取消
- 管理员侧：审核工单列表、批准/拒绝/重试三类操作
- 数据库迁移、API、Service、Admin UI 完整实现
- 通知、审计、测试

**不包含（明确 YAGNI）：**
- 自动审核阈值（小额免审）
- 双人复核 / 提交 Cobo 前再次确认
- 自动 worker 重试（仅手动重试）
- 提现并发限制
- 管理员二次 2FA（首版只用前端弹窗确认）
- 风控规则引擎（如黑名单地址、可疑模式检测）

## 设计决策摘要

| # | 决策 | 选择 |
|---|---|---|
| 1 | 审核范围 | 全部提现一律人工审核 |
| 2 | submit_failed 重试 | 仅管理员手动重试 |
| 3 | 用户取消窗口 | 仅 `pending_review` |
| 4 | 拒绝原因 | 必填，对用户可见 |
| 5 | 审核权限 | `super_admin` + `operator` 均可 |
| 6 | 批准与提交 Cobo | 合并为一步 |
| 7 | 操作二次确认 | 仅前端弹窗 |
| 8 | 通知频率 | 仅最终态（confirmed / rejected / failed） |
| 9 | 余额状态 | `pending_review/submit_failed` 期间金额位于 `frozen` |
| 10 | 24h 冷却判定时点 | 用户提交那一刻 |
| 11 | 并发提现 | 允许 |
| 12 | 数据层 | 复用 `transactions` 表加字段 |

## 状态机

```
                              [user submit]
                                    │
                                    ▼
                            ┌───────────────┐
                            │ pending_review│ ← available -amount, frozen +amount
                            └───────────────┘
                              │       │       │
              user cancel ────┘       │       └──── admin reject(reason)
                    │                 │                    │
                    ▼                 ▼                    ▼
              ┌──────────┐    [admin approve]       ┌──────────┐
              │cancelled │           │              │ rejected │
              │(refund)  │           │              │(refund)  │
              └──────────┘           │              └──────────┘
                                     ▼
                              call Cobo Withdraw
                                  │      │
                       success ◄──┘      └──► HTTP/12007/任何错误
                          │                          │
                          ▼                          ▼
                    ┌──────────┐              ┌──────────────┐
                    │ submitted│              │submit_failed │
                    │ +        │              │(still frozen,│
                    │processing│              │ admin retry) │
                    └──────────┘              └──────────────┘
                          │                          │
                          ▼                          │
                    [Cobo webhook]                   │
                       │      │                      │
                  confirmed  failed (Cobo            │
                   (success) reject only)            │
                                                     │
                          ▲                          │
                          └────── admin retry ───────┘
```

### 不变量

- **资金守恒**：任何状态转换都满足 `Δavailable + Δfrozen + Δoutflow = 0`
- **frozen 释放规则**：
  - `pending_review → cancelled / rejected`：frozen 退回 available
  - `(approved | submit_failed) → submitted`（Cobo 接受后）：frozen 真正扣除（计入 outflow），不退回
  - `submit_failed` 状态下：frozen 不动，等待重试或拒绝
- **submit_failed ≠ 终态**：不发用户通知，不退余额，仅记录 `last_submit_error` 供运维排查
- **终态**：`cancelled | rejected | confirmed | failed`（仅 Cobo webhook 报失败时进入 failed）

## 数据模型

### transactions 表新增字段

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `review_status` | varchar(32) | NULL | 仅 `type='withdraw'` 行有值。枚举：`pending_review / submitted / submit_failed / rejected / cancelled` |
| `reviewed_by` | uuid | FK admin_users.id, NULL | 审核/拒绝/重试操作人 |
| `reviewed_at` | timestamptz | NULL | 最近一次审核动作时间 |
| `reject_reason` | varchar(500) | NULL | 拒绝原因（必填，对用户可见） |
| `submit_attempts` | int | NOT NULL DEFAULT 0 | 已尝试调 Cobo 的次数 |
| `last_submit_error` | text | NULL | 最近一次 Cobo 错误（不展示给用户） |
| `last_submit_at` | timestamptz | NULL | 最近一次尝试调 Cobo 时间 |

### review_status 与现有 status 的映射

| review_status | status | 含义 |
|---|---|---|
| pending_review | pending | 等审核 |
| submit_failed | pending | Cobo 调用失败，等待重试 |
| submitted | processing | 已成功提交 Cobo，等待链上确认 |
| submitted | confirmed | 链上确认（终态成功） |
| submitted | failed | Cobo webhook 明确返回失败（终态失败） |
| rejected | cancelled | 管理员拒绝（终态） |
| cancelled | cancelled | 用户主动撤销（终态） |

> 现有 `status` 字段语义不变，由 Cobo webhook 推动；`review_status` 表达运营工单维度。两者必须始终满足上表映射，违反则视为数据异常需告警。

### 迁移 SQL

```sql
ALTER TABLE transactions
  ADD COLUMN review_status     varchar(32),
  ADD COLUMN reviewed_by       uuid REFERENCES admin_users(id),
  ADD COLUMN reviewed_at       timestamptz,
  ADD COLUMN reject_reason     varchar(500),
  ADD COLUMN submit_attempts   int NOT NULL DEFAULT 0,
  ADD COLUMN last_submit_error text,
  ADD COLUMN last_submit_at    timestamptz;

CREATE INDEX idx_tx_review_status ON transactions(review_status)
  WHERE type = 'withdraw';

-- 历史 withdraw 行回填：视为已是老逻辑直接提交过的
UPDATE transactions
   SET review_status = 'submitted'
 WHERE type = 'withdraw' AND review_status IS NULL;
```

## API

### 用户侧

#### `POST /api/v1/wallets/withdraw`（行为变更）

- **变更前**：扣 available → 创建 tx → 调 Cobo
- **变更后**：
  1. 既有校验（白名单 + 24h 冷却 + 2FA + 限额）保持不变
  2. `available -= amount, frozen += amount`
  3. 创建 `tx(status=pending, review_status=pending_review)`
  4. 写 `audit_logs`
  5. 返回 `tx.id` 与 `review_status`
  6. **不**调 Cobo，**不**发 `WithdrawRequested` 事件

#### `DELETE /api/v1/wallets/withdraw/:id`（新增）

- 仅 tx 属于当前用户、`type=withdraw`、`review_status=pending_review` 时允许
- `available += amount, frozen -= amount`
- `review_status=cancelled, status=cancelled`
- 写 `audit_logs`
- 不发邮件通知

#### `GET /api/v1/wallets/transactions`（字段补充）

响应中提现行追加：`review_status`、`reject_reason`（仅当 `rejected` 时返回）

### 管理后台

所有路径前缀 `/admin`，权限：`super_admin` 或 `operator`。

#### `GET /admin/withdrawals`

Query：`status`（review_status）、`user_id`、`from`、`to`、`page`、`size`
Response：与 admin transactions 列表一致结构 + 用户简要信息（email/name）

#### `POST /admin/withdrawals/:id/approve`

- 校验：`type=withdraw, review_status=pending_review`
- 更新 `reviewed_by/reviewed_at`
- 调 `submitToCobo(tx)`：
  - 成功 → `frozen -= amount`（计入 outflow），`status=processing, review_status=submitted, external_id`，发 `WithdrawRequested` 事件
  - 失败 → `submit_attempts++, last_submit_error, last_submit_at`，`review_status=submit_failed`，**不动余额**，返回 friendly error
- 写 `audit_logs`

#### `POST /admin/withdrawals/:id/reject`

Body: `{"reason": "string, 必填"}`
- 校验：`review_status ∈ {pending_review, submit_failed}`
- `available += amount, frozen -= amount`
- `review_status=rejected, status=cancelled, reject_reason=...`
- 写 `audit_logs`
- 发 `WithdrawRejected` 通知邮件

#### `POST /admin/withdrawals/:id/retry`

- 校验：`review_status=submit_failed`
- 与 approve 共享 `submitToCobo(tx)`，状态流转规则相同
- 写 `audit_logs`

## Service 层改造

### `wallet/service/wallet_service.go`

- `Withdraw` 方法删除当前 ~197-241 行调 Cobo 段
- 余额变更：`available -= amount, frozen += amount`（用 `walletRepo.UpdateBalance`）
- tx 落库 `status=pending, review_status=pending_review`
- 不发 `WithdrawRequested` 事件
- 新增 `CancelWithdraw(ctx, userID, txID) error`

### 新增 `wallet/service/withdraw_review_service.go`

```go
type WithdrawReviewService interface {
    List(ctx context.Context, f ListFilter) ([]Tx, int64, error)
    Approve(ctx context.Context, txID, adminID string) error
    Reject(ctx context.Context, txID, adminID, reason string) error
    Retry(ctx context.Context, txID, adminID string) error
}
```

私有方法：
```go
func (s *withdrawReviewService) submitToCobo(ctx, tx) error
// 共用逻辑：调 Cobo、根据成功/失败更新状态字段。
// 资金扣除（frozen → outflow）只在成功路径里发生；
// 失败路径永远不动 frozen / available。
```

## Admin 前端

### 新页面：`admin/src/pages/Withdrawals/index.tsx`

菜单位置：交易 → 提现审核

- 顶部 Tabs：`待审核 (pending_review)` / `提交失败 (submit_failed)` / `历史 (rejected | cancelled | submitted)`
- 表格列：用户（email + 跳转）、金额、币种/网络、目标地址（标记是否白名单）、申请时间、状态
- 行操作：
  - 待审核：`批准并提交`（弹窗确认）/ `拒绝`（弹窗填原因）
  - 提交失败：`重试提交` / `拒绝退款`，并展示 `submit_attempts` 与 `last_submit_error`
- 详情抽屉：完整链路（申请→审核→提交→链上）+ Cobo external_id 复制

### 既有 Transactions 页

保留，提现行加列 `审核状态`（review_status）。

## 通知

复用现有 `NotificationService`，仅在终态发邮件：

| 事件 | 触发 | 模板要点 |
|---|---|---|
| `WithdrawConfirmed` | Cobo webhook → status=confirmed | 已成功到账，附 tx_hash |
| `WithdrawRejected` | admin reject | 提现被拒，附 `reject_reason`，已退回 |
| `WithdrawFailed` | Cobo webhook → status=failed | Cobo 链上失败，已退回，请联系客服 |

**不发**：申请提交确认、approve 后处理中、submit_failed。

## 审计

所有 admin 动作（approve / reject / retry）写入既有 `admin_audit_logs` 表（actor=admin_user_id / target_type=transaction / target_id / action / metadata，含金额、原因等）。用户 cancel 操作走应用日志即可（不进 admin 审计）。

## 测试计划（TDD）

### Service 层单测

- `TestWithdraw_PendingReview`：提交后 `status=pending, review_status=pending_review`、`frozen` 增加、`available` 减少、未调 Cobo
- `TestWithdraw_CooldownStillEnforced`、`TestWithdraw_2FARequired`：原有规则不变
- `TestApprove_Success`：mock Cobo 成功 → frozen 扣除、状态流转、事件发布
- `TestApprove_CoboInsufficient_KeepsFrozen`：mock Cobo 12007 → review_status=submit_failed、frozen/available 不变、attempts++
- `TestApprove_NetworkError_KeepsFrozen`：mock Cobo HTTP 错误 → 同上
- `TestRetry_FromSubmitFailed_Success`、`TestRetry_FromOtherStatus_Forbidden`
- `TestReject_RefundsAndPersistsReason`、`TestReject_FromInvalidStatus_Forbidden`
- `TestUserCancel_OnlyInPendingReview`、`TestUserCancel_NotOwner_Forbidden`
- 资金守恒不变量测试：fuzz 一组随机操作序列，断言 available + frozen + outflow 守恒

### Handler / 集成测试

- `wallet_handler_test.go`：cancel 新路径
- 新建 `withdraw_review_handler_test.go`：approve / reject / retry，含权限断言（非 super_admin/operator 403）

### E2E（admin + front 联动）

1. 用户提现 → admin 拒绝 → 用户余额回退、收到拒绝邮件
2. 用户提现 → admin 批准（mock Cobo 成功）→ webhook → confirmed → 收到到账邮件
3. 用户提现 → admin 批准（mock Cobo 12007）→ submit_failed → admin 重试 → 成功 → confirmed
4. 用户提现 → 立即取消 → 余额回退

### 覆盖率目标

≥ 80%（与 `~/.claude/rules/testing.md` 一致）。

## 上线步骤

1. DB 迁移（含历史回填）
2. 部署后端：用户侧 Withdraw 行为变更 + 新 admin 接口（向后兼容，老 webhook 路径不受影响）
3. 部署 admin 前端：新提现审核页
4. 部署用户前端（可选）：调整 `withdrawSuccess` 文案为"提现申请已提交，等待审核"，加 cancel 按钮（仅在 `pending_review` 行显示）
5. 公告：通知用户审核制度

## 兼容性 / 风险

- 历史 `transactions(type=withdraw)` 行回填 `review_status=submitted`，对账口径不变
- `available + frozen + in_operation = total` 不变
- Webhook 处理路径完全保留
- 风险：审核 SLA 影响用户体验 → 通过 admin 工单页面 + 通知机制保证响应及时性（首版不内建告警，由运营 SOP 兜底）
- 风险：管理员误操作（误批/误拒）→ 已有 `audit_logs`，必要时人工补救（误拒 → 用户重发；误批已上链不可撤销，需人工对账）

## 后续可演进项（不在本期）

- 自动重试 worker（基于 `submit_failed + 时间阈值`）
- 自动审核阈值规则
- 双人复核 / 大额二次 2FA
- 风控规则引擎（地址黑名单、异常模式检测）
- Cobo 热钱包余额预警（独立健康检查）
