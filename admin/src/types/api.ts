export interface AdminUser {
  id: string
  email: string
  name: string
  role: string
  is_active: boolean
  last_login: string | null
  created_at: string
}

export interface LoginResponse {
  token: string
  expires_at: string
  admin: AdminUser
}

export interface UserListItem {
  id: string
  email: string
  full_name: string
  phone: string
  language: string
  is_active: boolean
  balance: string
  created_at: string
}

export interface WalletInfo {
  id: string
  network: string
  address: string
  balance: string
  created_at: string
}

export interface TransactionInfo {
  id: string
  type: string
  amount: string
  currency: string
  status: string
  created_at: string
}

export interface InvestmentInfo {
  id: string
  amount: string
  currency: string
  status: string
  started_at: string
  matured_at: string | null
}

export interface SettlementInfo {
  id: string
  amount: string
  currency: string
  type: string
  status: string
  settled_at: string
}

export interface UserDetail {
  id: string
  email: string
  full_name: string
  phone: string
  language: string
  is_active: boolean
  created_at: string
  wallets: WalletInfo[]
  recent_transactions: TransactionInfo[]
  investments: InvestmentInfo[]
  recent_settlements: SettlementInfo[]
}

export interface DashboardStats {
  total_users: number
  new_users_today: number
  total_invested: string
  total_deposits: string
  total_withdrawals: string
  active_investments: number
  user_trend: Array<{ date: string; count: number }>
  recent_transactions: TransactionInfo[]
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  meta?: {
    total: number
    page: number
    limit: number
  }
}
