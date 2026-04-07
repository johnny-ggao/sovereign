'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  NavBar,
  Cell,
  Button,
  Input,
  Popup,
  Dialog,
  Toast,
} from '@arco-design/mobile-react'
import { useAuthStore } from '@/lib/auth'
import { useChangePassword } from '@/hooks/use-api'

export default function ProfilePage() {
  const router = useRouter()
  const admin = useAuthStore((s) => s.admin)
  const logout = useAuthStore((s) => s.logout)
  const changePassword = useChangePassword()

  const [pwdOpen, setPwdOpen] = useState(false)
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')

  const handleChangePassword = async () => {
    if (!oldPwd.trim() || !newPwd.trim()) {
      Toast.toast('Please fill in all fields')
      return
    }
    try {
      await changePassword.mutateAsync({
        current_password: oldPwd,
        new_password: newPwd,
      })
      setPwdOpen(false)
      setOldPwd('')
      setNewPwd('')
      Toast.toast('Password changed')
    } catch {
      Toast.error('Failed to change password')
    }
  }

  const handleLogout = () => {
    Dialog.confirm({
      title: 'Logout',
      children: 'Are you sure you want to logout?',
      onOk: () => {
        logout()
        router.replace('/login')
      },
    })
  }

  return (
    <div>
      <NavBar title="Profile" leftContent={null} />
      <div style={{ padding: 16 }}>
        {/* Admin Info Card */}
        <div
          style={{
            background: '#fff',
            borderRadius: 12,
            padding: 20,
            textAlign: 'center',
            marginBottom: 16,
          }}
        >
          <div style={{ fontSize: 48, marginBottom: 8 }}>👤</div>
          <div style={{ fontSize: 18, fontWeight: 600, color: '#1d2129' }}>
            {admin?.name ?? 'Admin'}
          </div>
          <div style={{ fontSize: 13, color: '#86909c', marginTop: 4 }}>
            {admin?.email ?? ''}
          </div>
          <div
            style={{
              display: 'inline-block',
              marginTop: 8,
              padding: '2px 10px',
              borderRadius: 4,
              fontSize: 12,
              background:
                admin?.role === 'super_admin' ? '#f0e5ff' : '#e8f3ff',
              color: admin?.role === 'super_admin' ? '#722ed1' : '#165dff',
            }}
          >
            {admin?.role ?? 'admin'}
          </div>
        </div>

        {/* Settings */}
        <div
          style={{
            background: '#fff',
            borderRadius: 12,
            overflow: 'hidden',
            marginBottom: 24,
          }}
        >
          <Cell
            label="Change Password"
            showArrow
            clickable
            bordered={false}
            onClick={() => setPwdOpen(true)}
          />
        </div>

        {/* Logout */}
        <Button
          onClick={handleLogout}
          style={{
            width: '100%',
            height: 44,
            borderRadius: 8,
            color: '#f53f3f',
            borderColor: '#f53f3f',
            fontSize: 15,
            background: '#fff',
          }}
        >
          Logout
        </Button>
      </div>

      {/* Change Password Popup */}
      <Popup
        visible={pwdOpen}
        close={() => setPwdOpen(false)}
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
            Change Password
          </h3>
          <Input
            label="Current"
            type="password"
            placeholder="Current password"
            value={oldPwd}
            onChange={(_e, val) => setOldPwd(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="New"
            type="password"
            placeholder="New password"
            value={newPwd}
            onChange={(_e, val) => setNewPwd(val)}
            border="bottom"
            style={{ marginBottom: 20 }}
          />
          <Button
            onClick={handleChangePassword}
            disabled={changePassword.isPending}
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
            {changePassword.isPending ? 'Changing...' : 'Confirm'}
          </Button>
        </div>
      </Popup>
    </div>
  )
}
