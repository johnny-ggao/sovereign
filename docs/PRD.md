# Sovereign Kimchi Premium Arbitrage - 产品需求文档 (PRD)

> 基于 Stitch 设计稿 `projects/17437626401050297890` 整理
> 最后更新: 2026-03-25

---

## 1. 产品概述

### 1.1 产品定位
Sovereign 是一个面向加密货币套利基金投资者的客户端门户平台，专注于韩国交易所（Upbit/Bithumb）与国际交易所（Binance/Bybit）之间的"泡菜溢价"(Kimchi Premium) 套利策略。

### 1.2 目标用户
- 高净值个人投资者
- 机构投资者
- 基金 LP（有限合伙人）

### 1.3 核心价值
- 实时监控泡菜溢价行情与套利机会
- 透明化展示基金收益与投资表现
- 安全便捷的资产充值/提现管理
- 合规化的 KYC 验证与结算报告

### 1.4 技术栈
| 层级 | 技术选型 |
|------|----------|
| **前端** | Next.js 14+ (App Router), TypeScript, shadcn/ui, Tailwind CSS |
| **后端** | FastAPI 或 Express |
| **数据库** | PostgreSQL |
| **KYC** | Sumsub / Onfido |
| **加密** | AES-256, HTTPS |

---

## 2. 设计系统 — "Sovereign Obsidian"

### 2.1 设计理念
采用高端、编辑式的深色界面风格，灵感来源于机构财富管理和隐匿科技界面。核心隐喻为 "Glass & Obsidian"（玻璃与黑曜石）—— 深色半透明层叠，营造物理堆叠的结构感。

### 2.2 色彩体系
| Token | 色值 | 用途 |
|-------|------|------|
| `surface_dim` | `#0c1322` | 基底层 (Level 0) |
| `surface_container` | `#191f2f` | 内容块 (Level 1) |
| `surface_container_highest` | `#2e3545` | 焦点卡片 (Level 2) |
| `primary` | `#adc6ff` | 主色调 / CTA |
| `primary_container` | `#4d8eff` | 渐变配色 |
| `secondary` | `#3fe397` | 盈利 / 成功状态 |
| `secondary_container` | `#00c77e` | 买入/盈利操作按钮 |
| `tertiary_container` | `#ff5451` | 亏损 / 警告状态 |
| `on_surface` | `#dce2f7` | 主文本 |
| `on_surface_variant` | `#c2c6d6` | 次要文本 |

### 2.3 关键设计规则
- **"No-Line" 规则**: 禁止使用 1px 实线边框分隔区块，通过背景色差异实现分层
- **"Ghost Border"**: 需要边缘时使用 `outline_variant` 15% 透明度
- **字体**: Inter（标题/正文/标签统一）
- **圆角**: `0.375rem` (ROUND_FOUR)
- **阴影**: 用大模糊 (40px+) 环境光模拟替代传统投影

---

## 3. 平台架构 — 屏幕清单

### 3.1 响应式适配
- **Desktop (Web)**: 1280px / 2560px 宽度
- **Mobile**: 390px / 780px 宽度
- 所有核心功能均提供 Web + Mobile 双端设计

### 3.2 功能模块与屏幕映射

#### 模块 A: 认证系统 (Authentication)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| A1 | Login - Sovereign Arbitrage | Desktop | `13ccb5b42d134719851f285816400fd5` |
| A2 | Login (Web) | Desktop | `5cee19caf5e747b0bbc957a7e1909408` |
| A3 | Login (V5 Final Clean) | Mobile | `70cc1f3c64c94de48ba71e8ead3c0763` |
| A4 | Login Error - Wrong Password (Mobile) | Mobile | `102923478b9b49a2a9e20841d03d4675` |
| A5 | Login Error - No Account (Mobile) | Mobile | `de47af15b5e248a9bad1c2ac7df4d9a6` |
| A6 | Register (Web) | Desktop | `0d783454e2484c45be45f47192fde977` |
| A7 | Register - Join the Fund | Desktop | `c412a06df8f54961803e05a7cd29d61c` |
| A8 | Register - Verify Phone OTP | Desktop | `15de8f8c21cb4ade847013ee6ae58758` |
| A9 | Register (V5 Refined) | Mobile | `df85def23c264eedbecabc787e62e0b9` |
| A10 | Register Success (V5) | Mobile | `cd14c4ac75c249d3bdd847b56dafa18b` |
| A11 | Register Post-Onboarding (V5) | Mobile | `e87e774c07714f80a3314b99a55132b9` |
| A12 | Forgot Password - Recovery | Desktop | `86de1ce1dcd94c899b618576f9759a9f` |
| A13 | Forgot Password - Email (V5) | Mobile | `1f25bb189e5d431db11f3dd721d85fb9` |
| A14 | Forgot Password - Verify (V5) | Mobile | `9c67906515ab45f68d6adec636c3b5f7` |

**功能要求:**
- 邮箱 + 密码登录，支持 2FA (二次验证)
- 注册流程: 邮箱 → 手机号 OTP 验证 → 设置密码 → 入金引导
- 忘记密码: 邮箱验证 → OTP → 重置密码
- 登录错误提示: 密码错误 / 账户不存在的差异化处理

---

#### 模块 B: 仪表盘 (Dashboard)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| B1 | Dashboard (Web) | Desktop | `9b3c0cfa97cc45f6954c705fa399fd52` |
| B2 | Dashboard (Web) - Invest CTA | Desktop | `2146c2099c9b4fdc849f318ab8f27d70` |
| B3 | Dashboard - Mobile (V4 Final Unified) | Mobile | `4b2a3635689a4bc79552f97d6aa546dd` |
| B4 | Dashboard - Kimchi Premium Arbitrage | Mobile | `50e82338831f4370a4b133a0c62c608b` |

**功能要求:**
- **总资产价值**: Display-lg 级别展示，支持 USDT/BTC 切换
- **累计收益**: 累计收益金额 + 年化收益率
- **高水位线可视化**: 资产净值曲线图
- **收益图表**: 支持 1W / 1M / 3M / 6M / 1Y / ALL 时间跨度
- **泡菜溢价实时指标**: 当前溢价率 + 趋势
- **快捷操作**: 投资 CTA、充值/提现入口

---

#### 模块 C: 泡菜溢价行情 (Premium Ticker)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| C1 | Premium Ticker (Web) | Desktop | `73700cf79bdc4739a78f491ef8a78fc8` |
| C2 | Premium Ticker - Live Data | Mobile | `1bf3ba9b2bb246958369b8c916fd3848` |

**功能要求:**
- **实时溢价数据**: Upbit/Bithumb vs Binance/Bybit 价差
- **交易对**: BTC, ETH 及主流币种
- **溢价率展示**: 百分比 + 绝对值
- **历史溢价走势图**: 可选时间范围
- **外部数据源**: 实时行情 feed 接入

---

#### 模块 D: 投资管理 (Investment)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| D1 | Investment - Input Amount (Refined) | Mobile | `d01a37244bb94840b4a8836d43d72177` |
| D2 | Investment Success (V4 Nav Fixed) | Mobile | `14eac60cb02b4bed87912b6d03605a9c` |
| D3 | Investment - Management Hub (V4 Unified) | Mobile | `88e881ec872642ffbbd8ac2351c636c1` |
| D4 | Investment - Trade Log (V4) | Mobile | `abe9912adade4ce3bd8894a339349321` |
| D5 | Investment - Stop & Redeem (V4.3 Final Unified) | Mobile | `a7df89573bc44676bf56cbdb19d9e68d` |
| D6 | Investment - Settlement Reports (V4) | Mobile | `f6ed538164514eb8b92a9216890d6147` |

**功能要求:**
- **投入金额**: 输入投资金额（USDT），显示最低/最高限额
- **投资确认**: 确认后锁定资金进入套利运营
- **管理中心**: 查看当前投资状态、收益率、运营周期
- **止赎**: 申请停止套利并赎回本金 + 收益
- **交易日志**: 每笔套利交易的详细记录
- **结算报告**: 月度结算，含 50% 绩效费明细

---

#### 模块 E: 交易记录 (Trade Log)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| E1 | Trade Log (Web) | Desktop | `c438cfa9745d4bd08be0e446f42e7de7` |
| E2 | Trade Log - Historical Data | Mobile | `c7562d8f21994b658a973c3195713161` |

**功能要求:**
- **套利交易列表**: 时间、交易对、买入/卖出交易所、溢价率、盈亏
- **筛选与排序**: 按日期、交易对、盈亏状态筛选
- **导出**: 支持 CSV 下载

---

#### 模块 F: 钱包系统 (Wallet)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| F1 | Wallet (Web) | Desktop | `f790d2241ea247a1aa0ccedbf35ccebd` |
| F2 | Wallet Home (V4 Final Fixed Unified) | Mobile | `492b590980fb4346bc73dffe1b23860c` |
| F3 | Wallet - Portfolio Summary | Mobile | `9d996ef5bd6b4031ba90e78f00603862` |
| F4 | Wallet - Deposit (V4) | Mobile | `417d66a94c3341e68ad412c188b5676c` |
| F5 | Wallet - Deposit History (V4 Final Corrected) | Mobile | `1fe58e0e38a948b4b41f63077345e89a` |
| F6 | Wallet - Withdraw (V4.1 Fixed) | Mobile | `28251624956c4fe08b7b0db2a55ae723` |
| F7 | Withdrawal History (V4 Final Corrected) | Mobile | `68cfc7ec6c7c460fa72e27483168ce85` |

**功能要求:**
- **资产分类**: Available（可用）/ In-operation（运营中）/ Frozen（冻结）
- **支持币种**: USDT, BTC, ETH
- **充值**: 链上充值地址生成（支持多网络: ERC-20, TRC-20, BEP-20）
- **提现**: 地址白名单管理，新地址 24 小时冷却期
- **提现安全**: 2FA 验证，提现限额管理
- **交易明细**: 充值/提现历史记录，含状态追踪（Pending/Confirmed/Failed）

---

#### 模块 G: 交易详情 (Transaction Details)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| G1 | Transaction Details (V4 Final Corrected) | Mobile | `ac82315c48db4e6a9643c5a71caddac9` |
| G2 | Transactions (V4 Final Corrected) | Mobile | `e3a45802fc78401fbf3e21eb23582ec7` |

**功能要求:**
- 单笔交易详情: 金额、币种、网络、TX Hash、状态、时间戳
- 交易列表: 全部交易的统一视图

---

#### 模块 H: 结算报告 (Settlement Reports)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| H1 | Reports (Web) | Desktop | `93bbbfc5c091451bb91912f3639c9eed` |
| H2 | Settlement - Monthly Reports | Mobile | `7e4a45583d3242d38e1ff4116be24a1b` |

**功能要求:**
- **月度报告**: 本月总收益、套利次数、平均溢价率
- **绩效费计算**: 50% Performance Fee 明细拆分
- **Net 收益**: 扣费后投资者实际收益
- **历史报告**: 按月份归档，支持下载 PDF

---

#### 模块 I: 设置中心 (Settings)

| # | 屏幕名称 | 平台 | Screen ID |
|---|----------|------|-----------|
| I1 | Settings (Web) | Desktop | `a615b4674e084418a18e24c2fab541dc` |
| I2 | Settings - Account Hub | Mobile | `0e86f22013574625b1c5f25952e94067` |
| I3 | Settings - KYC Verification | Mobile | `c0e99fb31b4846e1bcca9c138ec66257` |
| I4 | Settings - Security & Privacy | Mobile | `1c40863faa704d329749bb57acbb1bdd` |
| I5 | Settings - Notifications | Mobile | `3c0cddf4f397405ab8042b6364d3290c` |
| I6 | Settings - Language (Updated) | Mobile | `eb0e76913e544466b20b3e93fc17722d` |

**功能要求:**
- **账户中心**: 个人资料编辑、头像、联系方式
- **KYC 验证**: 集成 Sumsub/Onfido，身份证/护照上传，人脸识别
- **安全与隐私**: 2FA 设置（TOTP/SMS）、密码修改、登录设备管理
- **通知设置**: 邮件/推送通知偏好（交易完成、溢价警报、结算通知）
- **语言设置**: 多语言支持（韩文/英文/中文）

---

## 4. 安全架构

### 4.1 传输层
- 全站 HTTPS (TLS 1.3)
- API 请求签名验证

### 4.2 数据层
- AES-256 加密存储敏感数据
- 数据库字段级加密（钱包地址、KYC 文件）

### 4.3 认证层
- JWT + Refresh Token 机制
- 2FA 强制要求（提现、敏感操作）
- Session 超时自动登出

### 4.4 钱包安全
- Hot/Cold Wallet 分离
- 新提现地址 24 小时冷却期
- 提现白名单机制
- 每日/单笔提现限额

---

## 5. 导航结构

### 5.1 Web 端侧边栏
```
Dashboard (仪表盘)
├── Overview (概览)
└── Invest CTA (投资入口)
Premium Ticker (溢价行情)
Trade Log (交易记录)
Wallet (钱包)
├── Deposit (充值)
└── Withdraw (提现)
Reports (报告)
Settings (设置)
```

### 5.2 Mobile 端底部导航
```
Home (首页/仪表盘)
Invest (投资)
Wallet (钱包)
Market (行情)
Profile (我的/设置)
```

---

## 6. API 端点规划

### 6.1 认证
| Method | Endpoint | 描述 |
|--------|----------|------|
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/verify-otp` | OTP 验证 |
| POST | `/api/auth/forgot-password` | 忘记密码 |
| POST | `/api/auth/reset-password` | 重置密码 |
| POST | `/api/auth/refresh` | 刷新 Token |
| POST | `/api/auth/2fa/setup` | 设置 2FA |
| POST | `/api/auth/2fa/verify` | 验证 2FA |

### 6.2 仪表盘
| Method | Endpoint | 描述 |
|--------|----------|------|
| GET | `/api/dashboard/summary` | 总览数据 |
| GET | `/api/dashboard/performance` | 收益曲线 |
| GET | `/api/dashboard/premium` | 当前溢价 |

### 6.3 投资
| Method | Endpoint | 描述 |
|--------|----------|------|
| POST | `/api/investment/create` | 创建投资 |
| GET | `/api/investment/status` | 投资状态 |
| POST | `/api/investment/redeem` | 赎回申请 |
| GET | `/api/investment/trade-log` | 交易日志 |
| GET | `/api/investment/settlements` | 结算报告 |

### 6.4 钱包
| Method | Endpoint | 描述 |
|--------|----------|------|
| GET | `/api/wallet/balance` | 余额查询 |
| GET | `/api/wallet/portfolio` | 资产组合 |
| POST | `/api/wallet/deposit/address` | 生成充值地址 |
| GET | `/api/wallet/deposit/history` | 充值记录 |
| POST | `/api/wallet/withdraw` | 发起提现 |
| GET | `/api/wallet/withdraw/history` | 提现记录 |
| GET | `/api/wallet/transactions` | 交易明细 |
| GET | `/api/wallet/transactions/:id` | 单笔详情 |

### 6.5 行情
| Method | Endpoint | 描述 |
|--------|----------|------|
| GET | `/api/market/premium` | 实时溢价数据 |
| GET | `/api/market/premium/history` | 历史溢价 |
| WS | `/ws/market/premium` | 溢价实时推送 |

### 6.6 设置
| Method | Endpoint | 描述 |
|--------|----------|------|
| GET | `/api/settings/profile` | 获取资料 |
| PUT | `/api/settings/profile` | 更新资料 |
| POST | `/api/settings/kyc/submit` | 提交 KYC |
| GET | `/api/settings/kyc/status` | KYC 状态 |
| PUT | `/api/settings/security` | 安全设置 |
| PUT | `/api/settings/notifications` | 通知偏好 |
| PUT | `/api/settings/language` | 语言设置 |

---

## 7. 数据模型概要

### 7.1 核心实体
```
User
├── id, email, phone, password_hash
├── kyc_status (pending/verified/rejected)
├── 2fa_enabled, 2fa_secret
└── language, created_at, updated_at

Investment
├── id, user_id, amount, currency
├── status (active/stopped/redeemed)
├── start_date, end_date
└── total_return, performance_fee

Wallet
├── id, user_id, currency
├── available, in_operation, frozen
└── total_balance

Transaction
├── id, user_id, type (deposit/withdraw)
├── amount, currency, network
├── tx_hash, status, address
└── created_at, confirmed_at

TradeLog
├── id, investment_id
├── pair, buy_exchange, sell_exchange
├── premium_rate, pnl
└── executed_at

Settlement
├── id, investment_id, period
├── gross_return, performance_fee
├── net_return, report_url
└── settled_at

WhitelistAddress
├── id, user_id, address, network
├── label, cooldown_expires_at
└── is_active
```

---

## 8. 非功能需求

| 维度 | 要求 |
|------|------|
| **性能** | 页面加载 < 2s，API 响应 < 500ms |
| **可用性** | 99.9% SLA |
| **国际化** | 韩文 / 英文 / 中文 |
| **响应式** | Desktop (1280px+) + Mobile (390px) |
| **无障碍** | WCAG 2.1 AA |
| **SEO** | SSR/SSG 关键页面 |

---

## 9. 开发优先级

### Phase 1 — MVP (核心功能)
1. 认证系统 (登录/注册/忘记密码)
2. 仪表盘 (资产概览/收益图表)
3. 钱包 (充值/提现/余额)
4. 基础设置 (个人资料/安全)

### Phase 2 — 增强功能
5. 泡菜溢价行情 (实时数据/WebSocket)
6. 投资管理 (投入/赎回/管理中心)
7. 交易记录 (交易日志/历史)
8. KYC 集成

### Phase 3 — 高级功能
9. 结算报告 (月度报告/PDF 导出)
10. 通知系统 (邮件/推送)
11. 多语言支持
12. 管理后台
