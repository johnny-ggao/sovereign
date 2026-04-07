# 后台管理系统 — 阶段 1 设计

## 概述

为 Sovereign Fund 平台新建后台管理系统，支持管理员登录、权限管理、用户管理和首页数据统计。后端复用现有 Go 服务，前端独立 Next.js 项目，使用 `@arco-design/mobile-react` 组件库，纯移动端 H5 设计。

## 范围

阶段 1 聚焦：
- 后台基础框架（认证、权限、布局）
- 管理员管理（CRUD）
- 用户管理（全功能：查/改/禁用/重置密码/调整余额）
- Dashboard 首页统计

不包含（后续阶段）：
- 充提币管理
- 投资管理
- 收益/结算管理

---

## 1. 整体架构

### 项目结构

```
sovereign/
├── server/          # 现有 Go 后端（新增 admin 模块）
├── front/           # 现有用户端 Next.js
└── admin/           # 新建 管理后台 Next.js 项目
```

- **后端**：在 `server/internal/modules/admin` 新增管理模块，路由挂载在 `/api/v1/admin/`，使用独立的 Admin JWT 认证中间件（与用户端完全隔离）
- **前端**：新建 `admin/` 目录，独立 Next.js 项目，使用 `@arco-design/mobile-react` 移动端组件库，纯 H5 移动端设计，独立部署
- 用户端和管理端共享同一个 Go 后端服务和数据库，但认证体系完全隔离

### 认证流程

```
Admin 登录 → POST /api/v1/admin/auth/login (email + password)
  → 验证 admin_users 表
  → 返回 Admin JWT（独立于用户端 JWT，claims 含 role）
  → 前端存储 token，后续请求 Bearer token
  → Admin 中间件校验 token + 权限
```

---

## 2. 数据库

### 新增表：admin_users

```sql
CREATE TABLE admin_users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255) NOT NULL,
  role          VARCHAR(20) NOT NULL DEFAULT 'viewer',
  is_active     BOOLEAN DEFAULT true,
  last_login    TIMESTAMP,
  created_at    TIMESTAMP DEFAULT NOW(),
  updated_at    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role);
```

### 角色定义

三个固定角色，权限写死在代码中：

| 操作 | super_admin | operator | viewer |
|------|:-----------:|:--------:|:------:|
| 查看所有数据 | O | O | O |
| 禁用/启用用户 | O | O | X |
| 重置用户密码 | O | O | X |
| 修改用户信息 | O | O | X |
| 创建/删除用户 | O | X | X |
| 手动调整余额 | O | X | X |
| 管理管理员 | O | X | X |

---

## 3. 后端 Admin 模块

### 模块结构

```
internal/modules/admin/
├── module.go                 # 模块初始化
├── middleware/
│   └── admin_auth.go         # Admin JWT 认证 + 角色权限中间件
├── model/
│   └── admin_user.go         # AdminUser 模型
├── repository/
│   └── admin_repo.go         # AdminUser CRUD
├── dto/
│   ├── request.go            # 登录、创建管理员等请求 DTO
│   └── response.go           # 响应 DTO
├── handler/
│   ├── auth_handler.go       # 登录/改密码
│   ├── admin_user_handler.go # 管理员 CRUD
│   ├── user_handler.go       # 用户管理
│   └── dashboard_handler.go  # 首页统计
├── service/
│   ├── auth_service.go       # Admin 认证逻辑
│   ├── admin_user_service.go # 管理员管理
│   ├── user_service.go       # 用户管理业务逻辑
│   └── dashboard_service.go  # 统计查询
└── routes.go                 # 路由注册
```

### API 路由

```
/api/v1/admin/
├── /auth                           # 公开（无需认证）
│   ├── POST /login                 # 管理员登录
│   └── POST /change-password       # 修改密码（需认证）
│
├── /admin-users                    # super_admin only
│   ├── GET /                       # 管理员列表
│   ├── POST /                      # 创建管理员
│   ├── PUT /:id                    # 修改管理员
│   └── DELETE /:id                 # 删除管理员
│
├── /users
│   ├── GET /                       # 用户列表（分页/搜索/筛选）  — viewer+
│   ├── GET /:id                    # 用户详情                    — viewer+
│   ├── PUT /:id                    # 修改用户信息                — operator+
│   ├── POST /:id/disable           # 禁用用户                    — operator+
│   ├── POST /:id/enable            # 启用用户                    — operator+
│   ├── POST /:id/reset-password    # 重置密码                    — operator+
│   └── POST /:id/adjust-balance    # 调整余额                    — super_admin only
│
└── /dashboard
    └── GET /stats                  # 首页统计数据                — viewer+
```

### 权限中间件

```go
// RequireAdmin 校验 Admin JWT，提取管理员信息到 context
func RequireAdmin() gin.HandlerFunc

// RequireRole 校验角色权限
func RequireRole(roles ...string) gin.HandlerFunc
```

路由注册时按角色分组，`RequireRole` 中间件应用在路由组级别。

### Dashboard 统计接口

`GET /api/v1/admin/dashboard/stats` 返回：

```json
{
  "total_users": 1234,
  "new_users_today": 12,
  "total_invested": "500000.00",
  "total_deposits": "1200000.00",
  "total_withdrawals": "300000.00",
  "active_investments": 89,
  "user_trend": [
    {"date": "2026-04-01", "count": 5},
    {"date": "2026-04-02", "count": 8}
  ],
  "recent_transactions": [...]
}
```

### 用户管理接口细节

**用户列表** `GET /users?page=1&limit=20&search=xxx&status=active`
- 分页、邮箱/姓名搜索、状态筛选
- 返回用户基本信息 + 钱包余额汇总

**用户详情** `GET /users/:id`
- 用户基本信息
- 钱包余额列表
- 最近交易记录
- 投资记录
- 结算记录

**调整余额** `POST /users/:id/adjust-balance`
- 请求体：`{ "currency": "USDT", "amount": "100.00", "reason": "手动补偿" }`
- 记录操作日志（谁、何时、调整多少、原因）

---

## 4. 前端 Admin 项目

### 技术栈

- Next.js + React + TypeScript
- `@arco-design/mobile-react`（移动端 UI 组件库）
- TanStack Query（数据请求）
- pnpm（包管理）

### 项目结构

```
admin/
├── src/
│   ├── app/
│   │   ├── login/              # 登录页
│   │   └── (admin)/            # 需认证的布局组
│   │       ├── layout.tsx      # 底部 TabBar + 顶部 NavBar 布局
│   │       ├── dashboard/      # 首页统计
│   │       ├── users/          # 用户管理
│   │       │   ├── page.tsx    # 用户列表（卡片列表 + 下拉加载）
│   │       │   └── [id]/       # 用户详情
│   │       └── admin-users/    # 管理员管理 (super_admin)
│   ├── components/
│   │   ├── layout/             # TabBar、NavBar
│   │   └── shared/             # 状态标签等通用组件
│   ├── hooks/
│   │   └── use-api.ts          # API 请求 hooks
│   ├── lib/
│   │   ├── api.ts              # axios + token 拦截器
│   │   └── auth.ts             # 登录态管理
│   └── types/
│       └── api.ts              # 类型定义
├── package.json
└── next.config.ts
```

### 移动端交互设计

**导航模式**
- 底部 TabBar：Dashboard / 用户管理 / 我的（管理员信息 + 设置）
- 顶部 NavBar：页面标题 + 返回按钮（子页面）
- super_admin 的 TabBar 多一个「管理员」Tab

**登录页**
- 邮箱 + 密码表单（移动端全屏布局）
- 登录后跳转 Dashboard

**Dashboard（首页）**
- 统计卡片网格（一行 2 个）：总用户数、今日新增、总投资额、总充值额、总提现额
- 近 7/30 天新用户趋势折线图
- 近期交易列表（最新 10 条，卡片样式）

**用户列表**
- 顶部搜索栏（邮箱/姓名）+ 筛选标签（全部/活跃/禁用）
- 卡片列表 + 下拉加载更多（非分页表格）
- 每张卡片显示：头像/姓名、邮箱、余额、状态标签
- 点击卡片进入详情页

**用户详情**
- 顶部用户基本信息卡片
- 操作按钮区：禁用/启用、重置密码、调整余额（按角色权限显隐）
- Tabs 切换：钱包 / 交易 / 投资 / 结算（每个 Tab 内为卡片列表 + 下拉加载）
- 编辑用户信息通过弹出 Popup 表单

**管理员管理**（仅 super_admin 可见 Tab）
- 管理员卡片列表
- 右上角「+」按钮创建管理员
- 长按/滑动操作：编辑、删除
- 角色分配通过 Picker 选择

---

## 6. 部署

### Docker

在 `deployments/` 中新增 admin 相关配置：
- `Dockerfile.admin` — 构建 admin 前端
- `docker-compose.yml` 中新增 `admin` 服务
- Nginx 反向代理新增 `/admin` 路径 → admin 容器

### 环境变量

```bash
# Admin JWT（独立于用户端）
SOVEREIGN_ADMIN__JWT_SECRET=xxx
SOVEREIGN_ADMIN__JWT_EXPIRY=24h
```

---

## 7. 初始化

系统部署后需要创建第一个 super_admin 账号。提供一个 CLI seed 命令：

```bash
go run scripts/seed_admin.go --email admin@example.com --password xxx --name "Super Admin"
```

---

## 8. 后续阶段规划

| 阶段 | 内容 |
|------|------|
| 阶段 2 | 充提币管理（交易列表/详情/审核）+ 投资管理（投资列表/详情） |
| 阶段 3 | 收益/结算管理 + 数据统计仪表板（更丰富的图表和导出） |
