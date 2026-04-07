declare module 'slash2';
declare module '*.css';
declare module '*.less';
declare module '*.scss';
declare module '*.sass';
declare module '*.svg';
declare module '*.png';
declare module '*.jpg';
declare module '*.jpeg';
declare module '*.gif';
declare module '*.bmp';
declare module '*.tiff';
declare module 'omit.js';
declare module 'numeral';
declare module 'mockjs';
declare module 'react-fittext';

declare const REACT_APP_ENV: 'test' | 'dev' | 'pre' | false;

declare namespace API {
  interface AdminUser {
    id: string;
    email: string;
    name: string;
    role: 'super_admin' | 'operator' | 'viewer';
    is_active: boolean;
    last_login: string | null;
    created_at: string;
  }

  interface LoginResponse {
    token: string;
    expires_at: number;
    admin: AdminUser;
  }

  interface UserListItem {
    id: string;
    email: string;
    full_name: string;
    phone: string;
    language: string;
    is_active: boolean;
    balance: string;
    created_at: string;
  }

  interface UserDetail {
    id: string;
    email: string;
    full_name: string;
    phone: string;
    language: string;
    is_active: boolean;
    created_at: string;
    wallets: WalletInfo[];
    recent_transactions: TransactionInfo[];
    investments: InvestmentInfo[];
    recent_settlements: SettlementInfo[];
  }

  interface WalletInfo {
    currency: string;
    available: string;
    in_operation: string;
    frozen: string;
    earnings: string;
    total: string;
  }

  interface TransactionInfo {
    id: string;
    type: string;
    currency: string;
    network: string;
    amount: string;
    status: string;
    tx_hash: string;
    created_at: string;
  }

  interface InvestmentInfo {
    id: string;
    amount: string;
    currency: string;
    status: string;
    net_return: string;
    start_date: string;
  }

  interface SettlementInfo {
    id: string;
    period: string;
    net_return: string;
    fee_rate: string;
    settled_at: string;
  }

  interface InvestmentListItem {
    id: string;
    user_id: string;
    user_email: string;
    amount: string;
    currency: string;
    status: string;
    net_return: string;
    start_date: string;
    created_at: string;
  }

  interface DashboardStats {
    total_users: number;
    new_users_today: number;
    total_invested: string;
    total_deposits: string;
    total_withdrawals: string;
    active_investments: number;
    user_trend: { date: string; count: number }[];
    recent_transactions: TransactionInfo[];
  }

  interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: { code: string; message: string };
    meta?: { total: number; page: number; per_page: number };
  }
}
