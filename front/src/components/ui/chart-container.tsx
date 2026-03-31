"use client"

import { useRef, useState, useEffect, type ReactNode } from "react"

export function ChartContainer({ children, className }: { children: ReactNode; className?: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    if (!ref.current) return
    const el = ref.current
    // 等容器有实际尺寸后才渲染图表
    const check = () => {
      if (el.clientWidth > 0 && el.clientHeight > 0) {
        setReady(true)
      }
    }
    check()
    if (!ready) {
      const observer = new ResizeObserver(check)
      observer.observe(el)
      return () => observer.disconnect()
    }
  }, [ready])

  return (
    <div ref={ref} className={className}>
      {ready ? children : null}
    </div>
  )
}
