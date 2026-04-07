'use client'

import { useState } from 'react'
import {
  NavBar,
  Tag,
  Button,
  Input,
  Popup,
  Picker,
  Dialog,
  Toast,
  Skeleton,
  SwipeAction,
} from '@arco-design/mobile-react'
import {
  useAdminList,
  useCreateAdmin,
  useUpdateAdmin,
  useDeleteAdmin,
} from '@/hooks/use-api'
import type { AdminUser } from '@/types/api'

const roleOptions = [
  [
    { label: 'Admin', value: 'admin' },
    { label: 'Super Admin', value: 'super_admin' },
  ],
]

function roleColor(role: string): string {
  if (role === 'super_admin') return '#722ed1'
  return '#165dff'
}

function AdminCard({
  admin,
  onEdit,
  onDelete,
}: {
  admin: AdminUser
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <SwipeAction
      rightActions={[
        {
          text: 'Edit',
          style: { background: '#165dff', color: '#fff', padding: '0 16px' },
          onClick: () => {
            onEdit()
          },
        },
        {
          text: 'Delete',
          style: { background: '#f53f3f', color: '#fff', padding: '0 16px' },
          onClick: () => {
            onDelete()
          },
        },
      ]}
    >
      <div
        style={{
          background: '#fff',
          padding: '14px 16px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <div style={{ fontSize: 15, fontWeight: 600, color: '#1d2129' }}>
            {admin.name}
          </div>
          <div style={{ fontSize: 12, color: '#86909c', marginTop: 4 }}>
            {admin.email}
          </div>
        </div>
        <Tag
          style={{
            fontSize: 10,
            borderColor: roleColor(admin.role),
            color: roleColor(admin.role),
          }}
        >
          {admin.role}
        </Tag>
      </div>
    </SwipeAction>
  )
}

export default function AdminUsersPage() {
  const { data: admins, isLoading } = useAdminList()
  const createAdmin = useCreateAdmin()
  const updateAdmin = useUpdateAdmin()
  const deleteAdmin = useDeleteAdmin()

  const [createOpen, setCreateOpen] = useState(false)
  const [createEmail, setCreateEmail] = useState('')
  const [createPassword, setCreatePassword] = useState('')
  const [createName, setCreateName] = useState('')
  const [createRole, setCreateRole] = useState('admin')
  const [rolePickerOpen, setRolePickerOpen] = useState(false)

  const [editOpen, setEditOpen] = useState(false)
  const [editId, setEditId] = useState('')
  const [editName, setEditName] = useState('')
  const [editRole, setEditRole] = useState('')
  const [editRolePickerOpen, setEditRolePickerOpen] = useState(false)

  const handleCreate = async () => {
    if (!createEmail.trim() || !createPassword.trim() || !createName.trim()) {
      Toast.toast('Please fill in all fields')
      return
    }
    try {
      await createAdmin.mutateAsync({
        email: createEmail,
        password: createPassword,
        name: createName,
        role: createRole,
      })
      setCreateOpen(false)
      setCreateEmail('')
      setCreatePassword('')
      setCreateName('')
      setCreateRole('admin')
      Toast.toast('Admin created')
    } catch {
      Toast.error('Failed to create admin')
    }
  }

  const handleOpenEdit = (admin: AdminUser) => {
    setEditId(admin.id)
    setEditName(admin.name)
    setEditRole(admin.role)
    setEditOpen(true)
  }

  const handleSaveEdit = async () => {
    try {
      await updateAdmin.mutateAsync({
        id: editId,
        data: { name: editName, role: editRole },
      })
      setEditOpen(false)
      Toast.toast('Admin updated')
    } catch {
      Toast.error('Failed to update admin')
    }
  }

  const handleDelete = (admin: AdminUser) => {
    Dialog.confirm({
      title: 'Delete Admin',
      children: `Are you sure you want to delete "${admin.name}"?`,
      onOk: async () => {
        try {
          await deleteAdmin.mutateAsync(admin.id)
          Toast.toast('Admin deleted')
        } catch {
          Toast.error('Failed to delete admin')
        }
      },
    })
  }

  return (
    <div>
      <NavBar
        title="Admins"
        leftContent={null}
        rightContent={
          <span
            style={{ fontSize: 22, cursor: 'pointer' }}
            onClick={() => setCreateOpen(true)}
          >
            +
          </span>
        }
      />
      <div style={{ padding: '0 0 16px' }}>
        {isLoading ? (
          <div style={{ padding: 16 }}>
            <Skeleton animation="gradient" />
            <Skeleton animation="gradient" style={{ marginTop: 12 }} />
          </div>
        ) : (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: 1,
              background: '#e5e6eb',
            }}
          >
            {(!admins || admins.length === 0) && (
              <div
                style={{
                  textAlign: 'center',
                  padding: 40,
                  color: '#86909c',
                  fontSize: 14,
                  background: '#fff',
                }}
              >
                No admins found
              </div>
            )}
            {admins?.map((admin) => (
              <AdminCard
                key={admin.id}
                admin={admin}
                onEdit={() => handleOpenEdit(admin)}
                onDelete={() => handleDelete(admin)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Create Admin Popup */}
      <Popup
        visible={createOpen}
        close={() => setCreateOpen(false)}
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
            Create Admin
          </h3>
          <Input
            label="Email"
            placeholder="admin@example.com"
            value={createEmail}
            onChange={(_e, val) => setCreateEmail(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="Password"
            type="password"
            placeholder="Password"
            value={createPassword}
            onChange={(_e, val) => setCreatePassword(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <Input
            label="Name"
            placeholder="Full name"
            value={createName}
            onChange={(_e, val) => setCreateName(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <div
            onClick={() => setRolePickerOpen(true)}
            style={{
              padding: '12px 0',
              borderBottom: '1px solid #e5e6eb',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              cursor: 'pointer',
              marginBottom: 20,
            }}
          >
            <span style={{ fontSize: 14, color: '#86909c' }}>Role</span>
            <span style={{ fontSize: 14, color: '#1d2129' }}>
              {createRole}
            </span>
          </div>
          <Picker
            visible={rolePickerOpen}
            value={[createRole]}
            data={roleOptions}
            cascade={false}
            onHide={() => setRolePickerOpen(false)}
            onOk={(val) => {
              setCreateRole(val[0] as string)
              setRolePickerOpen(false)
            }}
          />
          <Button
            onClick={handleCreate}
            disabled={createAdmin.isPending}
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
            {createAdmin.isPending ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </Popup>

      {/* Edit Admin Popup */}
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
            Edit Admin
          </h3>
          <Input
            label="Name"
            value={editName}
            onChange={(_e, val) => setEditName(val)}
            border="bottom"
            style={{ marginBottom: 12 }}
          />
          <div
            onClick={() => setEditRolePickerOpen(true)}
            style={{
              padding: '12px 0',
              borderBottom: '1px solid #e5e6eb',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              cursor: 'pointer',
              marginBottom: 20,
            }}
          >
            <span style={{ fontSize: 14, color: '#86909c' }}>Role</span>
            <span style={{ fontSize: 14, color: '#1d2129' }}>{editRole}</span>
          </div>
          <Picker
            visible={editRolePickerOpen}
            value={[editRole]}
            data={roleOptions}
            cascade={false}
            onHide={() => setEditRolePickerOpen(false)}
            onOk={(val) => {
              setEditRole(val[0] as string)
              setEditRolePickerOpen(false)
            }}
          />
          <Button
            onClick={handleSaveEdit}
            disabled={updateAdmin.isPending}
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
            {updateAdmin.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </Popup>
    </div>
  )
}
