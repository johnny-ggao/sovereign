"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useVerify2FA } from "@/hooks/use-api"
import { useAuthStore } from "@/stores/auth-store"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ApiError } from "@/lib/api-client"
import { useT } from "@/hooks/use-t"
import { ShieldCheck } from "lucide-react"

export default function Verify2FAPage() {
  const router = useRouter()
  const verify = useVerify2FA()
  const { setAuth } = useAuthStore()
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [email, setEmail] = useState("")
  const t = useT()

  useEffect(() => {
    const stored = sessionStorage.getItem("2fa_email")
    if (!stored) { router.replace("/login"); return }
    setEmail(stored)
  }, [router])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    try {
      const res = await verify.mutateAsync({ email, code })
      if (res.access_token && res.refresh_token && res.user) {
        sessionStorage.removeItem("2fa_email")
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
          <ShieldCheck className="h-7 w-7 text-primary-foreground" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight">{t("auth.twoFA")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("auth.twoFADesc")}</p>
      </div>

      <div className="w-full max-w-sm">
        <div className="glass rounded-2xl p-6">
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
            )}
            <div className="space-y-2">
              <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.verificationCode")}</Label>
              <Input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="000000"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
                required
                className="h-12 rounded-xl border-0 bg-input text-center text-xl tracking-[0.5em]"
                autoFocus
              />
            </div>
            <Button
              type="submit"
              className="h-12 w-full rounded-xl text-base font-semibold"
              disabled={verify.isPending || code.length !== 6}
            >
              {verify.isPending ? t("auth.verifying") : t("auth.verify")}
            </Button>
          </form>
        </div>
        <button
          onClick={() => { sessionStorage.removeItem("2fa_email"); router.push("/login") }}
          className="mt-6 block w-full text-center text-sm text-muted-foreground hover:text-foreground"
        >
          {t("auth.backToLogin")}
        </button>
      </div>
    </div>
  )
}
