"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { useLogin } from "@/hooks/use-api"
import { useAuthStore } from "@/stores/auth-store"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ApiError } from "@/lib/api-client"
import { useT } from "@/hooks/use-t"
import { ArrowRight } from "lucide-react"
import { GoogleLoginButton } from "@/components/auth/google-login-button"

export default function LoginPage() {
  const router = useRouter()
  const login = useLogin()
  const { setAuth } = useAuthStore()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const t = useT()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    try {
      const res = await login.mutateAsync({ email, password })
      if (res.requires_2fa) { router.push("/verify-2fa"); return }
      if (res.access_token && res.refresh_token && res.user) {
        setAuth(res.user, res.access_token, res.refresh_token)
        router.push("/dashboard")
      }
    } catch (err) {
      if (err instanceof ApiError) setError(err.message)
      else setError("An unexpected error occurred")
    }
  }

  return (
    <div className="flex min-h-[100dvh] flex-col items-center justify-center glow-bg-center px-6">
      <div className="mb-10 text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-ring">
          <span className="text-2xl font-bold text-primary-foreground">S</span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight">SOVEREIGN</h1>
        <p className="mt-1 text-sm text-muted-foreground">Institutional Crypto Fund</p>
      </div>

      <div className="w-full max-w-sm">
        <div className="glass rounded-2xl p-6">
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
            )}
            <div className="space-y-2">
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.email")}</Label>
              <Input type="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required className="h-12 rounded-xl border-0 bg-input text-base" />
            </div>
            <div className="space-y-2">
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.password")}</Label>
              <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required className="h-12 rounded-xl border-0 bg-input text-base" />
            </div>
            <Button type="submit" className="h-12 w-full rounded-xl text-base font-semibold" disabled={login.isPending}>
              {login.isPending ? t("auth.signingIn") : t("auth.signIn")}
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </form>

          <div className="mt-5 flex items-center gap-3">
            <div className="h-px flex-1 bg-accent" />
            <span className="text-xs text-muted-foreground">{t("auth.orContinueWith")}</span>
            <div className="h-px flex-1 bg-accent" />
          </div>

          <div className="mt-4">
            <GoogleLoginButton />
          </div>
        </div>
        <div className="mt-6 flex items-center justify-between px-2">
          <Link href="/register" className="text-sm text-primary hover:underline">{t("auth.register")}</Link>
          <Link href="/forgot-password" className="text-sm text-muted-foreground hover:text-foreground">{t("auth.forgotPassword")}</Link>
        </div>
      </div>
    </div>
  )
}
