# Internal API 文档

供套利交易机器人和内部系统调用的 API 接口。

## 认证

所有 Internal API 使用 **API Key + IP 白名单** 认证。

### API Key

通过以下方式之一传递：

- **Header:** `X-Internal-Key: your-api-key`
- **Query:** `?key=your-api-key`

API Key 在 `.env` 中配置：`SOVEREIGN_INTERNAL__API_KEY=your-api-key`

### IP 白名单

在 `config.yaml` 或 `.env` 中配置允许访问的 IP 地址：

```yaml
internal:
  allowed_ips:
    - "127.0.0.1"
    - "172.31.0.0/16"   # 支持 CIDR 格式
```

设置为 `*` 允许所有 IP 访问（仅用于开发）。

---

## 基础信息

- **Base URL:** `/api/v1/internal`
- **Content-Type:** `application/json`
- **响应格式:**

```json
{
  "success": true,
  "data": { ... }
}
```

错误响应：

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "error description"
  }
}
```

---

## 套利交易接口

### 推送单条交易记录

```
POST /api/v1/internal/trades
```

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `pair` | string | 是 | 交易对，如 `USDT/KRW` |
| `buy_exchange` | string | 是 | 买入交易所，如 `Binance` |
| `sell_exchange` | string | 是 | 卖出交易所，如 `Upbit` |
| `buy_price` | string | 是 | 买入价格 |
| `sell_price` | string | 是 | 卖出价格 |
| `amount` | string | 是 | 交易金额 |
| `premium_pct` | string | 是 | 溢价百分比 |
| `pnl` | string | 是 | 盈亏金额 |
| `fee` | string | 否 | 手续费，默认 `"0"` |
| `executed_at` | string | 是 | 执行时间，RFC3339 格式 |
| `investment_id` | string | 否 | 关联的投资 ID（UUID） |

**请求示例：**

```json
{
  "pair": "USDT/KRW",
  "buy_exchange": "Binance",
  "sell_exchange": "Upbit",
  "buy_price": "1.0000",
  "sell_price": "1.0350",
  "amount": "10000.00",
  "premium_pct": "3.50",
  "pnl": "350.00",
  "fee": "10.00",
  "executed_at": "2026-04-07T12:00:00Z"
}
```

**成功响应（201）：**

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "pair": "USDT/KRW",
    "buy_exchange": "Binance",
    "sell_exchange": "Upbit",
    "buy_price": "1.0000",
    "sell_price": "1.0350",
    "amount": "10000.00",
    "premium_pct": "3.50",
    "pnl": "350.00",
    "fee": "10.00",
    "executed_at": "2026-04-07T12:00:00Z"
  }
}
```

---

### 批量推送交易记录

```
POST /api/v1/internal/trades/batch
```

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `trades` | array | 是 | 交易记录数组（至少 1 条），每条格式同单条推送 |

**请求示例：**

```json
{
  "trades": [
    {
      "pair": "USDT/KRW",
      "buy_exchange": "Binance",
      "sell_exchange": "Upbit",
      "buy_price": "1.0000",
      "sell_price": "1.0350",
      "amount": "10000.00",
      "premium_pct": "3.50",
      "pnl": "350.00",
      "fee": "10.00",
      "executed_at": "2026-04-07T12:00:00Z"
    },
    {
      "pair": "USDT/KRW",
      "buy_exchange": "Bybit",
      "sell_exchange": "Bithumb",
      "buy_price": "1.0010",
      "sell_price": "1.0320",
      "amount": "5000.00",
      "premium_pct": "3.10",
      "pnl": "155.00",
      "fee": "5.00",
      "executed_at": "2026-04-07T12:05:00Z"
    }
  ]
}
```

**成功响应（201）：**

```json
{
  "success": true,
  "data": {
    "created": 2
  }
}
```

---

## 结算接口

### 手动触发结算

```
POST /api/v1/internal/settlement/trigger
```

手动触发今日的结算任务。通常由定时任务（cron）自动执行，此接口用于调试或手动补偿。

**请求体：** 无

**成功响应（200）：**

```json
{
  "success": true,
  "message": "settlement completed"
}
```

**失败响应（500）：**

```json
{
  "success": false,
  "error": "error description"
}
```

---

## 调用示例

### cURL

```bash
# 单条推送
curl -X POST http://your-server/api/v1/internal/trades \
  -H "Content-Type: application/json" \
  -H "X-Internal-Key: your-api-key" \
  -d '{
    "pair": "USDT/KRW",
    "buy_exchange": "Binance",
    "sell_exchange": "Upbit",
    "buy_price": "1.0000",
    "sell_price": "1.0350",
    "amount": "10000.00",
    "premium_pct": "3.50",
    "pnl": "350.00",
    "executed_at": "2026-04-07T12:00:00Z"
  }'

# 批量推送
curl -X POST http://your-server/api/v1/internal/trades/batch \
  -H "Content-Type: application/json" \
  -H "X-Internal-Key: your-api-key" \
  -d '{
    "trades": [
      { "pair": "USDT/KRW", "buy_exchange": "Binance", "sell_exchange": "Upbit", "buy_price": "1.0", "sell_price": "1.035", "amount": "10000", "premium_pct": "3.5", "pnl": "350", "executed_at": "2026-04-07T12:00:00Z" }
    ]
  }'

# 手动触发结算
curl -X POST http://your-server/api/v1/internal/settlement/trigger \
  -H "X-Internal-Key: your-api-key"
```

### Python

```python
import requests

API_URL = "http://your-server/api/v1/internal"
API_KEY = "your-api-key"
HEADERS = {
    "Content-Type": "application/json",
    "X-Internal-Key": API_KEY,
}

# 推送单条交易
trade = {
    "pair": "USDT/KRW",
    "buy_exchange": "Binance",
    "sell_exchange": "Upbit",
    "buy_price": "1.0000",
    "sell_price": "1.0350",
    "amount": "10000.00",
    "premium_pct": "3.50",
    "pnl": "350.00",
    "fee": "10.00",
    "executed_at": "2026-04-07T12:00:00Z",
}

resp = requests.post(f"{API_URL}/trades", json=trade, headers=HEADERS)
print(resp.json())

# 批量推送
trades = {"trades": [trade, trade]}
resp = requests.post(f"{API_URL}/trades/batch", json=trades, headers=HEADERS)
print(resp.json())  # {"success": true, "data": {"created": 2}}
```
