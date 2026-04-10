"use client"

import { useState } from "react"
import { useInvestments, useCreateInvestment, useRedeemInvestment } from "@/hooks/use-api"
import { ApiError } from "@/lib/api-client"
import { toast } from "sonner"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { PieChart, Plus, StopCircle, TrendingUp } from "lucide-react"
import { formatCurrency, formatPct, formatDate } from "@/lib/format"
import { useT } from "@/hooks/use-t"

export default function InvestmentsPage() {
  const { data, isLoading } = useInvestments()
  const createInv = useCreateInvestment()
  const redeemInv = useRedeemInvestment()
  const [amount, setAmount] = useState("")
  const [dialogOpen, setDialogOpen] = useState(false)
  const [confirmInvestOpen, setConfirmInvestOpen] = useState(false)
  const [confirmRedeemId, setConfirmRedeemId] = useState<string | null>(null)
  const t = useT()

  const [createError, setCreateError] = useState("")

  function handleCreate() {
    setConfirmInvestOpen(true)
  }

  async function confirmCreate() {
    setCreateError("")
    try {
      await createInv.mutateAsync({ amount })
      setAmount("")
      setConfirmInvestOpen(false)
      setDialogOpen(false)
      toast.success(t("investment.createSuccess"))
    } catch (err) {
      setConfirmInvestOpen(false)
      if (err instanceof ApiError) {
        if (err.code === "INVESTMENT_MIN_AMOUNT") {
          setCreateError(t("investment.minAmountError"))
        } else if (err.code === "INSUFFICIENT_FUNDS") {
          setCreateError(t("investment.insufficientFunds"))
        } else {
          setCreateError(err.message)
        }
      }
    }
  }

  async function confirmRedeem() {
    if (!confirmRedeemId) return
    try {
      await redeemInv.mutateAsync(confirmRedeemId)
      setConfirmRedeemId(null)
    } catch {
      setConfirmRedeemId(null)
    }
  }

  if (isLoading) return <Skeleton className="h-96" />

  const summary = data?.summary
  const investConfirmContent = t("investment.confirmInvestContent").replace("{amount}", amount || "0")

  return (
    <div className="space-y-6 glow-bg">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight md:text-3xl">{t("investment.title")}</h1>
          <p className="text-muted-foreground">{t("investment.subtitle")}</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger render={<Button><Plus className="mr-2 h-4 w-4" />{t("investment.newInvestment")}</Button>} />
          <DialogContent className="bg-card">
            <DialogHeader>
              <DialogTitle>{t("investment.createInvestment")}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 pt-4">
              {createError && (
                <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">{createError}</div>
              )}
              <div className="space-y-2">
                <Label>{t("investment.amount")}</Label>
                <Input
                  type="number"
                  placeholder={t("investment.minAmount")}
                  value={amount}
                  onChange={(e) => { setAmount(e.target.value); setCreateError("") }}
                  className="bg-input"
                />
                <p className="text-xs text-muted-foreground">{t("investment.minAmount")}</p>
              </div>
              <Button
                className="w-full"
                onClick={handleCreate}
                disabled={createInv.isPending || !amount}
              >
                {createInv.isPending ? t("investment.creating") : t("investment.invest")}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
        <Dialog open={confirmInvestOpen} onOpenChange={setConfirmInvestOpen}>
          <DialogContent className="bg-card">
            <DialogHeader>
              <DialogTitle>{t("investment.confirmInvestTitle")}</DialogTitle>
              <DialogDescription>{investConfirmContent}</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose render={<Button variant="outline" />}>
                {t("investment.cancel")}
              </DialogClose>
              <Button onClick={confirmCreate} disabled={createInv.isPending}>
                {createInv.isPending ? t("investment.creating") : t("investment.confirmInvest")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        <Dialog open={!!confirmRedeemId} onOpenChange={(open) => { if (!open) setConfirmRedeemId(null) }}>
          <DialogContent className="bg-card">
            <DialogHeader>
              <DialogTitle>{t("investment.confirmRedeemTitle")}</DialogTitle>
              <DialogDescription>{t("investment.confirmRedeemContent")}</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose render={<Button variant="outline" />}>
                {t("investment.cancel")}
              </DialogClose>
              <Button onClick={confirmRedeem} disabled={redeemInv.isPending}>
                {t("investment.confirmRedeem")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 gap-3 md:grid-cols-3 md:gap-4">
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
              <PieChart className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t("investment.totalInvested")}</p>
              <p className="text-xl font-bold">${formatCurrency(summary?.total_invested || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-success/10">
              <TrendingUp className="h-5 w-5 text-success" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t("investment.totalReturn")}</p>
              <p className="text-xl font-bold text-success">${formatCurrency(summary?.total_return || "0")}</p>
            </div>
          </CardContent>
        </Card>
        <Card className="glass border-0 rounded-2xl">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-chart-3/10">
              <span className="text-lg font-bold text-chart-3">{summary?.active_count || 0}</span>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t("investment.activeInvestments")}</p>
              <p className="text-xl font-bold">{summary?.active_count || 0}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Investment List */}
      <div className="space-y-4">
        {data?.investments && data.investments.length > 0 ? (
          data.investments.map((inv) => (
            <Card key={inv.id} className="glass border-0 rounded-2xl">
              <CardContent className="flex items-center justify-between p-6">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <p className="font-semibold">${formatCurrency(inv.amount)} {inv.currency}</p>
                    <Badge
                      variant="outline"
                      className={
                        inv.status === "active"
                          ? "border-success/30 text-success"
                          : inv.status === "stopping"
                          ? "border-warning text-warning"
                          : "border-muted-foreground/30"
                      }
                    >
                      {t(`investment.status_${inv.status}`)}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {t("investment.started")} {formatDate(inv.start_date)}
                    {inv.end_date && ` · ${t("investment.ended")} ${formatDate(inv.end_date)}`}
                  </p>
                </div>
                <div className="flex items-center gap-6">
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">{t("investment.netReturn")}</p>
                    <p className={`font-semibold ${parseFloat(inv.net_return) >= 0 ? "text-success" : "text-destructive"}`}>
                      ${formatCurrency(inv.net_return)} ({formatPct(inv.return_pct)})
                    </p>
                  </div>
                  {inv.status === "active" && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setConfirmRedeemId(inv.id)}
                      disabled={redeemInv.isPending}
                    >
                      <StopCircle className="mr-1 h-3 w-3" />
                      {t("investment.redeem")}
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          ))
        ) : (
          <Card className="glass border-0 rounded-2xl">
            <CardContent className="py-12 text-center text-muted-foreground">
              {t("investment.noInvestments")}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
