# 套利交易记录 - 内部 API 文档

> 供交易机器人/外部系统推送套利交易记录使用  
> Base URL: `http://172.31.1.31/api/v1/internal`

---

## 1. 创建单条交易记录

### `POST /trades`

创建一条套利交易记录，自动关联投资并更新投资收益（total_return / performance_fee / net_return）。

#### 请求

```json
{
  "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
  "pair": "BTC/KRW",
  "buy_exchange": "binance",
  "sell_exchange": "upbit",
  "buy_price": "101900000",
  "sell_price": "102100000",
  "amount": "150.50",
  "premium_pct": "0.1965",
  "pnl": "0.2958",
  "fee": "0.0003",
  "executed_at": "2026-04-01T10:30:00Z"
}
```

#### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `investment_id` | string (UUID) | ✅ | 关联的投资 ID，必须为 active 状态 |
| `pair` | string | ✅ | 交易对，如 `BTC/KRW`、`ETH/KRW`、`SOL/KRW`、`XRP/KRW` |
| `buy_exchange` | string | ✅ | 买入交易所，如 `binance`、`bybit` |
| `sell_exchange` | string | ✅ | 卖出交易所，如 `upbit`、`bithumb` |
| `buy_price` | string (decimal) | ✅ | 买入价格（KRW） |
| `sell_price` | string (decimal) | ✅ | 卖出价格（KRW） |
| `amount` | string (decimal) | ✅ | 交易数量（USDT） |
| `premium_pct` | string (decimal) | ✅ | 溢价率（百分比，如 `0.1965` 表示 0.1965%） |
| `pnl` | string (decimal) | ✅ | 盈亏金额（USDT） |
| `fee` | string (decimal) | ❌ | 手续费（USDT），默认 0 |
| `executed_at` | string (RFC3339) | ✅ | 执行时间，如 `2026-04-01T10:30:00Z` |

#### 响应

**成功 (201)**

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "investment_id": "e1a6d729-...",
    "pair": "BTC/KRW",
    "buy_exchange": "binance",
    "sell_exchange": "upbit",
    "buy_price": "101900000",
    "sell_price": "102100000",
    "amount": "150.50",
    "premium_pct": "0.1965",
    "pnl": "0.2958",
    "fee": "0.0003",
    "executed_at": "2026-04-01T10:30:00Z"
  }
}
```

**错误**

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | `INVALID_INVESTMENT` | 投资 ID 不存在 |
| 400 | `INVESTMENT_NOT_ACTIVE` | 投资不是 active 状态 |
| 400 | `INVALID_PRICE` | buy_price 或 sell_price 格式错误 |
| 400 | `INVALID_AMOUNT` | amount 格式错误 |
| 400 | `INVALID_PNL` | pnl 格式错误 |
| 400 | `INVALID_TIME` | executed_at 格式错误，需 RFC3339 |

---

## 2. 批量创建交易记录

### `POST /trades/batch`

一次推送多条交易记录。所有交易关联的投资必须为 active 状态。创建完成后自动更新各投资的收益。

#### 请求

```json
{
  "trades": [
    {
      "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
      "pair": "BTC/KRW",
      "buy_exchange": "binance",
      "sell_exchange": "upbit",
      "buy_price": "101900000",
      "sell_price": "102100000",
      "amount": "150.50",
      "premium_pct": "0.1965",
      "pnl": "0.2958",
      "fee": "0.0003",
      "executed_at": "2026-04-01T10:30:00Z"
    },
    {
      "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
      "pair": "ETH/KRW",
      "buy_exchange": "bybit",
      "sell_exchange": "bithumb",
      "buy_price": "5300000",
      "sell_price": "5350000",
      "amount": "85.30",
      "premium_pct": "0.9434",
      "pnl": "0.8047",
      "fee": "0.0008",
      "executed_at": "2026-04-01T10:31:00Z"
    }
  ]
}
```

#### 响应

**成功 (201)**

```json
{
  "success": true,
  "data": {
    "created": 2
  }
}
```

单条验证失败不会中断批量，跳过无效记录继续处理。`created` 返回实际成功创建的数量。

**错误**

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | `INVALID_INVESTMENT` | 某个 investment_id 不存在 |
| 400 | `INVESTMENT_NOT_ACTIVE` | 某个投资不是 active 状态 |
| 400 | `VALIDATION_ERROR` | trades 数组为空或字段校验失败 |

---

## 3. 业务逻辑

### 投资收益自动更新

每次创建交易记录后，系统自动重新计算关联投资的收益：

```
total_return    = SUM(trades.pnl)           -- 该投资所有交易的总盈亏
performance_fee = total_return * 50%        -- 平台绩效费（仅盈利时收取）
net_return      = total_return - performance_fee  -- 用户净收益
```

### 支持的交易对

| 交易对 | 韩国交易所 | 全球交易所 |
|--------|-----------|-----------|
| `BTC/KRW` | upbit, bithumb | binance, bybit |
| `ETH/KRW` | upbit, bithumb | binance, bybit |
| `SOL/KRW` | upbit, bithumb | binance, bybit |
| `XRP/KRW` | upbit, bithumb | binance, bybit |

### 认证

内部 API 不需要用户 JWT 认证。`user_id` 自动从 `investment_id` 关联获取。

> ⚠️ 生产环境应配置 IP 白名单或 API Key 限制访问。

---

## 4. 示例

### cURL - 创建单条

```bash
curl -X POST http://172.31.1.31/api/v1/internal/trades \
  -H "Content-Type: application/json" \
  -d '{
    "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
    "pair": "BTC/KRW",
    "buy_exchange": "binance",
    "sell_exchange": "upbit",
    "buy_price": "101900000",
    "sell_price": "102100000",
    "amount": "150.50",
    "premium_pct": "0.1965",
    "pnl": "0.2958",
    "fee": "0.0003",
    "executed_at": "2026-04-01T10:30:00Z"
  }'
```

### cURL - 批量创建

```bash
curl -X POST http://172.31.1.31/api/v1/internal/trades/batch \
  -H "Content-Type: application/json" \
  -d '{
    "trades": [
      {
        "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
        "pair": "BTC/KRW",
        "buy_exchange": "binance",
        "sell_exchange": "upbit",
        "buy_price": "101900000",
        "sell_price": "102100000",
        "amount": "150.50",
        "premium_pct": "0.1965",
        "pnl": "0.2958",
        "fee": "0.0003",
        "executed_at": "2026-04-01T10:30:00Z"
      }
    ]
  }'
```

### Python 示例

```python
import requests
from datetime import datetime

API_URL = "http://172.31.1.31/api/v1/internal/trades"

trade = {
    "investment_id": "e1a6d729-7b64-4605-a71a-d8889b897fd4",
    "pair": "BTC/KRW",
    "buy_exchange": "binance",
    "sell_exchange": "upbit",
    "buy_price": "101900000",
    "sell_price": "102100000",
    "amount": "150.50",
    "premium_pct": "0.1965",
    "pnl": "0.2958",
    "fee": "0.0003",
    "executed_at": datetime.utcnow().isoformat() + "Z"
}

resp = requests.post(API_URL, json=trade)
print(resp.json())
```
