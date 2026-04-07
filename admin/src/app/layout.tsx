import type { Metadata, Viewport } from 'next'
import './globals.css'
import '@arco-design/mobile-react/dist/style.css'
import { Providers } from './providers'

export const metadata: Metadata = {
  title: 'Sovereign Admin',
  description: 'Sovereign Fund Admin Panel',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
