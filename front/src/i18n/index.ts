import en from "./en.json"
import ko from "./ko.json"
import zh from "./zh.json"

export type Locale = "en" | "ko" | "zh"

export const locales: Record<Locale, Record<string, Record<string, string>>> = { en, ko, zh }

export const localeLabels: Record<Locale, string> = {
  ko: "한국어",
  en: "English",
  zh: "中文",
}
