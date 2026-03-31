"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useDashboardSummary, usePerformance, usePremiumLatest } from "@/hooks/use-api"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TrendingUp, TrendingDown, ArrowRight, Activity, Zap } from "lucide-react"
import { formatCurrency, formatPct } from "@/lib/format"
import { useT } from "@/hooks/use-t"
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts"
import { ChartContainer } from "@/components/ui/chart-container"

export default function DashboardPage() {
  const router = useRouter()
  const { data: summary, isLoading } = useDashboardSummary()
  const [period, setPeriod] = useState("1M")
  const { data: perf } = usePerformance(period)
  const { data: premiumLatest } = usePremiumLatest()
  const t = useT()

  if (isLoading) return <DashboardSkeleton />

  const portfolio = summary?.portfolio
  const ticks = premiumLatest?.ticks || []
  const btcTick = ticks.find((t) => t.pair === "BTC/KRW")
  const currentPremium = btcTick ? parseFloat(btcTick.premium_pct) : parseFloat(summary?.premium?.current_pct || "0")

  return (
    <div className="space-y-6 glow-bg">
      {/* Portfolio Hero Card */}
      <div className="glass rounded-2xl p-6">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">{t("dashboard.totalPortfolio")}</p>
        <h1 className="mt-2 text-4xl font-bold tracking-tight md:text-5xl">
          ${formatCurrency(portfolio?.total_value || "0")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">USDT</p>

        <div className="mt-6 grid grid-cols-2 gap-4">
          <div className="rounded-xl bg-accent/50 px-4 py-3">
            <p className="text-xs text-muted-foreground">{t("dashboard.available")}</p>
            <p className="mt-1 text-lg font-semibold">${formatCurrency(portfolio?.available || "0")}</p>
          </div>
          <div className="rounded-xl bg-accent/50 px-4 py-3">
            <p className="text-xs text-muted-foreground">{t("dashboard.inOperation")}</p>
            <p className="mt-1 text-lg font-semibold">${formatCurrency(portfolio?.in_operation || "0")}</p>
          </div>
        </div>

        <Button className="mt-5 h-12 w-full rounded-xl text-base font-semibold" onClick={() => router.push("/investments")}>
          {t("investment.invest")}
          <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-2 gap-3">
        {/* Kimchi Premium */}
        <div className="glass rounded-2xl p-4">
          <div className="flex items-center gap-2">
            {currentPremium >= 0 ? (
              <TrendingUp className="h-4 w-4 text-success" />
            ) : (
              <TrendingDown className="h-4 w-4 text-destructive" />
            )}
            <span className="text-xs text-muted-foreground">{t("dashboard.kimchiPremium")}</span>
          </div>
          <p className={`mt-2 text-2xl font-bold ${currentPremium >= 0 ? "text-success" : "text-destructive"}`}>
            {formatPct(currentPremium)}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">BTC/KRW</p>
        </div>

        {/* Active Status */}
        <div className="glass rounded-2xl p-4">
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4 text-primary" />
            <span className="text-xs text-muted-foreground">Active Nodes</span>
          </div>
          <p className="mt-2 text-2xl font-bold">12</p>
          <p className="mt-1 text-xs text-success">Online</p>
        </div>
      </div>

      {/* Performance Chart */}
      <div className="glass rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">{t("dashboard.performance")}</h2>
          <Tabs value={period} onValueChange={setPeriod}>
            <TabsList className="h-8 bg-accent/50">
              {["1W", "1M", "3M", "6M", "1Y", "ALL"].map((p) => (
                <TabsTrigger key={p} value={p} className="h-6 px-2 text-xs">{p}</TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </div>
        <ChartContainer className="mt-4 h-[200px] md:h-[280px]">
          {perf?.chart && perf.chart.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={perf.chart}>
                <defs>
                  <linearGradient id="fillValue" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#adc6ff" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#adc6ff" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" stroke="#424754" fontSize={10} tickLine={false} axisLine={false} />
                <YAxis stroke="#424754" fontSize={10} tickLine={false} axisLine={false} />
                <Tooltip contentStyle={{ backgroundColor: "#191f2f", border: "none", borderRadius: "0.75rem", color: "#dce2f7" }} />
                <Area type="monotone" dataKey="value" stroke="#adc6ff" fill="url(#fillValue)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">{t("dashboard.noPerformance")}</div>
          )}
        </ChartContainer>
      </div>

      {/* Live Arbitrage Stream */}
      <div className="glass rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">{t("dashboard.livePremium")}</h2>
          <Badge variant="outline" className="border-success/30 text-success pulse-glow">
            <Activity className="mr-1 h-3 w-3" />SYNCED
          </Badge>
        </div>
        <div className="mt-4 space-y-3">
          {ticks.length > 0 ? ticks.map((tick) => {
            const pct = parseFloat(tick.premium_pct)
            return (
              <div key={tick.pair} className="flex items-center justify-between rounded-xl bg-accent/30 px-4 py-3">
                <div className="flex items-center gap-3">
                  <Activity className="h-4 w-4 text-primary" />
                  <div>
                    <p className="text-sm font-medium">{tick.pair}</p>
                    <p className="text-xs text-muted-foreground">{tick.source_kr} → {tick.source_gl}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className={`text-sm font-bold ${pct >= 0 ? "text-success" : "text-destructive"}`}>{formatPct(pct)}</p>
                  <p className="text-xs text-muted-foreground">₩{formatCurrency(tick.korean_price, 0)}</p>
                </div>
              </div>
            )
          }) : (
            <div className="py-8 text-center text-sm text-muted-foreground">
              <Activity className="mx-auto mb-2 h-5 w-5" />{t("premium.waiting")}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="glass rounded-2xl p-6"><Skeleton className="h-12 w-48" /><Skeleton className="mt-4 h-20" /><Skeleton className="mt-4 h-12 w-full" /></div>
      <div className="grid grid-cols-2 gap-3"><Skeleton className="h-28 rounded-2xl" /><Skeleton className="h-28 rounded-2xl" /></div>
      <Skeleton className="h-[280px] rounded-2xl" />
    </div>
  )
}
