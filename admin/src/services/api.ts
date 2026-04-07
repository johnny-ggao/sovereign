import { request } from '@umijs/max';

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

export async function adjustBalance(id: string, body: { amount: string; reason: string }) {
  return request<API.ApiResponse<null>>(`/users/${id}/adjust-balance`, {
    method: 'POST',
    data: body,
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
