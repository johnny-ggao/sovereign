"use client"

import { useState } from "react"
import { useSearchParams } from "next/navigation"
import { useProfile, useUpdateProfile, useSecurityOverview, useNotificationPref, useUpdateNotificationPref } from "@/hooks/use-api"
import { api } from "@/lib/api-client"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { Globe, Smartphone, Trash2, Copy, Check, ShieldCheck } from "lucide-react"
import { QRCodeSVG } from "qrcode.react"
import { formatDateTime } from "@/lib/format"
import { toast } from "sonner"
import { useT } from "@/hooks/use-t"
import { useLocaleStore } from "@/stores/locale-store"
import { localeLabels, type Locale } from "@/i18n"

export default function SettingsPage() {
  const searchParams = useSearchParams()
  const tab = searchParams.get("tab") || "profile"
  const t = useT()
  const { locale, setLocale } = useLocaleStore()
  const { data: profile, isLoading } = useProfile()
  const updateProfile = useUpdateProfile()
  const { data: security } = useSecurityOverview()
  const { data: notifPref } = useNotificationPref()
  const updateNotif = useUpdateNotificationPref()
  const [profileForm, setProfileForm] = useState({ full_name: "", phone: "" })
  const [passwordForm, setPasswordForm] = useState({ current_password: "", new_password: "" })
  const [twoFASetup, setTwoFASetup] = useState<{ secret: string; qr_code_url: string } | null>(null)
  const [twoFACode, setTwoFACode] = useState("")
  const [twoFALoading, setTwoFALoading] = useState(false)
  const [secretCopied, setSecretCopied] = useState(false)
  const [initialized, setInitialized] = useState(false)

  if (profile && !initialized) {
    setProfileForm({ full_name: profile.full_name, phone: profile.phone })
    setInitialized(true)
  }

  async function handleUpdateProfile() {
    await updateProfile.mutateAsync(profileForm)
    toast.success("Profile updated")
  }

  async function handleChangePassword() {
    try {
      await api.put("/settings/password", passwordForm)
      setPasswordForm({ current_password: "", new_password: "" })
      toast.success("Password changed")
    } catch {
      toast.error("Failed to change password")
    }
  }

  async function handleSetup2FA() {
    try {
      const res = await api.post<{ secret: string; qr_code_url: string }>("/settings/2fa/setup")
      setTwoFASetup(res)
      setTwoFACode("")
    } catch {
      toast.error("Failed to setup 2FA")
    }
  }

  async function handleVerify2FA() {
    setTwoFALoading(true)
    try {
      await api.post("/settings/2fa/verify", { code: twoFACode })
      toast.success(t("settings.twoFaSuccess"))
      setTwoFASetup(null)
      setTwoFACode("")
      window.location.reload()
    } catch {
      toast.error(t("settings.disabled"))
    } finally {
      setTwoFALoading(false)
    }
  }

  function copySecret() {
    if (twoFASetup) {
      navigator.clipboard.writeText(twoFASetup.secret)
      setSecretCopied(true)
      toast.success(t("settings.copySecret"))
      setTimeout(() => setSecretCopied(false), 2000)
    }
  }

  function toggleNotif(key: string, value: boolean) {
    updateNotif.mutate({ [key]: value })
  }

  if (isLoading) return <Skeleton className="h-96" />

  return (
    <div className="space-y-6 glow-bg">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">
          {t(`settings.${tab}`)}
        </h1>
      </div>

      {tab === "profile" && (
        <div className="space-y-6">
          <Card className="glass border-0 rounded-2xl">
            <CardHeader>
              <CardTitle>{t("settings.personalInfo")}</CardTitle>
              <CardDescription>{t("settings.updateProfile")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t("auth.email")}</Label>
                <Input value={profile?.email || ""} disabled className="bg-input opacity-60" />
              </div>
              <div className="space-y-2">
                <Label>{t("auth.fullName")}</Label>
                <Input
                  value={profileForm.full_name}
                  onChange={(e) => setProfileForm((p) => ({ ...p, full_name: e.target.value }))}
                  className="bg-input"
                />
              </div>
              <div className="space-y-2">
                <Label>{t("auth.phone")}</Label>
                <Input
                  value={profileForm.phone}
                  onChange={(e) => setProfileForm((p) => ({ ...p, phone: e.target.value }))}
                  className="bg-input"
                />
              </div>
              <Button onClick={handleUpdateProfile} disabled={updateProfile.isPending}>
                {updateProfile.isPending ? t("common.saving") : t("common.save")}
              </Button>
            </CardContent>
          </Card>

          <Card className="glass border-0 rounded-2xl">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Globe className="h-4 w-4" />Language
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                {(Object.entries(localeLabels) as [Locale, string][]).map(([code, label]) => (
                  <Button
                    key={code}
                    variant={locale === code ? "default" : "outline"}
                    size="sm"
                    onClick={() => {
                      setLocale(code)
                      api.put("/settings/language", { language: code })
                    }}
                  >
                    {label}
                  </Button>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {tab === "security" && (
        <div className="space-y-6">
          <Card className="glass border-0 rounded-2xl">
            <CardHeader>
              <CardTitle>{t("settings.changePassword")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t("settings.currentPassword")}</Label>
                <Input
                  type="password"
                  value={passwordForm.current_password}
                  onChange={(e) => setPasswordForm((p) => ({ ...p, current_password: e.target.value }))}
                  className="bg-input"
                />
              </div>
              <div className="space-y-2">
                <Label>{t("settings.newPassword")}</Label>
                <Input
                  type="password"
                  value={passwordForm.new_password}
                  onChange={(e) => setPasswordForm((p) => ({ ...p, new_password: e.target.value }))}
                  className="bg-input"
                />
              </div>
              <Button onClick={handleChangePassword}>{t("settings.changePassword")}</Button>
            </CardContent>
          </Card>

          <Card className="glass border-0 rounded-2xl">
            <CardHeader>
              <CardTitle>{t("settings.twoFa")}</CardTitle>
              <CardDescription>
                {security?.two_fa_enabled ? t("settings.twoFaEnabled") : t("settings.twoFaSetup")}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <Badge variant="outline" className={security?.two_fa_enabled ? "border-success/30 text-success" : "border-destructive/30 text-destructive"}>
                  {security?.two_fa_enabled ? t("settings.enabled") : t("settings.disabled")}
                </Badge>
                {!security?.two_fa_enabled && !twoFASetup && (
                  <Button variant="outline" size="sm" onClick={handleSetup2FA}>
                    {t("settings.setup2FA")}
                  </Button>
                )}
              </div>

              {twoFASetup && (
                <div className="space-y-5 rounded-xl bg-accent/30 p-5">
                  {/* QR Code */}
                  <div className="flex flex-col items-center gap-3">
                    <div className="rounded-xl bg-white p-3">
                      <QRCodeSVG value={twoFASetup.qr_code_url} size={180} />
                    </div>
                    <p className="text-center text-sm text-muted-foreground">{t("settings.scanQr")}</p>
                  </div>

                  {/* Manual Secret */}
                  <div>
                    <p className="mb-2 text-xs text-muted-foreground">{t("settings.orManual")}</p>
                    <div className="flex items-center gap-2">
                      <code className="flex-1 break-all rounded-lg bg-input px-3 py-2 font-mono text-xs tracking-wider">{twoFASetup.secret}</code>
                      <button onClick={copySecret} className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-accent">
                        {secretCopied ? <Check className="h-4 w-4 text-success" /> : <Copy className="h-4 w-4" />}
                      </button>
                    </div>
                  </div>

                  {/* Verify Code */}
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">{t("settings.enterCode")}</Label>
                    <Input
                      placeholder="000000"
                      maxLength={6}
                      value={twoFACode}
                      onChange={(e) => setTwoFACode(e.target.value)}
                      className="h-12 rounded-xl border-0 bg-input text-center text-xl tracking-[0.5em]"
                    />
                  </div>

                  <div className="flex gap-2">
                    <Button className="flex-1 h-11 rounded-xl" disabled={twoFALoading || twoFACode.length !== 6} onClick={handleVerify2FA}>
                      <ShieldCheck className="mr-2 h-4 w-4" />
                      {twoFALoading ? t("settings.verifying") : t("settings.verifyAndEnable")}
                    </Button>
                    <Button variant="ghost" className="h-11 rounded-xl" onClick={() => setTwoFASetup(null)}>
                      {t("common.cancel")}
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="glass border-0 rounded-2xl">
            <CardHeader>
              <CardTitle>{t("settings.loginDevices")}</CardTitle>
            </CardHeader>
            <CardContent>
              {security?.devices && security.devices.length > 0 ? (
                <div className="space-y-3">
                  {security.devices.map((d) => (
                    <div key={d.id} className="flex items-center justify-between rounded-lg bg-accent/50 px-4 py-3">
                      <div className="flex items-center gap-3">
                        <Smartphone className="h-4 w-4 text-muted-foreground" />
                        <div>
                          <p className="text-sm">{d.ip}</p>
                          <p className="text-xs text-muted-foreground">{formatDateTime(d.last_login)}</p>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => api.delete(`/settings/devices/${d.id}`)}
                      >
                        <Trash2 className="h-3.5 w-3.5 text-destructive" />
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">{t("settings.noDevices")}</p>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {tab === "notifications" && (
        <Card className="glass border-0 rounded-2xl">
          <CardHeader>
            <CardTitle>{t("settings.notifPref")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {notifPref && (
              <>
                <p className="text-sm font-medium text-muted-foreground">{t("settings.emailNotif")}</p>
                {([
                  ["email_trade", "settings.tradeDone"],
                  ["email_deposit", "settings.depositConfirm"],
                  ["email_withdraw", "settings.withdrawUpdate"],
                  ["email_settlement", "settings.settlementReport"],
                ] as const).map(([key, label]) => (
                  <NotifToggle
                    key={key}
                    label={t(label)}
                    checked={notifPref[key]}
                    onChange={(v) => toggleNotif(key, v)}
                  />
                ))}

                <Separator className="opacity-10" />

                <p className="text-sm font-medium text-muted-foreground">{t("settings.pushNotif")}</p>
                {([
                  ["push_trade", "settings.tradeAlert"],
                  ["push_deposit", "settings.depositAlert"],
                  ["push_withdraw", "settings.withdrawAlert"],
                  ["push_premium_alert", "settings.premiumAlert"],
                ] as const).map(([key, label]) => (
                  <NotifToggle
                    key={key}
                    label={t(label)}
                    checked={notifPref[key]}
                    onChange={(v) => toggleNotif(key, v)}
                  />
                ))}
              </>
            )}
          </CardContent>
        </Card>
      )}

    </div>
  )
}

function NotifToggle({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm">{label}</span>
      <button
        onClick={() => onChange(!checked)}
        className={`relative h-6 w-11 rounded-full transition-colors ${checked ? "bg-primary" : "bg-accent"}`}
      >
        <span className={`absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white transition-transform ${checked ? "translate-x-5" : ""}`} />
      </button>
    </div>
  )
}
