"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { LayoutDashboard, TrendingUp, Wallet, PieChart, FileText } from "lucide-react"
import { cn } from "@/lib/utils"
import { useT } from "@/hooks/use-t"

const navItems = [
  { href: "/dashboard", key: "nav.home", icon: LayoutDashboard },
  { href: "/investments", key: "nav.invest", icon: PieChart },
  { href: "/wallet", key: "nav.wallet", icon: Wallet },
  { href: "/premium", key: "nav.market", icon: TrendingUp },
  { href: "/trades", key: "nav.trades", icon: FileText },
]

export function BottomNav() {
  const pathname = usePathname()
  const t = useT()

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-border/30 glass-light md:hidden">
      <div className="flex items-center justify-around py-2 pb-[calc(0.5rem+env(safe-area-inset-bottom))]">
        {navItems.map((item) => {
          const isActive = pathname.startsWith(item.href)
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center gap-0.5 px-3 py-1 text-[10px] transition-colors",
                isActive ? "text-primary" : "text-muted-foreground",
              )}
            >
              <item.icon className={cn("h-5 w-5", isActive && "text-primary")} />
              <span>{t(item.key)}</span>
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
