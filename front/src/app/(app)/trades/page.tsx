"use client"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api-client"
import type { TradeList } from "@/types/api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Skeleton } from "@/components/ui/skeleton"
import { Download, BarChart3, TrendingUp, Target, Trophy } from "lucide-react"
import { formatCurrency, formatPct, formatDateTime } from "@/lib/format"
import { useT } from "@/hooks/use-t"

export default function TradesPage() {
  const [page, setPage] = useState(1)
  const t = useT()

  const { data, isLoading } = useQuery({
    queryKey: ["trades", page],
    queryFn: () => api.get<TradeList>(`/trades?page=${page}&per_page=20`),
  })

  async function handleExport() {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/trades/export`,
      { headers: { Authorization: `Bearer ${localStorage.getItem("access_token")}` } },
    )
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = "trades.csv"
    a.click()
    URL.revokeObjectURL(url)
  }

  if (isLoading) return <Skeleton className="h-96" />

  const summary = data?.summary

  return (
    <div className="space-y-6 glow-bg">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">{t("trade.title")}</h1>
          <p className="text-muted-foreground">{t("trade.subtitle")}</p>
        </div>
        <Button variant="outline" onClick={handleExport}>
          <Download className="mr-2 h-4 w-4" />{t("common.export")}
        </Button>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <BarChart3 className="h-5 w-5 text-primary" />
            <div>
              <p className="text-xs text-muted-foreground">{t("trade.totalTrades")}</p>
              <p className="text-lg font-bold">{summary?.total_trades || 0}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <TrendingUp className="h-5 w-5 text-success" />
            <div>
              <p className="text-xs text-muted-foreground">{t("trade.totalPnl")}</p>
              <p className={`text-lg font-bold ${parseFloat(summary?.total_pnl || "0") >= 0 ? "text-success" : "text-destructive"}`}>
                ${formatCurrency(summary?.total_pnl || "0")}
              </p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <Target className="h-5 w-5 text-chart-3" />
            <div>
              <p className="text-xs text-muted-foreground">{t("trade.avgPremium")}</p>
              <p className="text-lg font-bold">{formatPct(summary?.avg_premium_pct || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-3 p-4">
            <Trophy className="h-5 w-5 text-warning" />
            <div>
              <p className="text-xs text-muted-foreground">{t("trade.winRate")}</p>
              <p className="text-lg font-bold">{formatPct(summary?.win_rate || "0")}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Trade Table */}
      <Card className="glass border-0 rounded-2xl">
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>{t("trade.pair")}</TableHead>
                <TableHead>{t("trade.buy")}</TableHead>
                <TableHead>{t("trade.sell")}</TableHead>
                <TableHead>{t("trade.amount")}</TableHead>
                <TableHead>{t("premium.premium")}</TableHead>
                <TableHead>{t("trade.pnl")}</TableHead>
                <TableHead>{t("trade.time")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.trades && data.trades.length > 0 ? (
                data.trades.map((trade) => {
                  const pnl = parseFloat(trade.pnl)
                  return (
                    <TableRow key={trade.id}>
                      <TableCell className="font-medium">{trade.pair}</TableCell>
                      <TableCell>
                        <div>
                          <p className="text-sm">{trade.buy_exchange}</p>
                          <p className="text-xs text-muted-foreground">${formatCurrency(trade.buy_price)}</p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div>
                          <p className="text-sm">{trade.sell_exchange}</p>
                          <p className="text-xs text-muted-foreground">${formatCurrency(trade.sell_price)}</p>
                        </div>
                      </TableCell>
                      <TableCell>${formatCurrency(trade.amount)}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="border-success/30 text-success">
                          {formatPct(trade.premium_pct)}
                        </Badge>
                      </TableCell>
                      <TableCell className={pnl >= 0 ? "text-success" : "text-destructive"}>
                        ${formatCurrency(trade.pnl)}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {formatDateTime(trade.executed_at)}
                      </TableCell>
                    </TableRow>
                  )
                })
              ) : (
                <TableRow>
                  <TableCell colSpan={7} className="py-12 text-center text-muted-foreground">
                    {t("trade.noTrades")}
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
