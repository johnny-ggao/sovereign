import { request } from '@umijs/max';

const adminApiPrefix = '/api/v1/admin';
const tradeTemplatePath = '/trades/template';

/** Auth */
export async function login(body: { email: string; password: string }) {
  return request<API.ApiResponse<API.LoginResponse>>('/auth/login', {
    method: 'POST',
    data: body,
  });
}

export async function getMe(options?: { [key: string]: any }) {
  return request<API.ApiResponse<API.AdminUser>>('/auth/me', {
    method: 'GET',
    ...(options || {}),
  });
}

export async function changePassword(body: { old_password: string; new_password: string }) {
  return request<API.ApiResponse<null>>('/auth/change-password', {
    method: 'POST',
    data: body,
  });
}

/** Dashboard */
export async function getDashboardStats() {
  return request<API.ApiResponse<API.DashboardStats>>('/dashboard/stats', {
    method: 'GET',
  });
}

/** Users */
export async function getUsers(params: {
  page?: number;
  per_page?: number;
  email?: string;
}) {
  return request<API.ApiResponse<API.UserListItem[]>>('/users', {
    method: 'GET',
    params,
  });
}

export async function getUserDetail(id: string) {
  return request<API.ApiResponse<API.UserDetail>>(`/users/${id}`, {
    method: 'GET',
  });
}

export async function updateUser(id: string, body: Record<string, any>) {
  return request<API.ApiResponse<API.UserDetail>>(`/users/${id}`, {
    method: 'PUT',
    data: body,
  });
}

export async function disableUser(id: string) {
  return request<API.ApiResponse<null>>(`/users/${id}/disable`, {
    method: 'POST',
  });
}

export async function enableUser(id: string) {
  return request<API.ApiResponse<null>>(`/users/${id}/enable`, {
    method: 'POST',
  });
}

export async function resetUserPassword(id: string) {
  return request<API.ApiResponse<{ temporary_password: string }>>(`/users/${id}/reset-password`, {
    method: 'POST',
  });
}

export async function adjustBalance(id: string, body: { currency: string; amount: string; reason: string }) {
  return request<API.ApiResponse<null>>(`/users/${id}/adjust-balance`, {
    method: 'POST',
    data: body,
  });
}

/** Investments */
export async function getInvestments(params: {
  page?: number;
  limit?: number;
  search?: string;
  status?: string;
  sort_by?: string;
  sort_order?: string;
}) {
  return request<API.ApiResponse<API.InvestmentListItem[]>>('/investments', {
    method: 'GET',
    params,
  });
}

/** Admin Users */
export async function getAdmins() {
  return request<API.ApiResponse<API.AdminUser[]>>('/admin-users', {
    method: 'GET',
  });
}

export async function createAdmin(body: {
  email: string;
  password: string;
  name: string;
  role: string;
}) {
  return request<API.ApiResponse<API.AdminUser>>('/admin-users', {
    method: 'POST',
    data: body,
  });
}

export async function updateAdmin(id: string, body: { name?: string; role?: string }) {
  return request<API.ApiResponse<API.AdminUser>>(`/admin-users/${id}`, {
    method: 'PUT',
    data: body,
  });
}

export async function deleteAdmin(id: string) {
  return request<API.ApiResponse<null>>(`/admin-users/${id}`, {
    method: 'DELETE',
  });
}

/** Audit Logs */
export async function getAuditLogs(params: {
  page?: number;
  limit?: number;
  action?: string;
  admin_id?: string;
  date_from?: string;
  date_to?: string;
}) {
  return request<API.ApiResponse<API.AuditLog[]>>('/audit-logs', {
    method: 'GET',
    params,
  });
}

/** Trades */
export async function getTrades(params: {
  page?: number;
  limit?: number;
  pair?: string;
  date_from?: string;
  date_to?: string;
}) {
  return request<API.ApiResponse<API.TradeListItem[]>>('/trades', {
    method: 'GET',
    params,
  });
}

export async function getTradeStats() {
  return request<API.ApiResponse<API.TradeStats>>('/trades/stats', {
    method: 'GET',
  });
}

export function getTradeTemplateUrl() {
  return `${adminApiPrefix}${tradeTemplatePath}`;
}

export async function deleteTrade(id: string) {
  return request<API.ApiResponse<null>>(`/trades/${id}`, { method: 'DELETE' });
}

export async function importTrades(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request<API.ApiResponse<API.TradeImportResult>>('/trades/import', {
    method: 'POST',
    data: formData,
    requestType: 'form',
  });
}

/** Reset 2FA */
export async function resetUser2FA(id: string) {
  return request<API.ApiResponse<null>>(`/users/${id}/reset-2fa`, {
    method: 'POST',
  });
}

/** Transactions */
export async function getTransactions(params: {
  page?: number;
  limit?: number;
  type?: string;
  search?: string;
  status?: string;
  date_from?: string;
  date_to?: string;
}) {
  return request<API.ApiResponse<API.TransactionListItem[]>>('/transactions', {
    method: 'GET',
    params,
  });
}

export async function getTransactionStats() {
  return request<API.ApiResponse<API.TransactionStats>>('/transactions/stats', {
    method: 'GET',
  });
}
