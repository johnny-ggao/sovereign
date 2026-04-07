'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { NavBar, Input, Button, Toast } from '@arco-design/mobile-react'
import { useLogin } from '@/hooks/use-api'
import { useAuthStore } from '@/lib/auth'

export default function LoginPage() {
  const router = useRouter()
  const login = useLogin()
  const setAuth = useAuthStore((s) => s.setAuth)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleLogin = async () => {
    if (!email.trim() || !password.trim()) {
      Toast.toast('Please enter email and password')
      return
    }

    try {
      const result = await login.mutateAsync({ email, password })
      setAuth(result.token, result.admin)
      router.replace('/dashboard')
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'Login failed, please try again'
      Toast.error(message)
    }
  }

  return (
    <div style={{ minHeight: '100vh', background: '#f5f6f7' }}>
      <NavBar title="Sovereign Admin" leftContent={null} />
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '80px 24px 0',
        }}
      >
        <div
          style={{
            fontSize: 48,
            marginBottom: 16,
          }}
        >
          🔐
        </div>
        <h2
          style={{
            fontSize: 20,
            fontWeight: 600,
            marginBottom: 32,
            color: '#1d2129',
          }}
        >
          Admin Login
        </h2>
        <div style={{ width: '100%', maxWidth: 360 }}>
          <Input
            placeholder="Email"
            value={email}
            onChange={(_e, val) => setEmail(val)}
            border="all"
            style={{ marginBottom: 16 }}
          />
          <Input
            placeholder="Password"
            type="password"
            value={password}
            onChange={(_e, val) => setPassword(val)}
            border="all"
            style={{ marginBottom: 24 }}
          />
          <Button
            onClick={handleLogin}
            style={{
              width: '100%',
              height: 44,
              borderRadius: 8,
              fontSize: 16,
              fontWeight: 500,
              background: '#165dff',
              color: '#fff',
              border: 'none',
              opacity: login.isPending ? 0.7 : 1,
            }}
            disabled={login.isPending}
          >
            {login.isPending ? 'Logging in...' : 'Login'}
          </Button>
        </div>
      </div>
    </div>
  )
}
