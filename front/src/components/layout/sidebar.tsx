"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  LayoutDashboard,
  TrendingUp,
  Wallet,
  BarChart3,
  FileText,
  PieChart,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { Separator } from "@/components/ui/separator"
import { useT } from "@/hooks/use-t"

const navItems = [
  { href: "/dashboard", key: "nav.dashboard", icon: LayoutDashboard },
  { href: "/premium", key: "nav.premiumTicker", icon: TrendingUp },
  { href: "/wallet", key: "nav.wallet", icon: Wallet },
  { href: "/investments", key: "nav.investments", icon: PieChart },
  { href: "/trades", key: "nav.tradeLog", icon: BarChart3 },
  { href: "/settlements", key: "nav.reports", icon: FileText },
]

export function Sidebar() {
  const pathname = usePathname()
  const t = useT()

  return (
    <aside className="fixed left-0 top-0 z-40 hidden h-screen w-64 flex-col glass-light md:flex">
      <div className="flex h-14 items-center gap-3 px-6">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary">
          <span className="text-sm font-bold text-primary-foreground">S</span>
        </div>
        <span className="text-lg font-semibold tracking-tight text-sidebar-foreground">
          Sovereign
        </span>
      </div>

      <Separator className="opacity-10" />

      <nav className="flex-1 space-y-1 px-3 py-4">
        {navItems.map((item) => {
          const isActive = pathname.startsWith(item.href)
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "bg-sidebar-accent text-sidebar-primary"
                  : "text-muted-foreground hover:bg-sidebar-accent/50 hover:text-sidebar-foreground",
              )}
            >
              <item.icon className="h-4 w-4" />
              {t(item.key)}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
