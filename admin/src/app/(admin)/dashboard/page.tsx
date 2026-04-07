'use client'

import { NavBar, Skeleton, Tag } from '@arco-design/mobile-react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { useDashboardStats } from '@/hooks/use-api'
import type { TransactionInfo } from '@/types/api'

interface StatCardProps {
  label: string
  value: string | number
  color: string
}

function StatCard({ label, value, color }: StatCardProps) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 12,
        padding: '16px 14px',
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
      }}
    >
      <span style={{ fontSize: 12, color: '#86909c' }}>{label}</span>
      <span style={{ fontSize: 20, fontWeight: 600, color }}>{value}</span>
    </div>
  )
}

function statusColor(status: string): string {
  const map: Record<string, string> = {
    completed: '#00b42a',
    pending: '#ff7d00',
    failed: '#f53f3f',
    processing: '#165dff',
  }
  return map[status] ?? '#86909c'
}

function TransactionCard({ tx }: { tx: TransactionInfo }) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 10,
        padding: '12px 14px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <div>
        <div style={{ fontSize: 14, fontWeight: 500, color: '#1d2129' }}>
          {tx.type}
        </div>
        <div style={{ fontSize: 12, color: '#86909c', marginTop: 2 }}>
          {new Date(tx.created_at).toLocaleDateString()}
        </div>
      </div>
      <div style={{ textAlign: 'right' }}>
        <div style={{ fontSize: 14, fontWeight: 600, color: '#1d2129' }}>
          {tx.amount} {tx.currency}
        </div>
        <Tag
          style={{
            marginTop: 4,
            fontSize: 10,
            borderColor: statusColor(tx.status),
            color: statusColor(tx.status),
          }}
        >
          {tx.status}
        </Tag>
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const { data: stats, isLoading } = useDashboardStats()

  if (isLoading || !stats) {
    return (
      <div>
        <NavBar title="Dashboard" leftContent={null} />
        <div style={{ padding: 16 }}>
          <Skeleton animation="gradient" />
          <Skeleton animation="gradient" style={{ marginTop: 16 }} />
          <Skeleton animation="gradient" style={{ marginTop: 16 }} />
        </div>
      </div>
    )
  }

  const chartData = stats.user_trend.map((item) => ({
    ...item,
    date: item.date.slice(5),
  }))

  return (
    <div>
      <NavBar title="Dashboard" leftContent={null} />
      <div style={{ padding: 16 }}>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: 12,
          }}
        >
          <StatCard
            label="Total Users"
            value={stats.total_users}
            color="#1d2129"
          />
          <StatCard
            label="New Today"
            value={stats.new_users_today}
            color="#165dff"
          />
          <StatCard
            label="Total Invested"
            value={`$${stats.total_invested}`}
            color="#00b42a"
          />
          <StatCard
            label="Deposits"
            value={`$${stats.total_deposits}`}
            color="#722ed1"
          />
          <StatCard
            label="Withdrawals"
            value={`$${stats.total_withdrawals}`}
            color="#ff7d00"
          />
          <StatCard
            label="Active Investments"
            value={stats.active_investments}
            color="#0fc6c2"
          />
        </div>

        <div
          style={{
            marginTop: 20,
            background: '#fff',
            borderRadius: 12,
            padding: '16px 8px 8px',
          }}
        >
          <h3
            style={{
              fontSize: 14,
              fontWeight: 600,
              marginBottom: 12,
              paddingLeft: 8,
              color: '#1d2129',
            }}
          >
            User Trend (30 days)
          </h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e6eb" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 10 }}
                stroke="#c9cdd4"
              />
              <YAxis tick={{ fontSize: 10 }} stroke="#c9cdd4" />
              <Tooltip />
              <Line
                type="monotone"
                dataKey="count"
                stroke="#165dff"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div style={{ marginTop: 20 }}>
          <h3
            style={{
              fontSize: 14,
              fontWeight: 600,
              marginBottom: 12,
              color: '#1d2129',
            }}
          >
            Recent Transactions
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {stats.recent_transactions.length === 0 && (
              <div
                style={{
                  textAlign: 'center',
                  padding: 24,
                  color: '#86909c',
                  fontSize: 14,
                }}
              >
                No recent transactions
              </div>
            )}
            {stats.recent_transactions.map((tx) => (
              <TransactionCard key={tx.id} tx={tx} />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
