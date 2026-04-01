"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useWallets, useTransactions, useDepositAddress, useWhitelistAddresses, useAddWhitelistAddress, useRemoveWhitelistAddress, useWithdraw, useSecurityOverview } from "@/hooks/use-api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { ArrowDownLeft, ArrowUpRight, Copy, Check, Wallet, ChevronRight, Plus, Trash2, Clock, ShieldAlert } from "lucide-react"
import { QRCodeSVG } from "qrcode.react"
import { formatCurrency, formatDateTime, shortenAddress } from "@/lib/format"
import { useT } from "@/hooks/use-t"
import { toast } from "sonner"

const NETWORKS = [
  { value: "BEP20", label: "BEP-20 (BSC)" },
  { value: "TRC20", label: "TRC-20 (TRON)" },
]

export default function WalletPage() {
  const { data: wallets, isLoading } = useWallets()
  const { data: transactions } = useTransactions(undefined, 1)
  const { data: addresses } = useWhitelistAddresses()
  const addAddress = useAddWhitelistAddress()
  const removeAddress = useRemoveWhitelistAddress()
  const depositAddr = useDepositAddress()
  const withdraw = useWithdraw()
  const { data: security } = useSecurityOverview()
  const router = useRouter()
  const t = useT()
  const twoFAEnabled = security?.two_fa_enabled ?? false
  const [copied, setCopied] = useState("")
  const [showAddForm, setShowAddForm] = useState(false)
  const [newAddr, setNewAddr] = useState({ currency: "USDT", network: "BEP20", address: "", label: "" })
  const [depositResult, setDepositResult] = useState<{ currency: string; network: string; address: string } | null>(null)
  const [tab, setTab] = useState<"activity" | "deposit" | "withdraw" | "addresses">("activity")
  const [withdrawForm, setWithdrawForm] = useState({ currency: "USDT", network: "BEP20", address: "", amount: "", two_fa_code: "" })
  const [depositNetwork, setDepositNetwork] = useState("BEP20")

  // 切换到充值 tab 或切换网络时自动获取/生成地址
  useEffect(() => {
    if (tab === "deposit") {
      depositAddr.mutate({ currency: "USDT", network: depositNetwork }, {
        onSuccess: (result) => setDepositResult(result),
      })
    }
  }, [tab, depositNetwork]) // eslint-disable-line react-hooks/exhaustive-deps

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text)
    setCopied(text)
    setTimeout(() => setCopied(""), 2000)
  }

  if (isLoading) return <Skeleton className="h-96 rounded-2xl" />

  const primaryWallet = wallets?.wallets.find((w) => w.currency === "USDT")
  const totalValue = wallets?.total_usdt || "0"

  return (
    <div className="space-y-6 glow-bg">
      {/* Balance Hero */}
      <div className="glass rounded-2xl p-6 text-center">
        <p className="text-xs uppercase tracking-wider text-muted-foreground">{t("wallet.title")}</p>
        <h1 className="mt-3 text-4xl font-bold tracking-tight">${formatCurrency(totalValue)}</h1>
        <p className="mt-1 text-sm text-muted-foreground">USDT</p>

        <div className="mt-5 grid grid-cols-3 gap-2">
          <div className="rounded-xl bg-accent/50 px-3 py-2.5">
            <p className="text-[10px] text-muted-foreground">{t("wallet.available")}</p>
            <p className="mt-0.5 text-sm font-semibold">{formatCurrency(primaryWallet?.available || "0")}</p>
          </div>
          <div className="rounded-xl bg-accent/50 px-3 py-2.5">
            <p className="text-[10px] text-muted-foreground">{t("wallet.inOp")}</p>
            <p className="mt-0.5 text-sm font-semibold">{formatCurrency(primaryWallet?.in_operation || "0")}</p>
          </div>
          <div className="rounded-xl bg-accent/50 px-3 py-2.5">
            <p className="text-[10px] text-muted-foreground">{t("wallet.frozen")}</p>
            <p className="mt-0.5 text-sm font-semibold">{formatCurrency(primaryWallet?.frozen || "0")}</p>
          </div>
        </div>

        <div className="mt-5 grid grid-cols-2 gap-3">
          <Button className="h-12 rounded-xl font-semibold" onClick={() => setTab("deposit")}>
            <ArrowDownLeft className="mr-2 h-4 w-4" />{t("wallet.deposit")}
          </Button>
          <Button variant="outline" className="h-12 rounded-xl font-semibold border-accent" onClick={() => setTab("withdraw")}>
            <ArrowUpRight className="mr-2 h-4 w-4" />{t("wallet.withdraw")}
          </Button>
        </div>
      </div>

      {/* Tab Selector */}
      <div className="flex gap-2">
        {(["activity", "deposit", "withdraw", "addresses"] as const).map((t2) => (
          <button key={t2} onClick={() => setTab(t2)} className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${tab === t2 ? "bg-primary text-primary-foreground" : "glass text-muted-foreground"}`}>
            {t2 === "activity" ? t("wallet.transactions") : t2 === "deposit" ? t("wallet.deposit") : t2 === "withdraw" ? t("wallet.withdraw") : t("wallet.whitelist")}
          </button>
        ))}
      </div>

      {/* Activity */}
      {tab === "activity" && (
        <div className="glass rounded-2xl p-4">
          {transactions && Array.isArray(transactions) && transactions.length > 0 ? (
            <div className="space-y-1">
              {transactions.map((tx) => (
                <div key={tx.id} className="flex items-center justify-between rounded-xl px-3 py-3 hover:bg-accent/30">
                  <div className="flex items-center gap-3">
                    <div className={`flex h-9 w-9 items-center justify-center rounded-full ${tx.type === "deposit" ? "bg-success/10" : "bg-destructive/10"}`}>
                      {tx.type === "deposit" ? <ArrowDownLeft className="h-4 w-4 text-success" /> : <ArrowUpRight className="h-4 w-4 text-destructive" />}
                    </div>
                    <div>
                      <p className="text-sm font-medium capitalize">{tx.type}</p>
                      <p className="text-xs text-muted-foreground">{tx.network} · {formatDateTime(tx.created_at)}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={`text-sm font-semibold ${tx.type === "deposit" ? "text-success" : "text-destructive"}`}>
                      {tx.type === "deposit" ? "+" : "-"}{formatCurrency(tx.amount)} USDT
                    </p>
                    <Badge variant="outline" className={`text-[10px] ${tx.status === "confirmed" ? "border-success/30 text-success" : "border-muted-foreground/30"}`}>
                      {tx.status}
                    </Badge>
                  </div>
                </div>
              ))}
              <button className="mt-2 flex w-full items-center justify-center gap-1 py-2 text-sm text-primary">
                View All <ChevronRight className="h-3.5 w-3.5" />
              </button>
            </div>
          ) : (
            <p className="py-8 text-center text-sm text-muted-foreground">{t("wallet.noTransactions")}</p>
          )}
        </div>
      )}

      {/* Deposit */}
      {tab === "deposit" && (
        <div className="glass rounded-2xl p-5">
          <h3 className="mb-4 font-semibold">{t("wallet.depositAddress")}</h3>
          <div className="grid grid-cols-2 gap-3">
            {NETWORKS.map((net) => (
              <button
                key={net.value}
                onClick={() => { setDepositNetwork(net.value); setDepositResult(null) }}
                className={`rounded-xl px-4 py-3 text-sm font-medium transition-colors ${depositNetwork === net.value ? "bg-primary text-primary-foreground" : "bg-accent/50 hover:bg-accent"}`}
              >
                {net.label}
              </button>
            ))}
          </div>
          <div className="mt-4 rounded-xl bg-accent/50 p-4">
            {depositAddr.isPending ? (
              <div className="flex items-center justify-center py-8">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              </div>
            ) : depositResult ? (
              <>
                <p className="mb-3 text-center text-xs text-muted-foreground">USDT ({NETWORKS.find((n) => n.value === depositResult.network)?.label || depositResult.network})</p>
                <div className="flex flex-col items-center gap-3">
                  <div className="rounded-xl bg-white p-3">
                    <QRCodeSVG value={depositResult.address} size={180} />
                  </div>
                  <div className="flex w-full items-center gap-2">
                    <code className="flex-1 break-all rounded-lg bg-input px-3 py-2 font-mono text-xs">{depositResult.address}</code>
                    <button onClick={() => copyToClipboard(depositResult.address)} className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-accent">
                      {copied === depositResult.address ? <Check className="h-4 w-4 text-success" /> : <Copy className="h-4 w-4" />}
                    </button>
                  </div>
                </div>
              </>
            ) : null}
          </div>
        </div>
      )}

      {/* Withdraw */}
      {tab === "withdraw" && (
        <div className="glass rounded-2xl p-5">
          <h3 className="mb-4 font-semibold">{t("wallet.withdraw")} USDT</h3>
          <div className="space-y-3">
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">{t("wallet.selectNetwork")}</Label>
              <select value={withdrawForm.network} onChange={(e) => setWithdrawForm((p) => ({ ...p, network: e.target.value, address: "" }))} className="h-10 w-full rounded-lg border-0 bg-input px-3 text-sm">
                {NETWORKS.map((net) => (
                  <option key={net.value} value={net.value}>{net.label}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">{t("wallet.withdrawAddress")}</Label>
              {addresses && addresses.length > 0 ? (
                <select value={withdrawForm.address} onChange={(e) => setWithdrawForm((p) => ({ ...p, address: e.target.value }))} className="h-10 w-full rounded-lg border-0 bg-input px-3 text-sm font-mono">
                  <option value="">{t("wallet.selectAddress")}</option>
                  {addresses.filter((a) => a.currency === "USDT" && a.network === withdrawForm.network).map((a) => (
                    <option key={a.id} value={a.address}>{a.label || shortenAddress(a.address)}</option>
                  ))}
                </select>
              ) : (
                <p className="rounded-lg bg-warning/10 px-3 py-2 text-xs text-warning">{t("wallet.noWhitelistWarn")}</p>
              )}
            </div>
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">{t("wallet.withdrawAmount")}</Label>
              <Input placeholder="0.00" value={withdrawForm.amount} onChange={(e) => setWithdrawForm((p) => ({ ...p, amount: e.target.value }))} className="h-10 rounded-lg border-0 bg-input" />
            </div>
            {twoFAEnabled ? (
              <>
                <div className="space-y-1">
                  <Label className="text-xs text-muted-foreground">{t("wallet.withdraw2fa")}</Label>
                  <Input placeholder="000000" maxLength={6} value={withdrawForm.two_fa_code} onChange={(e) => setWithdrawForm((p) => ({ ...p, two_fa_code: e.target.value }))} className="h-10 rounded-lg border-0 bg-input text-center tracking-widest" />
                </div>
                <Button className="h-12 w-full rounded-xl font-semibold" disabled={withdraw.isPending || !withdrawForm.address || !withdrawForm.amount}
                  onClick={() => withdraw.mutate(withdrawForm, { onSuccess: () => { toast.success(t("wallet.withdrawSuccess")); setWithdrawForm({ currency: "USDT", network: "BEP20", address: "", amount: "", two_fa_code: "" }); setTab("activity") } })}>
                  {withdraw.isPending ? t("wallet.withdrawing") : t("wallet.withdrawBtn")}
                </Button>
              </>
            ) : (
              <div className="rounded-xl bg-destructive/10 p-4 text-center">
                <ShieldAlert className="mx-auto mb-2 h-8 w-8 text-destructive" />
                <p className="text-sm font-medium">{t("wallet.require2fa")}</p>
                <p className="mt-1 text-xs text-muted-foreground">{t("wallet.require2faDesc")}</p>
                <Button className="mt-3" variant="outline" onClick={() => router.push("/settings?tab=security")}>
                  {t("wallet.setup2fa")}
                </Button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Whitelist */}
      {tab === "addresses" && (
        <div className="space-y-4">
          <div className="glass rounded-2xl p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">{t("wallet.whitelistAddresses")}</h3>
              <button onClick={() => setShowAddForm(!showAddForm)} className="flex items-center gap-1 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground">
                <Plus className="h-3.5 w-3.5" />{t("wallet.addAddress")}
              </button>
            </div>

            {showAddForm && (
              <div className="mb-4 rounded-xl bg-accent/30 p-4 space-y-3">
                <div className="space-y-1">
                  <Label className="text-xs text-muted-foreground">{t("wallet.selectNetwork")}</Label>
                  <select value={newAddr.network} onChange={(e) => setNewAddr((p) => ({ ...p, network: e.target.value }))} className="h-10 w-full rounded-lg border-0 bg-input px-3 text-sm">
                    {NETWORKS.map((net) => (
                      <option key={net.value} value={net.value}>{net.label}</option>
                    ))}
                  </select>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs text-muted-foreground">{t("wallet.addressLabel")}</Label>
                  <Input placeholder={t("wallet.labelPlaceholder")} value={newAddr.label} onChange={(e) => setNewAddr((p) => ({ ...p, label: e.target.value }))} className="h-10 rounded-lg border-0 bg-input" />
                </div>
                <div className="space-y-1">
                  <Label className="text-xs text-muted-foreground">{t("wallet.addressPlaceholder")}</Label>
                  <Input placeholder={newAddr.network === "TRC20" ? "T..." : "0x..."} value={newAddr.address} onChange={(e) => setNewAddr((p) => ({ ...p, address: e.target.value }))} className="h-10 rounded-lg border-0 bg-input font-mono text-xs" />
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock className="h-3.5 w-3.5" />{t("wallet.cooldownNote")}
                </div>
                <div className="flex gap-2">
                  <Button size="sm" className="rounded-lg" disabled={addAddress.isPending || !newAddr.address} onClick={() => { addAddress.mutate(newAddr, { onSuccess: () => { setNewAddr({ currency: "USDT", network: "BEP20", address: "", label: "" }); setShowAddForm(false) } }) }}>
                    {addAddress.isPending ? t("wallet.adding") : t("wallet.addAddress")}
                  </Button>
                  <Button size="sm" variant="ghost" className="rounded-lg" onClick={() => setShowAddForm(false)}>
                    {t("common.cancel")}
                  </Button>
                </div>
              </div>
            )}

            {addresses && addresses.length > 0 ? (
              <div className="space-y-2">
                {addresses.map((addr) => (
                  <div key={addr.id} className="flex items-center justify-between rounded-xl bg-accent/30 px-4 py-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium truncate">{addr.label || "Unnamed"}</p>
                        <Badge variant="outline" className="text-[10px] shrink-0">{NETWORKS.find((n) => n.value === addr.network)?.label || addr.network}</Badge>
                      </div>
                      <p className="mt-1 font-mono text-xs text-muted-foreground">{shortenAddress(addr.address)}</p>
                    </div>
                    <div className="flex items-center gap-2 ml-3">
                      <button onClick={() => removeAddress.mutate(addr.id)} className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-colors" title={t("common.delete")}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="py-8 text-center text-sm text-muted-foreground">{t("wallet.noAddresses")}</p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
