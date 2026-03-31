import { useLocaleStore } from "@/stores/locale-store"
import { locales } from "@/i18n"

export function useT() {
  const { locale } = useLocaleStore()
  const dict = locales[locale]

  return function t(key: string): string {
    const [section, field] = key.split(".")
    return dict?.[section]?.[field] ?? key
  }
}
