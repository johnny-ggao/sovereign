"use client"

import { useState, useMemo } from "react"
import { usePremiumHistory } from "@/hooks/use-api"
import { usePremiumWS } from "@/hooks/use-premium-ws"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  TrendingUp,
  TrendingDown,
  Activity,
  CircleDot,
  Zap,
  Wifi,
  Radio,
  Shield,
} from "lucide-react"
import { formatCurrency, formatPct } from "@/lib/format"
import { useT } from "@/hooks/use-t"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts"
import { ChartContainer } from "@/components/ui/chart-container"
import type { PremiumTick } from "@/types/api"

// Mock 24h 成交量（后续可从 API 获取）
const MOCK_VOLUMES: Record<string, string> = {
  "BTC/KRW": "₩847.2B",
  "ETH/KRW": "₩312.5B",
  "SOL/KRW": "₩89.3B",
  "XRP/KRW": "₩156.7B",
}

// USDT/KRW 汇率（后续可从 WS 推送）
const USDT_KRW_RATE = 1512.5

export default function PremiumPage() {
  const [selectedPair, setSelectedPair] = useState("BTC/KRW")
  const { data: history } = usePremiumHistory(selectedPair, 200)
  const { ticks: wsTicks, isConnected } = usePremiumWS()
  const t = useT()

  // WS 实时数据按固定顺序排列
  const allTicks = useMemo(() => {
    const tickMap = new Map<string, PremiumTick>()
    for (const tick of wsTicks) {
      tickMap.set(tick.pair, tick)
    }
    const order = ["BTC/KRW", "ETH/KRW", "SOL/KRW", "XRP/KRW"]
    return order.map((p) => tickMap.get(p)).filter(Boolean) as PremiumTick[]
  }, [wsTicks])

  // 计算平均溢价
  const avgPremium = useMemo(() => {
    if (allTicks.length === 0) return 0
    const sum = allTicks.reduce((acc, t) => acc + parseFloat(t.premium_pct), 0)
    return sum / allTicks.length
  }, [allTicks])

  const selectedTick = allTicks.find((t) => t.pair === selectedPair)

  // 从 tick 数据中提取所有活跃的交易所名称
  const exchangeSources = useMemo(() => {
    const sources = new Set<string>()
    for (const tick of allTicks) {
      if (tick.source_kr) sources.add(tick.source_kr)
      if (tick.source_gl) sources.add(tick.source_gl)
    }
    return Array.from(sources).map((s) => s.charAt(0).toUpperCase() + s.slice(1)).join(" / ") || "—"
  }, [allTicks])

  return (
    <div className="space-y-6 glow-bg-green">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">
            {t("premium.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("premium.subtitle")}
          </p>
        </div>
        <Badge
          variant="outline"
          className={`${isConnected ? "border-success/30 text-success pulse-glow" : "border-destructive/30 text-destructive"}`}
        >
          <CircleDot className="mr-1 h-3 w-3" />
          {isConnected ? t("common.live") : t("common.disconnected")}
        </Badge>
      </div>

      {/* Key Metrics Cards */}
      <div className="grid gap-3 md:grid-cols-3">
        {/* Average Premium */}
        <div className="glass rounded-2xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider">
            {t("premium.avgPremium")}
          </p>
          <div className="mt-2 flex items-center gap-2">
            <span
              className={`text-3xl font-bold ${avgPremium >= 0 ? "text-success" : "text-destructive"}`}
            >
              {formatPct(avgPremium)}
            </span>
            {avgPremium >= 0 ? (
              <TrendingUp className="h-5 w-5 text-success" />
            ) : (
              <TrendingDown className="h-5 w-5 text-destructive" />
            )}
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            {t("premium.avgDesc")}
          </p>
        </div>

        {/* USDT/KRW Rate */}
        <div className="glass rounded-2xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider">
            USDT/KRW
          </p>
          <p className="mt-2 text-3xl font-bold">
            {formatCurrency(USDT_KRW_RATE, 2)}
          </p>
          <p className="mt-2 text-xs text-muted-foreground">KRW</p>
        </div>

        {/* Network Status */}
        <div className="glass rounded-2xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider">
            {t("premium.networkLatency")}
          </p>
          {(() => {
            const latencies: Record<string, number> = {}
            for (const tick of allTicks) {
              if (tick.latencies) {
                for (const [k, v] of Object.entries(tick.latencies)) {
                  if (v > 0) latencies[k] = v
                }
              }
            }
            const values = Object.values(latencies).filter((v) => v > 0)
            const avgMs = values.length > 0 ? Math.round(values.reduce((a, b) => a + b, 0) / values.length) : 0
            return (
              <>
                <div className="mt-2 flex items-center gap-2">
                  <Zap className={`h-5 w-5 ${avgMs > 0 && avgMs < 500 ? "text-success" : "text-muted-foreground"}`} />
                  <span className="text-3xl font-bold">{avgMs > 0 ? `${avgMs}ms` : "—"}</span>
                </div>
                <div className="mt-2 space-y-0.5">
                  {Object.entries(latencies).map(([name, ms]) => (
                    <div key={name} className="flex items-center justify-between text-[10px] text-muted-foreground">
                      <span>{name}</span>
                      <span className={ms < 200 ? "text-success" : ms < 1000 ? "text-chart-3" : "text-destructive"}>{ms}ms</span>
                    </div>
                  ))}
                </div>
              </>
            )
          })()}
        </div>
      </div>

      {/* System Status Bar */}
      <div className="flex items-center gap-4 rounded-xl bg-accent/30 px-4 py-2.5 text-xs text-muted-foreground">
        <div className="flex items-center gap-1.5">
          <Radio className="h-3 w-3 text-success animate-pulse" />
          <span>
            {t("premium.activeStreams")}: {allTicks.length}{" "}
            {t("premium.pairs")}
          </span>
        </div>
        <div className="hidden items-center gap-1.5 md:flex">
          <span>{exchangeSources}</span>
        </div>
        <div className="ml-auto flex items-center gap-1.5">
          <div
            className={`h-1.5 w-1.5 rounded-full ${isConnected ? "bg-success" : "bg-destructive"}`}
          />
          <span>
            {isConnected ? "LIVE DATA WEBSOCKET: CONNECTED" : "DISCONNECTED"}
          </span>
        </div>
      </div>

      {/* Real-Time Arbitrage Monitor Table */}
      <div className="glass rounded-2xl p-5">
        <h2 className="mb-4 font-semibold">{t("premium.realtimeMonitor")}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-accent/20 text-xs text-muted-foreground uppercase tracking-wider">
                <th className="pb-3 text-left font-medium">
                  {t("premium.asset")}
                </th>
                <th className="pb-3 text-right font-medium">
                  {t("premium.forward")}
                </th>
                <th className="pb-3 text-right font-medium">
                  {t("premium.reverse")}
                </th>
                <th className="hidden pb-3 text-right font-medium md:table-cell">
                  {t("premium.krwPrice")}
                </th>
                <th className="hidden pb-3 text-right font-medium md:table-cell">
                  {t("premium.globalPriceCol")}
                </th>
                <th className="hidden pb-3 text-right font-medium lg:table-cell">
                  {t("premium.volume24h")}
                </th>
              </tr>
            </thead>
            <tbody>
              {allTicks.map((tick) => {
                const fwd = parseFloat(tick.premium_pct)
                const rev = parseFloat(tick.reverse_premium_pct || "0")
                const krPrice = parseFloat(tick.korean_price)
                const glPrice = parseFloat(tick.global_price)
                const isSelected = tick.pair === selectedPair
                return (
                  <tr
                    key={tick.pair}
                    onClick={() => setSelectedPair(tick.pair)}
                    className={`cursor-pointer border-b border-accent/10 transition-colors ${isSelected ? "bg-primary/5" : "hover:bg-accent/20"}`}
                  >
                    <td className="py-3.5">
                      <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-accent/50 text-xs font-bold">
                          {tick.pair.split("/")[0].slice(0, 3)}
                        </div>
                        <div>
                          <p className="font-medium">{tick.pair}</p>
                          <p className="text-[10px] text-muted-foreground">
                            {tick.source_kr} / {tick.source_gl}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 text-right">
                      <div
                        className={`inline-flex items-center gap-1 font-bold ${fwd >= 0 ? "text-success" : "text-destructive"}`}
                      >
                        {fwd >= 0 ? (
                          <TrendingUp className="h-3.5 w-3.5" />
                        ) : (
                          <TrendingDown className="h-3.5 w-3.5" />
                        )}
                        {formatPct(fwd)}
                      </div>
                      <p className="text-[10px] text-muted-foreground">{t("premium.forwardDesc")}</p>
                    </td>
                    <td className="py-3.5 text-right">
                      <div
                        className={`inline-flex items-center gap-1 font-bold ${rev >= 0 ? "text-success" : "text-destructive"}`}
                      >
                        {rev >= 0 ? (
                          <TrendingUp className="h-3.5 w-3.5" />
                        ) : (
                          <TrendingDown className="h-3.5 w-3.5" />
                        )}
                        {formatPct(rev)}
                      </div>
                      <p className="text-[10px] text-muted-foreground">{t("premium.reverseDesc")}</p>
                    </td>
                    <td className="hidden py-3.5 text-right md:table-cell">
                      <span>₩{formatCurrency(krPrice, 0)}</span>
                    </td>
                    <td className="hidden py-3.5 text-right md:table-cell">
                      <span>₩{formatCurrency(glPrice, 0)}</span>
                    </td>
                    <td className="hidden py-3.5 text-right text-muted-foreground lg:table-cell">
                      {MOCK_VOLUMES[tick.pair] || "—"}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Premium History Chart */}
      <div className="glass rounded-2xl p-5">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">
            {t("premium.history")} — {selectedPair}
          </h2>
          <Tabs value={selectedPair} onValueChange={setSelectedPair}>
            <TabsList className="h-8 bg-accent/50">
              {["BTC", "ETH", "SOL", "XRP"].map((sym) => (
                <TabsTrigger key={sym} value={`${sym}/KRW`} className="h-6 px-3 text-xs">
                  {sym}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </div>
        <ChartContainer className="mt-4 h-[280px] md:h-[360px]">
          {history?.points && history.points.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={[...history.points].reverse()}>
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="rgba(66,71,84,0.15)"
                />
                <XAxis
                  dataKey="timestamp"
                  stroke="#424754"
                  fontSize={10}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(v) =>
                    new Date(v).toLocaleDateString("en", {
                      month: "short",
                      day: "numeric",
                    })
                  }
                />
                <YAxis
                  stroke="#424754"
                  fontSize={10}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(v) => `${v}%`}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#191f2f",
                    border: "none",
                    borderRadius: "0.75rem",
                    color: "#dce2f7",
                  }}
                  formatter={(value) => [
                    `${Number(value).toFixed(2)}%`,
                    "Premium",
                  ]}
                />
                <Line
                  type="monotone"
                  dataKey="premium_pct"
                  stroke="#3fe397"
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              {t("premium.noHistory")}
            </div>
          )}
        </ChartContainer>
      </div>

      {/* Spread Detail + Execution Zone */}
      <div className="grid gap-3 md:grid-cols-2">
        {/* Spread Detail */}
        {selectedTick && (
          <div className="glass rounded-2xl p-5">
            <h2 className="mb-4 font-semibold">{t("premium.spreadDetail")}</h2>
            <div className="space-y-3">
              <div className="flex items-center justify-between rounded-xl bg-accent/30 p-4">
                <div>
                  <p className="text-xs text-muted-foreground">
                    {t("premium.koreanPrice")}
                  </p>
                  <p className="text-lg font-bold">
                    ₩{formatCurrency(selectedTick.korean_price, 0)}
                  </p>
                </div>
                <Badge variant="outline" className="text-[10px]">
                  {selectedTick.source_kr}
                </Badge>
              </div>
              <div className="flex items-center justify-between rounded-xl bg-accent/30 p-4">
                <div>
                  <p className="text-xs text-muted-foreground">
                    {t("premium.globalPrice")}
                  </p>
                  <p className="text-lg font-bold">
                    ${formatCurrency(selectedTick.global_price, 0)}
                  </p>
                </div>
                <Badge variant="outline" className="text-[10px]">
                  {selectedTick.source_gl}
                </Badge>
              </div>
              <div className="flex items-center justify-between rounded-xl bg-accent/30 p-4">
                <div>
                  <p className="text-xs text-muted-foreground">
                    {t("premium.premium")}
                  </p>
                  <p
                    className={`text-lg font-bold ${parseFloat(selectedTick.premium_pct) >= 0 ? "text-success" : "text-destructive"}`}
                  >
                    {formatPct(selectedTick.premium_pct)}
                  </p>
                </div>
                <span className="text-xs text-muted-foreground">
                  ₩
                  {formatCurrency(
                    parseFloat(selectedTick.korean_price) -
                      parseFloat(selectedTick.global_price),
                    0,
                  )}
                </span>
              </div>
            </div>
          </div>
        )}

        {/* Safe Execution Zone */}
        <div className="glass rounded-2xl p-5">
          <h2 className="mb-4 font-semibold">{t("premium.executionZone")}</h2>
          <div className="flex flex-col items-center justify-center rounded-xl bg-success/5 border border-success/10 p-6 text-center">
            <Shield className="mb-3 h-10 w-10 text-success/60" />
            <p className="text-sm font-medium text-success">
              {t("premium.liquidityStatus")}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              {t("premium.liquidityDesc")}
            </p>
          </div>
          <div className="mt-4 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">
                {t("premium.avgSpread")}
              </span>
              <span className="font-medium text-success">
                {formatPct(avgPremium)}
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">
                {t("premium.activePairs")}
              </span>
              <span className="font-medium">{allTicks.length}</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">
                {t("premium.dataSource")}
              </span>
              <span className="font-medium">{exchangeSources}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
