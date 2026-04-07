'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/auth'
import AppLayout from '@/components/layout/app-layout'

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    if (!isLoggedIn()) {
      router.replace('/login')
    } else {
      setChecked(true)
    }
  }, [isLoggedIn, router])

  if (!checked) {
    return null
  }

  return <AppLayout>{children}</AppLayout>
}
