# AWS SES 邮件通知模块设计

## 概述

为 Sovereign Fund 平台接入 AWS SES，实现事务性邮件通知。通过新建 `notification` 模块，监听 EventBus 事件，根据用户通知偏好异步发送多语言邮件。

## 需求

- 事务性邮件（非营销）
- 三语支持：中文、英文、韩文
- 本地 Go 模板（`html/template`）
- 事件驱动异步发送
- AWS SES 已配置完成（域名验证、脱离沙箱）

---

## 1. 模块结构

新建 `internal/modules/notification` 模块，遵循项目现有的模块化架构：

```
internal/modules/notification/
├── module.go                        # 模块初始化，注册事件订阅
├── provider/
│   ├── provider.go                  # EmailProvider 接口定义
│   ├── ses.go                       # AWS SES 实现
│   └── mock.go                      # Mock 实现（开发/测试）
├── service/
│   └── notification_service.go      # 核心逻辑：查偏好 → 选模板 → 发邮件
└── template/
    ├── renderer.go                  # 模板渲染引擎
    └── emails/                      # HTML 模板文件
        ├── deposit_confirmed/
        │   ├── zh.html
        │   ├── en.html
        │   └── ko.html
        ├── withdraw_completed/
        │   ├── zh.html
        │   ├── en.html
        │   └── ko.html
        ├── withdraw_failed/
        │   ├── zh.html
        │   ├── en.html
        │   └── ko.html
        ├── settlement_created/
        │   ├── zh.html
        │   ├── en.html
        │   └── ko.html
        └── password_reset/
            ├── zh.html
            ├── en.html
            └── ko.html
```

## 2. Provider 抽象

与 Cobo 的 `WalletProvider` 模式一致，定义接口 + 双实现。

### 接口定义

```go
type EmailProvider interface {
    Send(ctx context.Context, input SendInput) error
}

type SendInput struct {
    To      string
    Subject string
    HTML    string
}
```

### SES 实现

- 使用 AWS SDK v2（`github.com/aws/aws-sdk-go-v2/service/sesv2`）
- 调用 `ses.SendEmail` API
- AWS 凭证走标准 SDK 凭证链（环境变量 `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` 或 IAM Role）

### Mock 实现

- `config.notification.use_mock: true` 时启用
- 日志打印邮件内容，不实际发送
- 记录所有调用，方便测试断言

```go
type MockProvider struct {
    Sent []SendInput
}

func (m *MockProvider) Send(ctx context.Context, input SendInput) error {
    m.Sent = append(m.Sent, input)
    return nil
}
```

## 3. 事件订阅与通知偏好

### 监听的事件

| 事件                  | 邮件类型       | 偏好字段          | 可跳过 |
|-----------------------|---------------|-------------------|--------|
| `DepositConfirmed`    | 充值到账通知   | `EmailDeposit`    | 是     |
| `WithdrawCompleted`   | 提现完成通知   | `EmailWithdraw`   | 是     |
| `WithdrawFailed`      | 提现失败通知   | `EmailWithdraw`   | 是     |
| `SettlementCreated`   | 每日结算报告   | `EmailSettlement` | 是     |
| `UserPasswordReset`   | 密码重置 OTP   | —                 | 否     |

安全类邮件（密码重置）始终发送，无视用户偏好设置。

### 处理流程

```
EventBus 事件触发
  → NotificationService.Handle(eventType, payload)
    → 从 payload 取 userID
    → 查 User（获取 email、language）
    → 查 NotificationPref（检查对应开关）
    → 如果开关关闭，跳过
    → 如果是安全类邮件（密码重置），无视开关，始终发送
    → 用 Renderer 按 language 渲染对应模板
    → 调用 EmailProvider.Send()
    → 发送失败只记日志，不影响主流程
```

### 事件订阅注册（在 app.go 中）

```go
bus.Subscribe(events.DepositConfirmed, notifMod.Service.HandleDepositConfirmed)
bus.Subscribe(events.WithdrawCompleted, notifMod.Service.HandleWithdrawCompleted)
bus.Subscribe(events.WithdrawFailed, notifMod.Service.HandleWithdrawFailed)
bus.Subscribe(events.SettlementCreated, notifMod.Service.HandleSettlementCreated)
bus.Subscribe(events.UserPasswordReset, notifMod.Service.HandlePasswordReset)
```

### 跨模块依赖

NotificationService 需要注入：
- `settings` 模块的 `SettingsRepository`（查通知偏好）
- `auth` 模块的 `UserRepository`（查用户 email 和 language）

通过 `app.go` 初始化时注入 repository 实例，不直接依赖其他模块的 service 层。

## 4. 模板系统

### 渲染引擎

```go
type Renderer struct {
    templates map[string]map[string]*template.Template  // [eventType][lang] → template
}

func (r *Renderer) Render(eventType, lang string, data any) (subject string, html string, err error)
```

- 启动时预加载所有模板文件，避免运行时 IO
- 语言回退策略：请求的语言不存在时 fallback 到 `en`

### 模板文件约定

每个模板文件包含 `subject` 和 `body` 两个 block：

```html
{{define "subject"}}Your deposit of {{.Amount}} {{.Currency}} has been confirmed{{end}}
{{define "body"}}
<html>
  <body>
    <h2>Deposit Confirmed</h2>
    <p>{{.Amount}} {{.Currency}} has arrived in your account.</p>
    <p>Network: {{.Network}}</p>
    <p>TxHash: {{.TxHash}}</p>
  </body>
</html>
{{end}}
```

### 模板数据

| 模板                   | 数据字段                                  |
|------------------------|------------------------------------------|
| `deposit_confirmed`    | Amount, Currency, Network, TxHash        |
| `withdraw_completed`   | Amount, Currency, Network, TxHash, ToAddress |
| `withdraw_failed`      | Amount, Currency, Reason                 |
| `settlement_created`   | Date, TotalPnL, UserShare, FeeRate       |
| `password_reset`       | OTPCode, ExpiresIn                       |

### HTML 样式

- 内联 CSS（邮件客户端兼容性）
- 单列布局：品牌色 header + 内容区 + footer
- Footer 包含「管理通知偏好」链接
- 不依赖外部 CSS/JS 资源

## 5. 配置

### config.yaml

```yaml
notification:
  use_mock: true
  from_address: "noreply@example.com"
  from_name: "Sovereign Fund"
  aws_region: "ap-northeast-2"
```

### 环境变量

```bash
SOVEREIGN_NOTIFICATION__USE_MOCK=false
SOVEREIGN_NOTIFICATION__FROM_ADDRESS=noreply@example.com
SOVEREIGN_NOTIFICATION__FROM_NAME=Sovereign Fund
SOVEREIGN_NOTIFICATION__AWS_REGION=ap-northeast-2
# AWS 凭证走标准 SDK 链
# AWS_ACCESS_KEY_ID=xxx
# AWS_SECRET_ACCESS_KEY=xxx
```

## 6. 错误处理

邮件发送是非关键路径，失败不影响主业务：

| 场景               | 处理方式                                |
|--------------------|----------------------------------------|
| SES API 调用失败   | 记录错误日志，不重试（避免重复发送）       |
| 用户邮箱为空       | 跳过，记 warn 日志                      |
| 模板渲染失败       | 记录错误日志（代码 bug，应在开发阶段发现） |
| 通知偏好查询失败   | 记录错误日志，跳过发送                    |

## 7. 测试策略

| 层级     | 测试内容                             | 方式                              |
|----------|--------------------------------------|-----------------------------------|
| Provider | SES 实现的参数构建                   | 单元测试，mock AWS SDK            |
| Service  | 偏好检查、模板选择、语言回退          | 单元测试，用 MockProvider          |
| Renderer | 模板渲染正确性                       | 单元测试，验证输出 HTML            |
| 集成     | EventBus → Service → Provider 全链路 | 集成测试，用 MockProvider 验证调用 |

目标覆盖率：80%+

## 8. 依赖

新增 Go 依赖：
- `github.com/aws/aws-sdk-go-v2` — AWS SDK 核心
- `github.com/aws/aws-sdk-go-v2/config` — SDK 配置加载
- `github.com/aws/aws-sdk-go-v2/service/sesv2` — SES v2 客户端
