import type { ApiResponse, AuthResponse } from "@/types/api"
import { useAuthStore } from "@/stores/auth-store"

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"

class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message)
    this.name = "ApiError"
  }
}

let refreshPromise: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  const { refreshToken, clearAuth } = useAuthStore.getState()
  if (!refreshToken) {
    clearAuth()
    return null
  }

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    const json: ApiResponse<AuthResponse> = await res.json()
    if (!json.success || !json.data?.access_token) {
      clearAuth()
      return null
    }

    const { setAuth } = useAuthStore.getState()
    setAuth(json.data.user!, json.data.access_token, json.data.refresh_token!)
    return json.data.access_token
  } catch {
    clearAuth()
    return null
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = typeof window !== "undefined"
    ? localStorage.getItem("access_token")
    : null

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  let res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  // 401 时尝试刷新 token 并重试一次
  if (res.status === 401 && token) {
    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null
      })
    }

    const newToken = await refreshPromise
    if (newToken) {
      headers["Authorization"] = `Bearer ${newToken}`
      res = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers,
      })
    }
  }

  if (res.status === 204) {
    return undefined as T
  }

  const json: ApiResponse<T> = await res.json()

  if (!json.success) {
    if (res.status === 401) {
      useAuthStore.getState().clearAuth()
      if (typeof window !== "undefined") {
        window.location.href = "/login"
      }
    }
    throw new ApiError(
      res.status,
      json.error?.code || "UNKNOWN",
      json.error?.message || "An error occurred",
    )
  }

  return json.data as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),

  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(path: string) =>
    request<T>(path, { method: "DELETE" }),
}

export { ApiError }
