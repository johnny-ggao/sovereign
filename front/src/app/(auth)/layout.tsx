import { LocaleSwitcher } from "@/components/layout/locale-switcher"

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative min-h-screen bg-background">
      <div className="absolute right-4 top-4 z-10">
        <LocaleSwitcher />
      </div>
      {children}
    </div>
  )
}
