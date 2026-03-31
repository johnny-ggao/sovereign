"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { useSendRegisterOTP, useRegister } from "@/hooks/use-api"
import { useAuthStore } from "@/stores/auth-store"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ApiError } from "@/lib/api-client"
import { useT } from "@/hooks/use-t"
import { ArrowRight, ArrowLeft, Mail, Shield } from "lucide-react"
import { GoogleLoginButton } from "@/components/auth/google-login-button"

export default function RegisterPage() {
  const router = useRouter()
  const sendOTP = useSendRegisterOTP()
  const register = useRegister()
  const t = useT()

  const [step, setStep] = useState<1 | 2>(1)
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [form, setForm] = useState({ password: "", full_name: "", phone: "" })
  const [error, setError] = useState("")

  function update(field: string, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    try {
      await sendOTP.mutateAsync({ email })
      setStep(2)
    } catch (err) {
      if (err instanceof ApiError) setError(err.message)
      else setError("An unexpected error occurred")
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    try {
      const res = await register.mutateAsync({ email, code, password: form.password, full_name: form.full_name, phone: form.phone || undefined })
      if (res.user && res.access_token && res.refresh_token) {
        useAuthStore.getState().setAuth(res.user, res.access_token, res.refresh_token)
      }
      router.push("/onboarding")
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
        <p className="mt-1 text-sm text-muted-foreground">{t("auth.createAccount")}</p>
      </div>

      <div className="w-full max-w-sm">
        {error && (
          <div className="mb-4 rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
        )}

        {step === 1 && (
          <div className="glass rounded-2xl p-6">
            <div className="mb-5 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10">
                <Mail className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="font-semibold">STEP 01</p>
                <p className="text-xs text-muted-foreground">{t("auth.sendOtp")}</p>
              </div>
            </div>
            <form onSubmit={handleSendOTP} className="space-y-4">
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.email")}</Label>
                <Input type="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required className="h-12 rounded-xl border-0 bg-input text-base" />
              </div>
              <Button type="submit" className="h-12 w-full rounded-xl text-base font-semibold" disabled={sendOTP.isPending}>
                {sendOTP.isPending ? t("auth.sending") : t("auth.sendOtp")}
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
        )}

        {step === 2 && (
          <div className="glass rounded-2xl p-6">
            <button type="button" onClick={() => { setStep(1); setError("") }} className="mb-4 flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
              <ArrowLeft className="h-3.5 w-3.5" />{t("auth.back")}
            </button>
            <div className="mb-5 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-success/10">
                <Shield className="h-5 w-5 text-success" />
              </div>
              <div>
                <p className="font-semibold">STEP 02</p>
                <p className="text-xs text-muted-foreground">{t("auth.verifyCode")}</p>
              </div>
            </div>
            <div className="mb-4 rounded-lg bg-success/10 px-4 py-3 text-sm text-success">{t("auth.otpSent")}</div>
            <form onSubmit={handleRegister} className="space-y-4">
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.verifyCode")}</Label>
                <Input placeholder={t("auth.otpPlaceholder")} value={code} onChange={(e) => setCode(e.target.value)} required maxLength={6} className="h-12 rounded-xl border-0 bg-input text-center text-xl tracking-[0.5em]" />
              </div>
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.fullName")}</Label>
                <Input value={form.full_name} onChange={(e) => update("full_name", e.target.value)} required className="h-12 rounded-xl border-0 bg-input text-base" />
              </div>
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.phone")}</Label>
                <Input value={form.phone} onChange={(e) => update("phone", e.target.value)} className="h-12 rounded-xl border-0 bg-input text-base" />
              </div>
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.password")}</Label>
                <Input type="password" value={form.password} onChange={(e) => update("password", e.target.value)} required minLength={8} className="h-12 rounded-xl border-0 bg-input text-base" />
                <p className="text-xs text-muted-foreground">{t("auth.minPassword")}</p>
              </div>
              <Button type="submit" className="h-12 w-full rounded-xl text-base font-semibold" disabled={register.isPending}>
                {register.isPending ? t("auth.creating") : t("auth.createBtn")}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </form>
          </div>
        )}

        <div className="mt-6 text-center">
          <Link href="/login" className="text-sm text-muted-foreground hover:text-foreground">
            {t("auth.hasAccount")} <span className="text-primary">{t("auth.signInLink")}</span>
          </Link>
        </div>
      </div>
    </div>
  )
}
