import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api'
import type {
  ApiResponse,
  LoginResponse,
  AdminUser,
  DashboardStats,
  UserListItem,
  UserDetail,
} from '@/types/api'

interface LoginParams {
  email: string
  password: string
}

interface ChangePasswordParams {
  current_password: string
  new_password: string
}

interface UserListParams {
  page?: number
  limit?: number
  search?: string
}

interface UpdateUserParams {
  id: string
  data: Partial<Pick<UserDetail, 'full_name' | 'phone' | 'language'>>
}

interface AdjustBalanceParams {
  id: string
  amount: string
  reason: string
}

interface CreateAdminParams {
  email: string
  name: string
  password: string
  role: string
}

interface UpdateAdminParams {
  id: string
  data: Partial<Pick<AdminUser, 'name' | 'role' | 'is_active'>>
}

// --- Auth hooks ---

export function useLogin() {
  return useMutation({
    mutationFn: async (params: LoginParams) => {
      const res = await apiClient.post<ApiResponse<LoginResponse>>(
        '/auth/login',
        params
      )
      return res.data.data!
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (params: ChangePasswordParams) => {
      const res = await apiClient.post<ApiResponse<null>>(
        '/auth/change-password',
        params
      )
      return res.data
    },
  })
}

export function useCurrentAdmin() {
  return useQuery({
    queryKey: ['currentAdmin'],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<AdminUser>>('/auth/me')
      return res.data.data!
    },
  })
}

// --- Dashboard hooks ---

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboardStats'],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<DashboardStats>>(
        '/dashboard/stats'
      )
      return res.data.data!
    },
  })
}

// --- User hooks ---

export function useUserList(params: UserListParams = {}) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<UserListItem[]>>('/users', {
        params,
      })
      return {
        data: res.data.data!,
        meta: res.data.meta,
      }
    },
  })
}

export function useUserDetail(id: string | null) {
  return useQuery({
    queryKey: ['user', id],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<UserDetail>>(`/users/${id}`)
      return res.data.data!
    },
    enabled: !!id,
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: UpdateUserParams) => {
      const res = await apiClient.put<ApiResponse<UserDetail>>(
        `/users/${id}`,
        data
      )
      return res.data.data!
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', variables.id] })
    },
  })
}

export function useDisableUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const res = await apiClient.post<ApiResponse<null>>(
        `/users/${id}/disable`
      )
      return res.data
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', id] })
    },
  })
}

export function useEnableUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const res = await apiClient.post<ApiResponse<null>>(
        `/users/${id}/enable`
      )
      return res.data
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', id] })
    },
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await apiClient.post<ApiResponse<{ temporary_password: string }>>(
        `/users/${id}/reset-password`
      )
      return res.data.data!
    },
  })
}

export function useAdjustBalance() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, amount, reason }: AdjustBalanceParams) => {
      const res = await apiClient.post<ApiResponse<null>>(
        `/users/${id}/adjust-balance`,
        { amount, reason }
      )
      return res.data
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', variables.id] })
    },
  })
}

// --- Admin user management hooks ---

export function useAdminList() {
  return useQuery({
    queryKey: ['adminUsers'],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<AdminUser[]>>('/admin-users')
      return res.data.data!
    },
  })
}

export function useCreateAdmin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: CreateAdminParams) => {
      const res = await apiClient.post<ApiResponse<AdminUser>>(
        '/admin-users',
        params
      )
      return res.data.data!
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
    },
  })
}

export function useUpdateAdmin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: UpdateAdminParams) => {
      const res = await apiClient.put<ApiResponse<AdminUser>>(
        `/admin-users/${id}`,
        data
      )
      return res.data.data!
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
    },
  })
}

export function useDeleteAdmin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const res = await apiClient.delete<ApiResponse<null>>(
        `/admin-users/${id}`
      )
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
    },
  })
}
