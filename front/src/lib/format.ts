export function formatCurrency(value: string | number, decimals = 2): string {
  const num = typeof value === "string" ? parseFloat(value) : value
  if (isNaN(num)) return "0.00"
  return new Intl.NumberFormat("en-US", {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(num)
}

export function formatPct(value: string | number, decimals = 2): string {
  const num = typeof value === "string" ? parseFloat(value) : value
  if (isNaN(num)) return "0.00%"
  return `${num >= 0 ? "+" : ""}${num.toFixed(decimals)}%`
}

export function formatCompactCurrency(value: string | number): string {
  const num = typeof value === "string" ? parseFloat(value) : value
  if (isNaN(num)) return "$0"
  if (num >= 1_000_000) return `$${(num / 1_000_000).toFixed(2)}M`
  if (num >= 1_000) return `$${(num / 1_000).toFixed(1)}K`
  return `$${num.toFixed(2)}`
}

export function shortenAddress(addr: string, chars = 6): string {
  if (addr.length <= chars * 2 + 2) return addr
  return `${addr.slice(0, chars + 2)}...${addr.slice(-chars)}`
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}
