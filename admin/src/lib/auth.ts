import { create } from 'zustand'
import type { AdminUser } from '@/types/api'

interface AuthState {
  token: string | null
  admin: AdminUser | null
  setAuth: (token: string, admin: AdminUser) => void
  logout: () => void
  isLoggedIn: () => boolean
}

function loadToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('admin_token')
}

function loadAdmin(): AdminUser | null {
  if (typeof window === 'undefined') return null
  const raw = localStorage.getItem('admin_user')
  if (!raw) return null
  try {
    return JSON.parse(raw) as AdminUser
  } catch {
    return null
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: loadToken(),
  admin: loadAdmin(),

  setAuth: (token: string, admin: AdminUser) => {
    localStorage.setItem('admin_token', token)
    localStorage.setItem('admin_user', JSON.stringify(admin))
    set({ token, admin })
  },

  logout: () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    set({ token: null, admin: null })
  },

  isLoggedIn: () => {
    return get().token !== null
  },
}))
