"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { api } from "@/lib/api-client"
import { toast } from "sonner"
import { useT } from "@/hooks/use-t"
import {
  CheckCircle2,
  Circle,
  Loader2,
  Upload,
  Camera,
  FileText,
  ShieldCheck,
  Clock,
  Headphones,
} from "lucide-react"

type Step = "id" | "liveness" | "funds"

interface StepConfig {
  key: Step
  icon: React.ReactNode
  completed: boolean
  active: boolean
}

interface KYCVerificationProps {
  onComplete: () => void
}

export function KYCVerification({ onComplete }: KYCVerificationProps) {
  const t = useT()
  const [currentStep, setCurrentStep] = useState<Step>("id")
  const [idType, setIdType] = useState<string>("")
  const [idFile, setIdFile] = useState<File | null>(null)
  const [selfieCompleted, setSelfieCompleted] = useState(false)
  const [selfieLoading, setSelfieLoading] = useState(false)
  const [fundsFile, setFundsFile] = useState<File | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const completedSteps = {
    id: !!idType && !!idFile,
    liveness: selfieCompleted,
    funds: !!fundsFile,
  }

  const steps: StepConfig[] = [
    { key: "id", icon: <FileText className="h-4 w-4" />, completed: completedSteps.id, active: currentStep === "id" },
    { key: "liveness", icon: <Camera className="h-4 w-4" />, completed: completedSteps.liveness, active: currentStep === "liveness" },
    { key: "funds", icon: <Upload className="h-4 w-4" />, completed: completedSteps.funds, active: currentStep === "funds" },
  ]

  const stepLabels: Record<Step, string> = {
    id: t("kyc.governmentId"),
    liveness: t("kyc.livenessCheck"),
    funds: t("kyc.proofOfFunds"),
  }

  const stepDescriptions: Record<Step, string> = {
    id: t("kyc.governmentIdDesc"),
    liveness: t("kyc.livenessDesc"),
    funds: t("kyc.proofOfFundsDesc"),
  }

  const completedCount = Object.values(completedSteps).filter(Boolean).length

  function handleSimulateSelfie() {
    setSelfieLoading(true)
    setTimeout(() => {
      setSelfieLoading(false)
      setSelfieCompleted(true)
      toast.success(t("kyc.livenessSuccess"))
    }, 2500)
  }

  async function handleSubmit() {
    setSubmitting(true)
    try {
      await api.post("/settings/kyc/submit", {})
      toast.success(t("settings.kycSubmitted"))
      onComplete()
    } catch {
      toast.error(t("settings.kycSubmitFailed"))
    } finally {
      setSubmitting(false)
    }
  }

  const allCompleted = completedSteps.id && completedSteps.liveness && completedSteps.funds

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card className="glass border-0 rounded-2xl">
        <CardHeader>
          <CardTitle>{t("kyc.title")}</CardTitle>
          <CardDescription>{t("kyc.subtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          {/* Progress */}
          <div className="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
            <span>{t("kyc.step")} {completedCount + (allCompleted ? 0 : 1)} / 3</span>
            <Badge variant="outline" className="border-chart-3/30 text-chart-3">
              {allCompleted ? t("kyc.ready") : t("kyc.inProgress")}
            </Badge>
          </div>

          {/* Steps indicator */}
          <div className="space-y-3">
            {steps.map((step) => (
              <button
                key={step.key}
                onClick={() => setCurrentStep(step.key)}
                className={`flex w-full items-center gap-3 rounded-xl p-3 text-left transition-all ${
                  step.active ? "bg-accent/50 ring-1 ring-primary/30" : "hover:bg-accent/30"
                }`}
              >
                {step.completed ? (
                  <CheckCircle2 className="h-5 w-5 shrink-0 text-success" />
                ) : step.active ? (
                  <Loader2 className="h-5 w-5 shrink-0 animate-spin text-primary" />
                ) : (
                  <Circle className="h-5 w-5 shrink-0 text-muted-foreground/40" />
                )}
                <div className="flex-1">
                  <p className="text-sm font-medium">{stepLabels[step.key]}</p>
                  <p className="text-xs text-muted-foreground">{stepDescriptions[step.key]}</p>
                </div>
                {step.completed && (
                  <Badge variant="outline" className="border-success/30 text-success text-[10px]">
                    {t("kyc.completed")}
                  </Badge>
                )}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Step Content */}
      <Card className="glass border-0 rounded-2xl">
        <CardHeader>
          <CardTitle className="text-lg">{stepLabels[currentStep]}</CardTitle>
        </CardHeader>
        <CardContent>
          {currentStep === "id" && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>{t("kyc.idType")}</Label>
                <div className="flex gap-2">
                  {["passport", "national_id", "residence_permit"].map((type) => (
                    <Button
                      key={type}
                      variant={idType === type ? "default" : "outline"}
                      size="sm"
                      onClick={() => setIdType(type)}
                    >
                      {t(`kyc.${type}`)}
                    </Button>
                  ))}
                </div>
              </div>
              <div className="space-y-2">
                <Label>{t("kyc.uploadId")}</Label>
                <div
                  className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed p-8 transition-colors ${
                    idFile ? "border-success/30 bg-success/5" : "border-muted-foreground/20 hover:border-primary/30"
                  }`}
                >
                  {idFile ? (
                    <>
                      <CheckCircle2 className="mb-2 h-8 w-8 text-success" />
                      <p className="text-sm font-medium">{idFile.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {(idFile.size / 1024).toFixed(1)} KB
                      </p>
                    </>
                  ) : (
                    <>
                      <Upload className="mb-2 h-8 w-8 text-muted-foreground/40" />
                      <p className="text-sm text-muted-foreground">{t("kyc.dragOrClick")}</p>
                    </>
                  )}
                  <Input
                    type="file"
                    accept="image/*,.pdf"
                    className="absolute inset-0 cursor-pointer opacity-0"
                    onChange={(e) => {
                      const file = e.target.files?.[0]
                      if (file) setIdFile(file)
                    }}
                  />
                </div>
              </div>
              {completedSteps.id && (
                <Button className="w-full" onClick={() => setCurrentStep("liveness")}>
                  {t("kyc.next")}
                </Button>
              )}
            </div>
          )}

          {currentStep === "liveness" && (
            <div className="space-y-4">
              <div className="flex flex-col items-center rounded-xl bg-accent/30 p-8">
                {selfieCompleted ? (
                  <>
                    <CheckCircle2 className="mb-3 h-16 w-16 text-success" />
                    <p className="text-sm font-medium">{t("kyc.livenessSuccess")}</p>
                  </>
                ) : selfieLoading ? (
                  <>
                    <div className="relative mb-3">
                      <Camera className="h-16 w-16 text-primary" />
                      <div className="absolute inset-0 animate-ping rounded-full border-2 border-primary/30" />
                    </div>
                    <p className="text-sm text-muted-foreground">{t("kyc.scanning")}</p>
                  </>
                ) : (
                  <>
                    <Camera className="mb-3 h-16 w-16 text-muted-foreground/40" />
                    <p className="mb-4 text-center text-sm text-muted-foreground">{t("kyc.livenessInstruction")}</p>
                    <Button onClick={handleSimulateSelfie}>
                      <Camera className="mr-2 h-4 w-4" />
                      {t("kyc.startCamera")}
                    </Button>
                  </>
                )}
              </div>
              {selfieCompleted && (
                <Button className="w-full" onClick={() => setCurrentStep("funds")}>
                  {t("kyc.next")}
                </Button>
              )}
            </div>
          )}

          {currentStep === "funds" && (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">{t("kyc.fundsNote")}</p>
              <div
                className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed p-8 transition-colors ${
                  fundsFile ? "border-success/30 bg-success/5" : "border-muted-foreground/20 hover:border-primary/30"
                }`}
              >
                {fundsFile ? (
                  <>
                    <CheckCircle2 className="mb-2 h-8 w-8 text-success" />
                    <p className="text-sm font-medium">{fundsFile.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {(fundsFile.size / 1024).toFixed(1)} KB
                    </p>
                  </>
                ) : (
                  <>
                    <Upload className="mb-2 h-8 w-8 text-muted-foreground/40" />
                    <p className="text-sm text-muted-foreground">{t("kyc.dragOrClick")}</p>
                  </>
                )}
                <Input
                  type="file"
                  accept="image/*,.pdf"
                  className="absolute inset-0 cursor-pointer opacity-0"
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) setFundsFile(file)
                  }}
                />
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Submit */}
      {allCompleted && (
        <Button
          className="w-full h-12 rounded-xl text-base"
          disabled={submitting}
          onClick={handleSubmit}
        >
          <ShieldCheck className="mr-2 h-5 w-5" />
          {submitting ? t("common.loading") : t("kyc.submitVerification")}
        </Button>
      )}

      {/* Info Cards */}
      <div className="grid gap-3 md:grid-cols-3">
        <div className="glass rounded-xl p-4">
          <ShieldCheck className="mb-2 h-5 w-5 text-primary" />
          <p className="text-xs font-medium">{t("kyc.dataPrivacy")}</p>
          <p className="text-[10px] text-muted-foreground">{t("kyc.dataPrivacyDesc")}</p>
        </div>
        <div className="glass rounded-xl p-4">
          <Headphones className="mb-2 h-5 w-5 text-primary" />
          <p className="text-xs font-medium">{t("kyc.support")}</p>
          <p className="text-[10px] text-muted-foreground">{t("kyc.supportDesc")}</p>
        </div>
        <div className="glass rounded-xl p-4">
          <Clock className="mb-2 h-5 w-5 text-primary" />
          <p className="text-xs font-medium">{t("kyc.processingSpeed")}</p>
          <p className="text-[10px] text-muted-foreground">{t("kyc.processingSpeedDesc")}</p>
        </div>
      </div>
    </div>
  )
}
