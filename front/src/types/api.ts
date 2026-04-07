export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
  meta?: {
    total: number
    page: number
    per_page: number
  }
}

// Auth
export interface AuthResponse {
  access_token?: string
  refresh_token?: string
  expires_at?: number
  requires_2fa?: boolean
  user?: UserProfile
}

export interface UserProfile {
  id: string
  email: string
  full_name: string
  phone: string
  language: string
  kyc_status: string
  two_fa_enabled: boolean
  created_at?: string
}

// Dashboard
export interface DashboardSummary {
  portfolio: PortfolioSummary
  performance: PerformanceData
  premium: PremiumSummary
}

export interface PortfolioSummary {
  total_value: string
  total_value_usd: string
  available: string
  in_operation: string
  frozen: string
  currency: string
}

export interface PerformanceData {
  cumulative_return: string
  cumulative_return_pct: string
  annualized_return_pct: string
  high_water_mark: string
  chart: PerformancePoint[]
}

export interface PerformancePoint {
  date: string
  value: string
}

export interface PremiumSummary {
  current_pct: string
  pair: string
  trend: "up" | "down" | "flat"
  last_updated: string
}

// Premium
export interface PremiumTick {
  pair: string
  korean_price: string
  global_price: string
  premium_pct: string
  reverse_premium_pct: string
  source_kr: string
  source_gl: string
  latencies?: Record<string, number>
  timestamp: string
}

export interface PremiumLatest {
  ticks: PremiumTick[]
}

export interface PremiumHistory {
  pair: string
  points: PremiumTick[]
}

// Wallet
export interface WalletOverview {
  wallets: WalletBalance[]
  total_usdt: string
}

export interface WalletBalance {
  id: string
  currency: string
  available: string
  in_operation: string
  frozen: string
  earnings: string
  total: string
}

export interface DepositAddress {
  currency: string
  network: string
  address: string
}

export interface Transaction {
  id: string
  type: string
  currency: string
  network: string
  amount: string
  fee: string
  address: string
  tx_hash: string
  status: string
  confirmed_at: string | null
  created_at: string
}

export interface WhitelistAddress {
  id: string
  currency: string
  network: string
  address: string
  label: string
  cooldown_until: string
  is_active: boolean
}

// Investment
export interface Investment {
  id: string
  amount: string
  currency: string
  status: string
  total_return: string
  performance_fee: string
  net_return: string
  return_pct: string
  start_date: string
  end_date: string | null
}

export interface InvestmentList {
  investments: Investment[]
  summary: {
    total_invested: string
    total_return: string
    active_count: number
  }
}

// Trade
export interface Trade {
  id: string
  investment_id: string
  pair: string
  buy_exchange: string
  sell_exchange: string
  buy_price: string
  sell_price: string
  amount: string
  premium_pct: string
  pnl: string
  fee: string
  executed_at: string
}

export interface TradeList {
  trades: Trade[]
  summary: {
    total_trades: number
    total_pnl: string
    avg_premium_pct: string
    win_rate: string
  }
}

// Settlement
export interface Settlement {
  id: string
  investment_id: string
  period: string
  gross_return: string
  performance_fee: string
  fee_rate: string
  net_return: string
  trade_count: number
  avg_premium_pct: string
  report_url: string
  settled_at: string
}

export interface SettlementList {
  settlements: Settlement[]
  summary: {
    total_gross_return: string
    total_performance_fee: string
    total_net_return: string
    period_count: number
  }
}

// Settings
export interface NotificationPref {
  email_trade: boolean
  email_deposit: boolean
  email_withdraw: boolean
  email_settlement: boolean
  push_premium_alert: boolean
  push_trade: boolean
  push_deposit: boolean
  push_withdraw: boolean
  premium_threshold: number
}

export interface SecurityOverview {
  two_fa_enabled: boolean
  devices: LoginDevice[]
}

export interface LoginDevice {
  id: string
  user_agent: string
  ip: string
  location: string
  last_login: string
}

