"use client"

import { useRouter } from "next/navigation"
import { useAuthStore } from "@/stores/auth-store"
import { useT } from "@/hooks/use-t"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  User,
  Shield,
  Bell,
  Globe,
  LogOut,
  ChevronDown,
} from "lucide-react"

export function TopBar() {
  const router = useRouter()
  const { user, clearAuth } = useAuthStore()
  const t = useT()

  const initials = user?.full_name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) || "SV"

  function handleLogout() {
    clearAuth()
    router.push("/login")
  }

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border/30 glass px-4 md:px-8">
      <div className="flex items-center gap-2 md:hidden">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary">
          <span className="text-xs font-bold text-primary-foreground">S</span>
        </div>
        <span className="text-sm font-semibold">Sovereign</span>
      </div>

      <div className="hidden md:block" />

      <DropdownMenu>
        <DropdownMenuTrigger className="flex items-center gap-2 rounded-full px-2 py-1 transition-colors hover:bg-accent/50 focus:outline-none">
          <Avatar className="h-8 w-8">
            <AvatarFallback className="bg-accent text-xs">{initials}</AvatarFallback>
          </Avatar>
          <div className="hidden items-start text-left md:flex md:flex-col">
            <span className="text-sm font-medium">{user?.full_name || "User"}</span>
            <span className="text-xs text-muted-foreground">{user?.email}</span>
          </div>
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        </DropdownMenuTrigger>

        <DropdownMenuContent align="end" className="w-56 bg-card">
          <div className="px-3 py-2 md:hidden">
            <p className="text-sm font-medium">{user?.full_name}</p>
            <p className="text-xs text-muted-foreground">{user?.email}</p>
          </div>
          <DropdownMenuSeparator className="md:hidden" />

          <DropdownMenuItem onClick={() => router.push("/settings?tab=profile")}>
            <User className="mr-2 h-4 w-4" />
            {t("userCenter.profile")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => router.push("/settings?tab=security")}>
            <Shield className="mr-2 h-4 w-4" />
            {t("userCenter.security")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => router.push("/settings?tab=notifications")}>
            <Bell className="mr-2 h-4 w-4" />
            {t("userCenter.notifications")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => router.push("/settings?tab=profile")}>
            <Globe className="mr-2 h-4 w-4" />
            {t("userCenter.language")}
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          <DropdownMenuItem onClick={handleLogout} className="text-destructive focus:text-destructive">
            <LogOut className="mr-2 h-4 w-4" />
            {t("userCenter.logout")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
