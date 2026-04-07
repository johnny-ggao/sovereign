'use client'

import { usePathname, useRouter } from 'next/navigation'
import { TabBar } from '@arco-design/mobile-react'
import { useAuthStore } from '@/lib/auth'

interface TabItem {
  title: string
  icon: string
  path: string
}

const baseTabs: TabItem[] = [
  { title: 'Dashboard', icon: '📊', path: '/dashboard' },
  { title: 'Users', icon: '👥', path: '/users' },
]

const adminTab: TabItem = { title: 'Admins', icon: '🔑', path: '/admin-users' }
const meTab: TabItem = { title: 'Me', icon: '👤', path: '/profile' }

function getTabsForRole(role: string): TabItem[] {
  if (role === 'super_admin') {
    return [...baseTabs, adminTab, meTab]
  }
  return [...baseTabs, meTab]
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const admin = useAuthStore((s) => s.admin)

  const tabs = getTabsForRole(admin?.role ?? '')

  const activeIndex = tabs.findIndex((tab) => pathname.startsWith(tab.path))
  const safeIndex = activeIndex >= 0 ? activeIndex : 0

  return (
    <div style={{ minHeight: '100vh', paddingBottom: 50 }}>
      {children}
      <TabBar
        fixed
        activeIndex={safeIndex}
        onChange={(index) => {
          const target = tabs[index]
          if (target) {
            router.push(target.path)
          }
        }}
      >
        {tabs.map((tab) => (
          <TabBar.Item
            key={tab.path}
            title={tab.title}
            icon={<span style={{ fontSize: 20 }}>{tab.icon}</span>}
          />
        ))}
      </TabBar>
    </div>
  )
}
