import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type {
  AuthResponse,
  DashboardSummary,
  PerformanceData,
  PremiumLatest,
  PremiumHistory,
  WalletOverview,
  DepositAddress,
  Transaction,
  WhitelistAddress,
  InvestmentList,
  Investment,
  Settlement,
  SettlementList,
  NotificationPref,
  SecurityOverview,
  UserProfile,
} from "@/types/api"

// Auth
export function useLogin() {
  return useMutation({
    mutationFn: (data: { email: string; password: string }) =>
      api.post<AuthResponse>("/auth/login", data),
  })
}

export function useVerify2FA() {
  return useMutation({
    mutationFn: (data: { email: string; code: string }) =>
      api.post<AuthResponse>("/auth/verify-2fa", data),
  })
}

export function useGoogleLogin() {
  return useMutation({
    mutationFn: (idToken: string) =>
      api.post<AuthResponse>("/auth/google", { id_token: idToken }),
  })
}

export function useSendRegisterOTP() {
  return useMutation({
    mutationFn: (data: { email: string }) =>
      api.post("/auth/register/send-otp", data),
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: (data: { email: string; code: string; password: string; full_name: string; phone?: string }) =>
      api.post<AuthResponse>("/auth/register", data),
  })
}

export function useRefreshToken() {
  return useMutation({
    mutationFn: (refreshToken: string) =>
      api.post<AuthResponse>("/auth/refresh", { refresh_token: refreshToken }),
  })
}

// Dashboard
export function useDashboardSummary() {
  return useQuery({
    queryKey: ["dashboard", "summary"],
    queryFn: () => api.get<DashboardSummary>("/dashboard/summary"),
  })
}

export function usePerformance(period: string = "1M") {
  return useQuery({
    queryKey: ["dashboard", "performance", period],
    queryFn: () => api.get<PerformanceData>(`/dashboard/performance?period=${period}`),
  })
}

// Premium
export function usePremiumLatest() {
  return useQuery({
    queryKey: ["premium", "latest"],
    queryFn: () => api.get<PremiumLatest>("/premium/latest"),
    refetchInterval: 10000,
  })
}

export function usePremiumHistory(pair: string = "BTC/KRW", limit: number = 500) {
  return useQuery({
    queryKey: ["premium", "history", pair, limit],
    queryFn: () => api.get<PremiumHistory>(`/premium/history?pair=${encodeURIComponent(pair)}&limit=${limit}`),
  })
}

// Wallet
export function useWallets() {
  return useQuery({
    queryKey: ["wallets"],
    queryFn: () => api.get<WalletOverview>("/wallets"),
  })
}

export function useDepositAddress() {
  return useMutation({
    mutationFn: (data: { currency: string; network: string }) =>
      api.post<DepositAddress>("/wallets/deposit-address", data),
  })
}

export function useWithdraw() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { currency: string; network: string; address: string; amount: string; two_fa_code: string }) =>
      api.post("/wallets/withdraw", data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["wallets"] })
      qc.invalidateQueries({ queryKey: ["transactions"] })
    },
  })
}

export function useClaimEarnings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post("/wallets/claim-earnings"),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["wallets"] })
    },
  })
}

export function useTransactions(type?: string, page: number = 1) {
  const params = new URLSearchParams({ page: String(page) })
  if (type) params.set("type", type)

  return useQuery({
    queryKey: ["transactions", type, page],
    queryFn: () => api.get<Transaction[]>(`/transactions?${params}`),
  })
}

export function useWhitelistAddresses() {
  return useQuery({
    queryKey: ["whitelist-addresses"],
    queryFn: () => api.get<WhitelistAddress[]>("/wallets/addresses"),
  })
}

export function useAddWhitelistAddress() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { currency: string; network: string; address: string; label?: string }) =>
      api.post("/wallets/addresses", data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["whitelist-addresses"] }),
  })
}

export function useRemoveWhitelistAddress() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/wallets/addresses/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["whitelist-addresses"] }),
  })
}

// Investment
export function useInvestments() {
  return useQuery({
    queryKey: ["investments"],
    queryFn: () => api.get<InvestmentList>("/investments"),
  })
}

export function useInvestment(id: string) {
  return useQuery({
    queryKey: ["investments", id],
    queryFn: () => api.get<Investment>(`/investments/${id}`),
    enabled: !!id,
  })
}

export function useCreateInvestment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { amount: string; currency?: string }) =>
      api.post("/investments", data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["investments"] })
      qc.invalidateQueries({ queryKey: ["wallets"] })
    },
  })
}

export function useRedeemInvestment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (investmentId: string) =>
      api.post("/investments/redeem", { investment_id: investmentId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["investments"] }),
  })
}

// Settlement
export function useSettlements() {
  return useQuery({
    queryKey: ["settlements"],
    queryFn: () => api.get<SettlementList>("/settlements"),
  })
}

export function useSettlement(id: string) {
  return useQuery({
    queryKey: ["settlements", id],
    queryFn: () => api.get<Settlement>(`/settlements/${id}`),
    enabled: !!id,
  })
}

// Settings
export function useProfile() {
  return useQuery({
    queryKey: ["profile"],
    queryFn: () => api.get<UserProfile>("/settings/profile"),
  })
}

export function useUpdateProfile() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { full_name?: string; phone?: string }) =>
      api.put<UserProfile>("/settings/profile", data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["profile"] }),
  })
}

export function useNotificationPref() {
  return useQuery({
    queryKey: ["notification-pref"],
    queryFn: () => api.get<NotificationPref>("/settings/notifications"),
  })
}

export function useUpdateNotificationPref() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: Partial<NotificationPref>) =>
      api.put("/settings/notifications", data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notification-pref"] }),
  })
}

export function useSecurityOverview() {
  return useQuery({
    queryKey: ["security"],
    queryFn: () => api.get<SecurityOverview>("/settings/security"),
  })
}

