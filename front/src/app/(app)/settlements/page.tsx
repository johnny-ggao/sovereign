"use client"

import { useMemo, useState } from "react"
import { useSettlements, useProfile } from "@/hooks/use-api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Skeleton } from "@/components/ui/skeleton"
import { FileText, DollarSign, TrendingUp, Calculator } from "lucide-react"
import { formatCurrency, formatPct, formatDate } from "@/lib/format"
import { useT } from "@/hooks/use-t"

export default function SettlementsPage() {
  const { data, isLoading } = useSettlements()
  const { data: profile } = useProfile()
  const t = useT()
  const [tab, setTab] = useState<"current" | "historical">("current")

  const currentType = profile?.investment_type ?? "arbitrage"
  const allSettlements = useMemo(() => data?.settlements ?? [], [data])
  const currentSettlements = useMemo(
    () => allSettlements.filter((s) => (s.product_type ?? "arbitrage") === currentType),
    [allSettlements, currentType]
  )
  const historicalSettlements = useMemo(
    () => allSettlements.filter((s) => (s.product_type ?? "arbitrage") !== currentType),
    [allSettlements, currentType]
  )
  const showTabs = currentSettlements.length > 0 && historicalSettlements.length > 0
  const visibleSettlements = !showTabs
    ? allSettlements
    : tab === "current"
    ? currentSettlements
    : historicalSettlements

  if (isLoading) return <Skeleton className="h-96" />

  const summary = data?.summary

  return (
    <div className="space-y-6 glow-bg">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">{t("settlement.title")}</h1>
        <p className="text-muted-foreground">{t("settlement.subtitle")}</p>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <DollarSign className="h-5 w-5 text-primary" />
            <div>
              <p className="text-xs text-muted-foreground">{t("settlement.grossReturn")}</p>
              <p className="text-lg font-bold">${formatCurrency(summary?.total_gross_return || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <Calculator className="h-5 w-5 text-warning" />
            <div>
              <p className="text-xs text-muted-foreground">{t("settlement.performanceFee")}</p>
              <p className="text-lg font-bold text-warning">${formatCurrency(summary?.total_performance_fee || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <TrendingUp className="h-5 w-5 text-success" />
            <div>
              <p className="text-xs text-muted-foreground">{t("settlement.netReturn")}</p>
              <p className="text-lg font-bold text-success">${formatCurrency(summary?.total_net_return || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <FileText className="h-5 w-5 text-chart-3" />
            <div>
              <p className="text-xs text-muted-foreground">{t("settlement.periods")}</p>
              <p className="text-lg font-bold">{summary?.period_count || 0}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Settlement Table */}
      {showTabs && (
        <div className="flex gap-2 border-b">
          <button
            type="button"
            className={
              tab === "current"
                ? "border-b-2 border-primary px-3 py-1.5 text-sm font-medium"
                : "px-3 py-1.5 text-sm text-muted-foreground"
            }
            onClick={() => setTab("current")}
          >
            {t("wallet.tabCurrent")}
          </button>
          <button
            type="button"
            className={
              tab === "historical"
                ? "border-b-2 border-primary px-3 py-1.5 text-sm font-medium"
                : "px-3 py-1.5 text-sm text-muted-foreground"
            }
            onClick={() => setTab("historical")}
          >
            {t("wallet.tabHistory")}
          </button>
        </div>
      )}
      <Card className="glass border-0 rounded-2xl">
        <CardHeader>
          <CardTitle>{t("settlement.monthlyReports")}</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>{t("settlement.period")}</TableHead>
                <TableHead>{t("settlement.trades")}</TableHead>
                <TableHead>{t("settlement.avgPremium")}</TableHead>
                <TableHead>{t("settlement.grossReturn")}</TableHead>
                <TableHead>{t("settlement.fee")}</TableHead>
                <TableHead>{t("settlement.netReturn")}</TableHead>
                <TableHead>{t("settlement.settled")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {visibleSettlements.length > 0 ? (
                visibleSettlements.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-medium">{s.period}</TableCell>
                    <TableCell>{s.trade_count}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="border-success/30 text-success">
                        {formatPct(s.avg_premium_pct)}
                      </Badge>
                    </TableCell>
                    <TableCell>${formatCurrency(s.gross_return)}</TableCell>
                    <TableCell className="text-warning">${formatCurrency(s.performance_fee)}</TableCell>
                    <TableCell className="font-semibold text-success">
                      ${formatCurrency(s.net_return)}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {formatDate(s.settled_at)}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={7} className="py-12 text-center text-muted-foreground">
                    {t("settlement.noReports")}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
