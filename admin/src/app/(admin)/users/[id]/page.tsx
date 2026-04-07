'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import {
  NavBar,
  Tabs,
  Tag,
  Button,
  Input,
  Popup,
  Dialog,
  Toast,
  Skeleton,
} from '@arco-design/mobile-react'
import {
  useUserDetail,
  useUpdateUser,
  useDisableUser,
  useEnableUser,
  useResetPassword,
  useAdjustBalance,
} from '@/hooks/use-api'
import { useAuthStore } from '@/lib/auth'
import type {
  WalletInfo,
  TransactionInfo,
  InvestmentInfo,
  SettlementInfo,
} from '@/types/api'

function WalletCard({ w }: { w: WalletInfo }) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 10,
        padding: '12px 14px',
        marginBottom: 8,
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <span style={{ fontSize: 13, fontWeight: 500, color: '#1d2129' }}>
          {w.network}
        </span>
        <span style={{ fontSize: 14, fontWeight: 600 }}>{w.balance}</span>
      </div>
      <div
        style={{
          fontSize: 11,
          color: '#86909c',
          marginTop: 4,
          wordBreak: 'break-all',
        }}
      >
        {w.address}
      </div>
    </div>
  )
}

function TransactionCard({ tx }: { tx: TransactionInfo }) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 10,
        padding: '12px 14px',
        marginBottom: 8,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <div>
        <div style={{ fontSize: 13, fontWeight: 500, color: '#1d2129' }}>
          {tx.type}
        </div>
        <div style={{ fontSize: 11, color: '#86909c', marginTop: 2 }}>
          {new Date(tx.created_at).toLocaleDateString()}
        </div>
      </div>
      <div style={{ textAlign: 'right' }}>
        <div style={{ fontSize: 13, fontWeight: 600 }}>
          {tx.amount} {tx.currency}
        </div>
        <Tag style={{ fontSize: 10, marginTop: 2 }}>{tx.status}</Tag>
      </div>
    </div>
  )
}

function InvestmentCard({ inv }: { inv: InvestmentInfo }) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 10,
        padding: '12px 14px',
        marginBottom: 8,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <div>
        <div style={{ fontSize: 13, fontWeight: 500, color: '#1d2129' }}>
          {inv.amount} {inv.currency}
        </div>
        <div style={{ fontSize: 11, color: '#86909c', marginTop: 2 }}>
          Started: {new Date(inv.started_at).toLocaleDateString()}
        </div>
      </div>
      <Tag style={{ fontSize: 10 }}>{inv.status}</Tag>
    </div>
  )
}

function SettlementCard({ s }: { s: SettlementInfo }) {
  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 10,
        padding: '12px 14px',
        marginBottom: 8,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <div>
        <div style={{ fontSize: 13, fontWeight: 500, color: '#1d2129' }}>
          {s.type} - {s.amount} {s.currency}
        </div>
        <div style={{ fontSize: 11, color: '#86909c', marginTop: 2 }}>
          {new Date(s.settled_at).toLocaleDateString()}
        </div>
      </div>
      <Tag style={{ fontSize: 10 }}>{s.status}</Tag>
    </div>
  )
}

export default function UserDetailPage() {
  const router = useRouter()
  const params = useParams()
  const userId = params.id as string
  const admin = useAuthStore((s) => s.admin)
  const isSuperAdmin = admin?.role === 'super_admin'

  const { data: user, isLoading } = useUserDetail(userId)
  const updateUser = useUpdateUser()
  const disableUser = useDisableUser()
  const enableUser = useEnableUser()
  const resetPassword = useResetPassword()
  const adjustBalance = useAdjustBalance()

  const [editOpen, setEditOpen] = useState(false)
  const [editName, setEditName] = useState('')
  const [editPhone, setEditPhone] = useState('')
  const [editLang, setEditLang] = useState('')

  const [balanceOpen, setBalanceOpen] = useState(false)
  const [balanceAmount, setBalanceAmount] = useState('')
  const [balanceReason, setBalanceReason] = useState('')

  const handleOpenEdit = () => {
    if (!user) return
    setEditName(user.full_name)
    setEditPhone(user.phone)
    setEditLang(user.language)
    setEditOpen(true)
  }

  const handleSaveEdit = async () => {
    try {
      await updateUser.mutateAsync({
        id: userId,
        data: {
          full_name: editName,
          phone: editPhone,
          language: editLang,
        },
      })
      setEditOpen(false)
      Toast.toast('User updated')
    } catch {
      Toast.error('Failed to update user')
    }
  }

  const handleToggleDisable = () => {
    if (!user) return
    const action = user.is_active ? 'disable' : 'enable'
    Dialog.confirm({
      title: `${action === 'disable' ? 'Disable' : 'Enable'} User`,
      children: `Are you sure you want to ${action} this user?`,
      onOk: async () => {
        try {
          if (action === 'disable') {
            await disableUser.mutateAsync(userId)
          } else {
            await enableUser.mutateAsync(userId)
          }
          Toast.toast(`User ${action}d`)
        } catch {
          Toast.error(`Failed to ${action} user`)
        }
      },
    })
  }

  const handleResetPassword = () => {
    Dialog.confirm({
      title: 'Reset Password',
      children: 'Are you sure you want to reset this user\'s password?',
      onOk: async () => {
        try {
          const result = await resetPassword.mutateAsync(userId)
          Dialog.alert({
            title: 'New Password',
            children: result.temporary_password,
          })
        } catch {
          Toast.error('Failed to reset password')
        }
      },
    })
  }

  const handleAdjustBalance = async () => {
    if (!balanceAmount.trim() || !balanceReason.trim()) {
      Toast.toast('Please fill in all fields')
      return
    }
    try {
      await adjustBalance.mutateAsync({
        id: userId,
        amount: balanceAmount,
        reason: balanceReason,
      })
      setBalanceOpen(false)
      setBalanceAmount('')
      setBalanceReason('')
      Toast.toast('Balance adjusted')
    } catch {
      Toast.error('Failed to adjust balance')
    }
  }

  if (isLoading || !user) {
    return (
      <div>
        <NavBar
          title="User Detail"
          onClickLeft={() => router.back()}
        />
        <div style={{ padding: 16 }}>
          <Skeleton animation="gradient" />
          <Skeleton animation="gradient" style={{ marginTop: 16 }} />
        </div>
      </div>
    )
  }

  return (
    <div>
      <NavBar
        title="User Detail"
        onClickLeft={() => router.back()}
      />
      <div style={{ padding: 16 }}>
        {/* User Info Card */}
        <div
          style={{
            background: '#fff',
            borderRadius: 12,
            padding: 16,
            marginBottom: 12,
          }}
        >
          <div style={{ fontSize: 17, fontWeight: 600, color: '#1d2129' }}>
            {user.full_name || 'Unnamed'}
          </div>
          <div style={{ fontSize: 13, color: '#86909c', marginTop: 4 }}>
            {user.email}
          </div>
          <div style={{ fontSize: 12, color: '#c9cdd4', marginTop: 4 }}>
            Joined: {new Date(user.created_at).toLocaleDateString()}
          </div>
          <Tag
            style={{
              marginTop: 8,
              fontSize: 10,
              borderColor: user.is_active ? '#00b42a' : '#f53f3f',
              color: user.is_active ? '#00b42a' : '#f53f3f',
            }}
          >
            {user.is_active ? 'Active' : 'Disabled'}
          </Tag>
        </div>

        {/* Action Buttons */}
        <div
          style={{
            display: 'flex',
            gap: 8,
            flexWrap: 'wrap',
            marginBottom: 16,
          }}
        >
          <Button
            onClick={handleOpenEdit}
            style={{
              flex: 1,
              minWidth: 'auto',
              height: 36,
              fontSize: 13,
              borderRadius: 8,
            }}
          >
            Edit
          </Button>
          <Button
            onClick={handleToggleDisable}
            style={{
              flex: 1,
              minWidth: 'auto',
              height: 36,
              fontSize: 13,
              borderRadius: 8,
              borderColor: user.is_active ? '#f53f3f' : '#00b42a',
              color: user.is_active ? '#f53f3f' : '#00b42a',
            }}
          >
            {user.is_active ? 'Disable' : 'Enable'}
          </Button>
          {isSuperAdmin && (
            <>
              <Button
                onClick={handleResetPassword}
                style={{
                  flex: 1,
                  minWidth: 'auto',
                  height: 36,
                  fontSize: 13,
                  borderRadius: 8,
                }}
              >
                Reset Pwd
              </Button>
              <Button
                onClick={() => setBalanceOpen(true)}
                style={{
                  flex: 1,
                  minWidth: 'auto',
                  height: 36,
                  fontSize: 13,
                  borderRadius: 8,
                }}
              >
                Adjust Bal
              </Button>
            </>
          )}
        </div>

        {/* Tabs */}
        <Tabs
          tabs={['Wallets', 'Transactions', 'Investments', 'Settlements']}
          type="line-divide"
          defaultActiveTab={0}
        >
          <div style={{ padding: '12px 0' }}>
            {user.wallets.length === 0 && (
              <EmptyHint text="No wallets" />
            )}
            {user.wallets.map((w) => (
              <WalletCard key={w.id} w={w} />
            ))}
          </div>
          <div style={{ padding: '12px 0' }}>
            {user.recent_transactions.length === 0 && (
              <EmptyHint text="No transactions" />
            )}
            {user.recent_transactions.map((tx) => (
              <TransactionCard key={tx.id} tx={tx} />
            ))}
          </div>
          <div style={{ padding: '12px 0' }}>
            {user.investments.length === 0 && (
              <EmptyHint text="No investments" />
            )}
            {user.investments.map((inv) => (
              <InvestmentCard key={inv.id} inv={inv} />
            ))}
          </div>
          <div style={{ padding: '12px 0' }}>
            {user.recent_settlements.length === 0 && (
              <EmptyHint text="No settlements" />
            )}
            {user.recent_settlements.map((s) => (
              <SettlementCard key={s.id} s={s} />
            ))}
          </div>
        </Tabs>
      </div>

      {/* Edit Popup */}
      <Popup
        visible={editOpen}
        close={() => setEditOpen(false)}
        direction="bottom"
        maskClosable
      >
        <div style={{ padding: 20, paddingBottom: 40 }}>
          <h3
            style={{
              fontSize: 16,
              fontWeight: 600,
              marginBottom: 16,
              color: '#1d2129',
            }}
          >
            Edit User
          </h3>
          <Input
            label="Name"
            value={editName}
            onChange={(_e, val) => setEditName(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="Phone"
            value={editPhone}
            onChange={(_e, val) => setEditPhone(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="Language"
            value={editLang}
            onChange={(_e, val) => setEditLang(val)}
            border="bottom"
            style={{ marginBottom: 20 }}
          />
          <Button
            onClick={handleSaveEdit}
            disabled={updateUser.isPending}
            style={{
              width: '100%',
              height: 44,
              borderRadius: 8,
              background: '#165dff',
              color: '#fff',
              border: 'none',
              fontSize: 15,
            }}
          >
            {updateUser.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </Popup>

      {/* Adjust Balance Popup */}
      <Popup
        visible={balanceOpen}
        close={() => setBalanceOpen(false)}
        direction="bottom"
        maskClosable
      >
        <div style={{ padding: 20, paddingBottom: 40 }}>
          <h3
            style={{
              fontSize: 16,
              fontWeight: 600,
              marginBottom: 16,
              color: '#1d2129',
            }}
          >
            Adjust Balance
          </h3>
          <Input
            label="Amount"
            placeholder="e.g. 100 or -50"
            value={balanceAmount}
            onChange={(_e, val) => setBalanceAmount(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="Reason"
            placeholder="Reason for adjustment"
            value={balanceReason}
            onChange={(_e, val) => setBalanceReason(val)}
            border="bottom"
            style={{ marginBottom: 20 }}
          />
          <Button
            onClick={handleAdjustBalance}
            disabled={adjustBalance.isPending}
            style={{
              width: '100%',
              height: 44,
              borderRadius: 8,
              background: '#165dff',
              color: '#fff',
              border: 'none',
              fontSize: 15,
            }}
          >
            {adjustBalance.isPending ? 'Adjusting...' : 'Confirm'}
          </Button>
        </div>
      </Popup>
    </div>
  )
}

function EmptyHint({ text }: { text: string }) {
  return (
    <div
      style={{
        textAlign: 'center',
        padding: 32,
        color: '#86909c',
        fontSize: 13,
      }}
    >
      {text}
    </div>
  )
}
