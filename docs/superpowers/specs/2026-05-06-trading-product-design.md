# 交易投资产品（Trading Product）设计

- 日期：2026-05-06
- 状态：Draft → 待 Review
- 作者：Sovereign Engineering

## 背景

现有产品是套利投资（KIMP 套利），所有用户共用一个产品池，结算逻辑：每天 T+1 把 fund 当日交易盈利的 50% 按用户份额比例分给当时活跃的投资。

新增第二个产品「交易投资」（Trading Product，使用别的策略），内测期不在用户端 UI 区分 —— 同一个「投资」按钮，由后台标记决定本笔投资归属哪个产品池。

**目标：**
- 用户被标记为 trading 后，新投资走 trading 池
- 老的 arbitrage 投资保留可见、保留收益分配，直到赎回
- trading 数据（fund 交易记录、用户份额、结算）和 arbitrage 完全隔离，互不影响
- 后台可单人/批量给用户打标 + 导入 trading 交易数据
- 算法和现有套利完全相同，只是数据源不同

## 范围

**包含：**
- 用户标记机制（`users.investment_type` + 审计表）
- 数据模型扩展（trades 物理分表 + investments/user_trades/settlements 加 product_type 字段）
- 投资创建路由按用户标记决定产品归属
- 结算 Job 改造为双循环
- 用户视图按 product_type 分组展示
- 钱包 earnings 拆两列存储、单一池子展示
- 后台用户标记 UI（单/批）
- 后台 trading 数据导入与列表

**不包含（明确 YAGNI）：**
- 用户端产品选择 UI（内测期不暴露）
- 不同产品差异化 fee rate（先都 50%）
- 不同产品差异化模板字段（先共用同一 schema）
- 第三个产品扩展（架构允许，但本期不做）
- 用户主动申请切换产品
- 跨产品余额转移
- 自动重新分配（用户改 type 不影响在跑投资，新投资才走新 type）

## 设计决策摘要

| # | 决策 | 选择 |
|---|---|---|
| 1 | 用户与产品的关系基数 | 一刀切（一个用户当下只能投一个产品；老的另一个产品的投资保留可见） |
| 2 | 数据存储结构 | 混合：`trades` 物理分表（新建 `trading_trades`），`investments`/`user_trades`/`settlements` 共用表加 `product_type` |
| 3 | 用户标记 | `users.investment_type` 字段 + 独立审计表 `user_product_change_logs` |
| 4 | 结算 Job | 单 Job 双循环（按 product_type 跑两遍）|
| 5 | 用户视图过滤 | 按 `product_type` 分 tab 展示，默认显示当前产品 |
| 6 | 钱包收益池 | 后端两列分开（`earnings_arbitrage` + `earnings_trading`），前端聚合显示 |
| 7 | 后台导入入口 | 完全独立路由 `/admin/trading-trades/import` + 独立菜单 |
| 8 | 后台打标 UI | 用户详情页（单人）+ 用户列表页（批量）|
| 9 | 绩效费率 | 两个产品都 50% |
| 10 | 旧 `wallets.earnings` 字段 | 删除并迁移到 `earnings_arbitrage` |

## 数据模型

### 新建表

```sql
-- 用户产品类型变更审计（独立于 admin_audit_logs，便于按用户/时间筛选）
CREATE TABLE user_product_change_logs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL,
  from_type   varchar(20) NOT NULL,
  to_type     varchar(20) NOT NULL,
  admin_id    uuid NOT NULL,
  admin_email varchar(255) NOT NULL,
  reason      varchar(500) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_upcl_user ON user_product_change_logs(user_id);
CREATE INDEX idx_upcl_created ON user_product_change_logs(created_at);
```

```sql
-- 交易策略 fund 级交易（与套利 trades 表完全隔离）
CREATE TABLE trading_trades (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair          varchar(20) NOT NULL,
  buy_exchange  varchar(20),
  sell_exchange varchar(20),
  buy_price     decimal(28,8),
  sell_price    decimal(28,8),
  amount        decimal(28,18) NOT NULL,
  premium_pct   decimal(8,4) DEFAULT 0,
  pnl           decimal(28,18) NOT NULL,
  fee           decimal(28,18) DEFAULT 0,
  source        varchar(20) NOT NULL DEFAULT 'import',
  executed_at   timestamptz NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_trading_trades_executed_at ON trading_trades(executed_at);
```

> 字段刻意和 `trades` 完全一致 —— 算法相同，复用导入/结算代码逻辑。`buy_exchange` / `sell_exchange` 等字段对 trading 策略可能不严格适用，但保留以保证 schema 一致；后期 trading 真有专属字段再 ALTER。

### 扩展现有表

```sql
-- 用户标记
ALTER TABLE users
  ADD COLUMN investment_type varchar(20) NOT NULL DEFAULT 'arbitrage';

-- 投资记录归属
ALTER TABLE investments
  ADD COLUMN product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX idx_investments_product ON investments(product_type, status);

-- 用户份额归属
ALTER TABLE user_trades
  ADD COLUMN product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX idx_user_trades_product ON user_trades(product_type);

-- 结算记录归属
ALTER TABLE settlements
  ADD COLUMN product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX idx_settlements_product ON settlements(product_type, period);

-- 钱包按产品分别计收益（前端聚合显示）
ALTER TABLE wallets
  ADD COLUMN earnings_arbitrage decimal(28,18) DEFAULT 0,
  ADD COLUMN earnings_trading   decimal(28,18) DEFAULT 0;

-- 回填存量数据
UPDATE wallets SET earnings_arbitrage = earnings;

-- 删除旧字段（破坏性变更）
ALTER TABLE wallets DROP COLUMN earnings;
```

> 所有现存数据回填为 `'arbitrage'`，老用户/老投资语义不变。

### product_type 枚举

```go
const (
    ProductTypeArbitrage = "arbitrage"
    ProductTypeTrading   = "trading"
)
```

## 投资创建路由

`POST /api/v1/investments` 行为变更：

1. 校验 amount、available 余额（保持原逻辑）
2. **读 `users.investment_type`** → 决定本笔投资的 `product_type`
3. 创建 Investment 记录时 `product_type = users.investment_type`
4. 用户感知不到变化 —— 同一个按钮，但走哪个产品池由后台标记决定

```go
// investment_service.Create
user, _ := userRepo.FindByID(ctx, userID)
inv := &model.Investment{
    UserID:      userID,
    Amount:      amount,
    Currency:    "USDT",
    Status:      InvestStatusActive,
    ProductType: user.InvestmentType,   // 关键
    StartDate:   time.Now(),
}
```

## 结算 Job

`SettlementJob.RunForDate(ctx, date)` 改造为按 product_type 跑两遍：

```go
func (j *SettlementJob) RunForDate(ctx context.Context, date time.Time) error {
    for _, productType := range []string{
        model.ProductTypeArbitrage,
        model.ProductTypeTrading,
    } {
        if err := j.settleProduct(ctx, date, productType); err != nil {
            j.logger.Error("settle product failed",
                slog.String("product", productType),
                slog.String("error", err.Error()),
            )
            // 一个产品失败不影响另一个
        }
    }
    return nil
}

func (j *SettlementJob) settleProduct(ctx context.Context, date time.Time, productType string) error {
    dayStart := truncateToDay(date)
    dayEnd := dayStart.Add(24 * time.Hour)
    period := date.Format("2006-01-02")

    // 1. 该产品的 active 投资（T+1 规则保持不变）
    activeInvs, err := j.invRepo.FindAllActiveBeforeDateByProduct(ctx, dayStart, productType)
    if err != nil { return err }
    if len(activeInvs) == 0 {
        j.logger.Info("no active investments", slog.String("product", productType))
        return nil
    }

    // 2. 该产品的当日 fund pnl（数据源根据 productType 切换）
    var summary TradeSummary
    if productType == model.ProductTypeArbitrage {
        summary, err = j.tradeRepo.SummarizeByPeriod(ctx, dayStart, dayEnd)
    } else {
        summary, err = j.tradingTradeRepo.SummarizeByPeriod(ctx, dayStart, dayEnd)
    }
    if err != nil { return err }

    if summary.TotalPnL <= 0 {
        j.logger.Info("no profit", slog.String("product", productType))
        return nil
    }

    // 3. 50% 分给用户，按份额比例（同现有算法）
    // 写入 user_trades 和 settlements 时 product_type = productType
    // 钱包累加到 earnings_<productType> 列
}
```

错误隔离：一个产品 settle 失败不影响另一个。

## 用户侧 API 改造

| 路径 | 变更 |
|---|---|
| `POST /api/v1/investments` | 读 `users.investment_type` 设 product_type |
| `GET /api/v1/investments` | query `?product_type=arbitrage|trading|all`，默认按用户当前 type |
| `GET /api/v1/settlements` | 同上 |
| `GET /api/v1/wallets` | 响应中聚合 `earnings = earnings_arbitrage + earnings_trading`，前端不感知拆分 |
| `POST /api/v1/wallets/claim-earnings` | 一次清空两列到 `available` |

`investment_type` 字段保留在 `users` 响应中（小字段），方便前端决定是否显示「其他产品」tab。

## 后台 API 新增

### 用户标记

| 方法 & 路径 | 说明 |
|---|---|
| `POST /admin/users/:id/investment-type` | body `{type, reason}`，必填 reason ≥ 5 字符。单人切换。写入 `user_product_change_logs` + `admin_audit_logs` |
| `POST /admin/users/bulk-investment-type` | body `{user_ids[], type, reason}`，循环单条逻辑。共用同一 reason；单条失败不影响其他。返回 `{succeeded[], failed[]}` |
| `GET /admin/users/:id/product-changes` | 该用户产品类型变更历史 |

权限：`super_admin` + `operator`。

### 交易策略数据（新建模块 `trading_tradelog`）

| 方法 & 路径 | 说明 |
|---|---|
| `GET /admin/trading-trades` | 列表 + 分页 + 日期/pair 筛选（同 `/admin/trades`） |
| `POST /admin/trading-trades/import` | Excel 导入。复用现有 trade import parser 但写入 `trading_trades` 表。`executed_at` 同样由系统自动生成（之前已统一规则） |
| `GET /admin/trading-trades/template` | 下载模板（同套利模板，文件名标 `_trading`） |
| `DELETE /admin/trading-trades/:id` | 仅 super_admin |
| `GET /admin/trading-trades/stats` | 总笔数 / 总 PnL / 当日/7d/30d |

权限：`super_admin` + `operator`（删除仅 super_admin）。

## Service 层改造

- 新建 `server/internal/modules/trading_tradelog/`（镜像 `tradelog/` 的目录结构）：
  - `model/trade.go` —— `TradingTrade` struct，TableName=`trading_trades`
  - `repository/trading_trade_repo.go` —— `Create / FindByID / List / SummarizeByPeriod / Delete`
  - 不需要独立 service —— 复用 admin 模块的 service 层（因为这只是 admin 入口数据）
- `investment_service.Create` 读 user.InvestmentType 设 product_type
- `investment_service.List` 接受 product_type 过滤参数
- `settlement_service.List` 同上
- `wallet_service.GetWallets` 聚合 earnings；`ClaimEarnings` 清空两列
- `user_service` 加 `UpdateInvestmentType(userID, newType, reason, adminID, adminEmail)` —— 写 users + 写 audit
- `user_service` 加 `BulkUpdateInvestmentType(userIDs[], ...)`

## 前端改造

### 用户端

- **投资管理页**：当用户既有 arbitrage 又有 trading 投资时，顶部显示 tab `当前产品 / 历史产品`；纯单产品用户不显示 tab
- **钱包页 earnings 卡片**：保持单一显示（聚合后的 total）
- **结算/报表页**：tab 同投资管理
- **dashboard 总收益/总投入卡片**：默认显示当前产品；tab 同上
- 用户感知到的字段命名继续是「投资」「收益」，不出现 trading/arbitrage 字样

### 后台

- **用户详情页**：加「投资产品类型」卡片
  - 当前类型徽章（套利 / 交易）
  - 「切换」按钮 → 弹窗（目标类型 + 必填原因 ≥ 5 字符）
  - 弹窗下方折叠展示该用户的「变更历史」（来自 `user_product_change_logs`）
- **用户列表页**：
  - 加「投资类型」列
  - 列头筛选（套利 / 交易）
  - 多选行后顶部出现「批量打标」按钮 → 弹窗（目标 + 必填原因，提示影响 N 个用户）
- **新菜单**：「交易策略」一级菜单（与「套利交易」并列），子菜单：
  - 交易记录（列表 + 统计）
  - 导入交易记录
- 套利现有菜单文案可考虑改名「套利交易」，将来更对称（YAGNI 暂不改）

## 通知

- 用户产品类型变更**不通知用户**（内测期，避免引起疑问）
- 用户标记后立即生效，无邮件通知

## 审计

- 所有 admin 标记动作（单人 / 批量）：写 `admin_audit_logs`（target_type=`user`、action=`change_investment_type`，detail 含 from→to + reason + 批量时含 user_count）
- 同时写 `user_product_change_logs`（专门为该用户筛选历史）
- trading_trades 的导入/删除：写 `admin_audit_logs`（与现有 trade 操作同模式）

## 测试计划

### 单元测试

- 用户标记：from→to 时写入审计记录、users 表更新；reason 必填校验
- 批量标记：N 个用户成功 + 1 个失败时返回 succeeded/failed 列表
- Investment.Create：mock user.investment_type 不同时 product_type 正确
- Investment.List：按 product_type 过滤
- Settlement Job：mock 两个产品的 trades + invs，验证两个池子独立分配，互不污染
- Settlement Job：某产品当天无 trade → skip，不影响另一个
- Settlement Job：某产品无 active investment → skip
- Settlement Job：某产品 settle 失败 → 不影响另一个产品
- Wallet GetWallets：earnings 聚合正确
- ClaimEarnings：两列清零、available 正确累加
- TradingTradeRepo：Create / SummarizeByPeriod / Delete

### 集成测试

- 标记用户 → 该用户创建投资 → product_type=trading
- 导入 trading_trades → 触发结算 → 仅 trading 产品的用户钱包 earnings_trading 增加
- arbitrage 用户在 trading 数据导入后**没有**收到分配（验证隔离）
- 旧 arbitrage 数据迁移后所有读路径（dashboard / wallet / investments / settlements）正常返回

### 覆盖率目标

≥ 80%（与 `~/.claude/rules/testing.md` 一致）。

## 上线步骤

1. DB 迁移：建新表 + 给现有表加列 + 回填 + 删 `wallets.earnings`
2. 部署后端：含新 trading_tradelog 模块、settlement job 双循环、users.investment_type 读写
3. 部署 admin 前端：标记 UI + trading-trades 导入页
4. 部署 user 前端：investment/settlement tab、wallet earnings 聚合
5. 运营给种子用户打 trading 标
6. 第一笔 trading_trade 导入；T+1 跑结算验证
7. 监控：observe `wallets.earnings_trading` 是否增加、`settlements.product_type='trading'` 是否有新记录

## 兼容性 / 风险

- 所有现存数据回填到 `'arbitrage'`，旧用户/旧投资行为不变
- `wallets.earnings` 字段删除是破坏性变更：
  - 任何外部脚本/对账报表读这个字段都需要改用聚合
  - 回滚需要恢复字段并把两列求和回写
  - 部署期间短暂的 schema 不一致由 Docker compose 部署顺序保证：迁移 → 后端 → 前端
- Settlement Job 改造为双循环后单次运行时间略增；当 trading 数据为空时几乎无开销
- 一个用户从 trading 改回 arbitrage 后：
  - 老 trading 投资按 Q5=B 规则继续显示在「历史产品」tab
  - 用户能看到收益、能赎回，但不能新建 trading
- 误标用户场景：管理员可立即改回，但已经创建的投资 product_type 不会回滚（这是设计选择，避免数据修复混乱）；如需修复需要 DBA 手工 update

## 后续可演进项（不在本期）

- 不同产品差异化 fee rate（`product_type → fee_rate` 配置表）
- trading 策略专属字段（如 strategy_id、symbol、leverage 等）
- 用户主动产品切换申请流程
- 跨产品余额转移
- 第三个产品：架构允许，按 trading 模式照搬一份即可
- 用户端显示「投资了哪个产品」（如果将来产品定位需要透明）
- 自动化 trading_trades 数据接入（替代手工导入）
