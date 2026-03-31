"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useAuthStore } from "@/stores/auth-store"
import { Sidebar } from "./sidebar"
import { TopBar } from "./top-bar"
import { BottomNav } from "./bottom-nav"

export function AppShell({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { isAuthenticated } = useAuthStore()
  const [hydrated, setHydrated] = useState(false)

  useEffect(() => {
    setHydrated(true)
  }, [])

  useEffect(() => {
    if (hydrated && !isAuthenticated) {
      router.replace("/login")
    }
  }, [hydrated, isAuthenticated, router])

  if (!hydrated || !isAuthenticated) return null

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex flex-1 flex-col md:ml-64">
        <TopBar />
        <main className="flex-1 pb-20 md:pb-0">
          <div className="px-4 py-6 md:p-8">{children}</div>
        </main>
      </div>
      <BottomNav />
    </div>
  )
}
