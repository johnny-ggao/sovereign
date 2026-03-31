"use client"

import { useLocaleStore } from "@/stores/locale-store"
import { localeLabels, type Locale } from "@/i18n"
import { Globe } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export function LocaleSwitcher() {
  const { locale, setLocale } = useLocaleStore()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground focus:outline-none">
        <Globe className="h-4 w-4" />
        <span>{localeLabels[locale]}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="bg-card">
        {(Object.entries(localeLabels) as [Locale, string][]).map(([code, label]) => (
          <DropdownMenuItem
            key={code}
            onClick={() => setLocale(code)}
            className={locale === code ? "text-primary" : ""}
          >
            {label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
