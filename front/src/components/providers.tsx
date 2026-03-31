"use client"

import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { useState, useEffect, type ReactNode } from "react"
import { Toaster } from "@/components/ui/sonner"
import { GoogleOAuthProvider } from "@react-oauth/google"

const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || ""

export function Providers({ children }: { children: ReactNode }) {
  useEffect(() => {
    // 抑制浏览器扩展注入脚本的 unhandledRejection
    function handleRejection(e: PromiseRejectionEvent) {
      if (e.reason?.stack?.includes("chrome-extension://")) {
        e.preventDefault()
      }
    }
    window.addEventListener("unhandledrejection", handleRejection)

    // 抑制 Recharts ResponsiveContainer 的 -1 尺寸警告（SSR hydration 期间的已知问题）
    const origWarn = console.error
    console.error = (...args: unknown[]) => {
      if (typeof args[0] === "string" && args[0].includes("width(-1) and height(-1)")) return
      origWarn.apply(console, args)
    }

    return () => {
      window.removeEventListener("unhandledrejection", handleRejection)
      console.error = origWarn
    }
  }, [])

  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30 * 1000,
            retry: 1,
            refetchOnWindowFocus: false,
          },
        },
      }),
  )

  return (
    <GoogleOAuthProvider clientId={GOOGLE_CLIENT_ID}>
      <QueryClientProvider client={queryClient}>
        {children}
        <Toaster />
      </QueryClientProvider>
    </GoogleOAuthProvider>
  )
}
