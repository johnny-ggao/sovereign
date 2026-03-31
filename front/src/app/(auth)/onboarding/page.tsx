"use client"

import { useRouter } from "next/navigation"
import { useAuthStore } from "@/stores/auth-store"
import { useT } from "@/hooks/use-t"
import { Button } from "@/components/ui/button"
import { Fingerprint, Wallet, Rocket, ArrowRight, X } from "lucide-react"

const steps = [
  {
    num: "01",
    icon: Fingerprint,
    titleKey: "onboarding.step1Title",
    descKey: "onboarding.step1Desc",
    color: "text-primary",
    bg: "bg-primary/10",
  },
  {
    num: "02",
    icon: Wallet,
    titleKey: "onboarding.step2Title",
    descKey: "onboarding.step2Desc",
    color: "text-success",
    bg: "bg-success/10",
  },
  {
    num: "03",
    icon: Rocket,
    titleKey: "onboarding.step3Title",
    descKey: "onboarding.step3Desc",
    color: "text-chart-3",
    bg: "bg-chart-3/10",
  },
]

export default function OnboardingPage() {
  const router = useRouter()
  const { user } = useAuthStore()
  const t = useT()

  const welcomeText = t("onboarding.welcome").replace("{name}", user?.full_name || "User")

  return (
    <div className="relative flex min-h-screen flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-4">
        <button
          onClick={() => router.push("/dashboard")}
          className="flex h-8 w-8 items-center justify-center rounded-full text-muted-foreground hover:bg-accent"
        >
          <X className="h-5 w-5" />
        </button>
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
            <span className="text-xs font-bold text-primary-foreground">S</span>
          </div>
          <span className="text-sm font-semibold">SOVEREIGN</span>
        </div>
        <div className="w-8" />
      </div>

      {/* Content */}
      <div className="flex flex-1 flex-col px-6 py-8">
        {/* Welcome */}
        <div className="mb-10">
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">
            {welcomeText}
          </h1>
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground md:text-base">
            {t("onboarding.welcomeSub")}
          </p>
        </div>

        {/* Steps */}
        <div className="space-y-6">
          {steps.map((step, idx) => (
            <div key={step.num} className="flex gap-4">
              {/* Step indicator */}
              <div className="flex flex-col items-center">
                <div className={`flex h-12 w-12 items-center justify-center rounded-xl ${step.bg}`}>
                  <step.icon className={`h-6 w-6 ${step.color}`} />
                </div>
                {idx < steps.length - 1 && (
                  <div className="mt-2 h-full w-px bg-border/50" />
                )}
              </div>

              {/* Step content */}
              <div className="flex-1 pb-6">
                <div className="mb-1 flex items-center gap-2">
                  <span className="text-xs font-medium text-muted-foreground">
                    STEP {step.num}
                  </span>
                </div>
                <h3 className="text-base font-semibold">{t(step.titleKey)}</h3>
                <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                  {t(step.descKey)}
                </p>
              </div>
            </div>
          ))}
        </div>

        {/* Spacer */}
        <div className="flex-1" />

        {/* Actions */}
        <div className="space-y-3 pb-8 pt-6">
          <Button
            className="w-full"
            size="lg"
            onClick={() => router.push("/settings?tab=kyc")}
          >
            {t("onboarding.completeKyc")}
            <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            className="w-full text-muted-foreground"
            onClick={() => router.push("/dashboard")}
          >
            {t("onboarding.maybeLater")}
          </Button>
        </div>
      </div>
    </div>
  )
}
