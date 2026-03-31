"use client"

import { useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { api } from "@/lib/api-client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useT } from "@/hooks/use-t"
import { ArrowLeft, ArrowRight, Mail, CheckCircle2 } from "lucide-react"
import { toast } from "sonner"

type Step = "email" | "reset" | "done"

export default function ForgotPasswordPage() {
  const t = useT()
  const router = useRouter()
  const [step, setStep] = useState<Step>("email")
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  async function handleSendOtp(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setLoading(true)
    try {
      await api.post("/auth/forgot-password", { email })
      setStep("reset")
    } catch {
      setError(t("forgot.sendFailed"))
    } finally {
      setLoading(false)
    }
  }

  async function handleResetPassword(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setLoading(true)
    try {
      await api.post("/auth/reset-password", { email, code, new_password: newPassword })
      setStep("done")
      toast.success(t("forgot.resetSuccess"))
    } catch {
      setError(t("forgot.invalidCode"))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-[100dvh] flex-col items-center justify-center glow-bg-center px-6">
      <div className="mb-10 text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-ring">
          <span className="text-2xl font-bold text-primary-foreground">S</span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight">SOVEREIGN</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("auth.forgotPassword")}</p>
      </div>

      <div className="w-full max-w-sm">
        <div className="glass rounded-2xl p-6">
          {error && (
            <div className="mb-4 rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
          )}

          {step === "email" && (
            <form onSubmit={handleSendOtp} className="space-y-5">
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.email")}</Label>
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="h-12 rounded-xl border-0 bg-input text-base"
                />
              </div>
              <Button type="submit" className="h-12 w-full rounded-xl text-base font-semibold" disabled={loading}>
                {loading ? t("auth.sending") : t("auth.sendOtp")}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </form>
          )}

          {step === "reset" && (
            <form onSubmit={handleResetPassword} className="space-y-5">
              <div className="space-y-3 text-center">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                  <Mail className="h-6 w-6 text-primary" />
                </div>
                <p className="text-sm text-muted-foreground">
                  {t("forgot.codeSentTo")} <span className="text-foreground">{email}</span>
                </p>
              </div>
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("auth.verifyCode")}</Label>
                <Input
                  type="text"
                  inputMode="numeric"
                  placeholder={t("auth.otpPlaceholder")}
                  maxLength={6}
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
                  required
                  className="h-12 rounded-xl border-0 bg-input text-center text-xl tracking-[0.5em]"
                />
              </div>
              <div className="space-y-2">
                <Label className="text-xs uppercase tracking-wider text-muted-foreground">{t("forgot.newPassword")}</Label>
                <Input
                  type="password"
                  placeholder={t("auth.minPassword")}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  required
                  minLength={8}
                  className="h-12 rounded-xl border-0 bg-input text-base"
                />
              </div>
              <Button type="submit" className="h-12 w-full rounded-xl text-base font-semibold" disabled={loading || code.length !== 6 || newPassword.length < 8}>
                {loading ? t("common.loading") : t("forgot.resetPassword")}
              </Button>
              <button type="button" className="w-full text-center text-xs text-muted-foreground hover:text-foreground" onClick={() => { setStep("email"); setCode(""); setNewPassword("") }}>
                {t("forgot.resend")}
              </button>
            </form>
          )}

          {step === "done" && (
            <div className="space-y-4 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-success/10">
                <CheckCircle2 className="h-6 w-6 text-success" />
              </div>
              <p className="text-sm font-medium">{t("forgot.resetSuccess")}</p>
              <p className="text-xs text-muted-foreground">{t("forgot.canLogin")}</p>
              <Button className="h-12 w-full rounded-xl" onClick={() => router.push("/login")}>
                {t("auth.signIn")}
              </Button>
            </div>
          )}
        </div>

        <div className="mt-6 text-center">
          <Link href="/login" className="text-sm text-muted-foreground hover:text-foreground">
            <ArrowLeft className="mr-1 inline h-3.5 w-3.5" />
            {t("auth.signInLink")}
          </Link>
        </div>
      </div>
    </div>
  )
}
