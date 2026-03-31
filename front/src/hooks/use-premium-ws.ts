"use client"

import { useEffect, useRef, useState, useMemo, useCallback } from "react"
import type { PremiumTick } from "@/types/api"

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws/v1/premium"

interface UsePremiumWSOptions {
  pairs?: string[]
  enabled?: boolean
}

export function usePremiumWS({ pairs = ["BTC/KRW", "ETH/KRW", "SOL/KRW", "XRP/KRW"], enabled = true }: UsePremiumWSOptions = {}) {
  const [ticks, setTicks] = useState<PremiumTick[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const bufferRef = useRef<Map<string, PremiumTick>>(new Map())
  const flushTimer = useRef<ReturnType<typeof setTimeout>>(undefined)

  const stablePairs = useMemo(() => pairs, [JSON.stringify(pairs)])

  // 节流刷新：收集 WS 消息，每 500ms 批量更新一次 state
  const flushBuffer = useCallback(() => {
    const values = Array.from(bufferRef.current.values())
    if (values.length > 0) {
      setTicks(values)
    }
  }, [])

  useEffect(() => {
    if (!enabled) return

    let disposed = false

    function connect() {
      if (disposed) return

      const ws = new WebSocket(WS_URL)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        ws.send(JSON.stringify({ action: "subscribe", pairs: stablePairs }))
      }

      ws.onmessage = (event) => {
        const msg = JSON.parse(event.data)
        if (msg.type === "tick") {
          const tick = msg.data as PremiumTick
          bufferRef.current.set(tick.pair, tick)

          // 节流：500ms 内只刷新一次
          if (!flushTimer.current) {
            flushTimer.current = setTimeout(() => {
              flushBuffer()
              flushTimer.current = undefined
            }, 500)
          }
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        if (!disposed) {
          reconnectTimer.current = setTimeout(connect, 3000)
        }
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    connect()

    return () => {
      disposed = true
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      if (flushTimer.current) clearTimeout(flushTimer.current)
      wsRef.current?.close()
    }
  }, [enabled, stablePairs, flushBuffer])

  return {
    ticks,
    isConnected,
  }
}
