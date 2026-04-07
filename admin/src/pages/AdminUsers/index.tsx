import {
  ModalForm,
  PageContainer,
  ProFormSelect,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { App, Button, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import dayjs from 'dayjs';
import {
  getAdmins,
  createAdmin,
  updateAdmin,
  deleteAdmin,
} from '@/services/api';

const roleColorMap: Record<string, string> = {
  super_admin: 'red',
  operator: 'blue',
  viewer: 'green',
};

const AdminUsers: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const { message, modal } = App.useApp();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editingAdmin, setEditingAdmin] = useState<API.AdminUser | null>(null);

  const handleDelete = (record: API.AdminUser) => {
    modal.confirm({
      title: 'Delete Admin',
      content: `Are you sure you want to delete admin "${record.name}"?`,
      okType: 'danger',
      onOk: async () => {
        try {
          await deleteAdmin(record.id);
          message.success('Admin deleted');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error?.message ?? 'Failed to delete admin');
        }
      },
    });
  };

  const columns: ProColumns<API.AdminUser>[] = [
    { title: 'Name', dataIndex: 'name' },
    { title: 'Email', dataIndex: 'email' },
    {
      title: 'Role',
      dataIndex: 'role',
      render: (_, record) => (
        <Tag color={roleColorMap[record.role] ?? 'default'}>{record.role}</Tag>
      ),
    },
    {
      title: 'Active',
      dataIndex: 'is_active',
      render: (_, record) => (
        <Tag color={record.is_active ? 'green' : 'red'}>
          {record.is_active ? 'Yes' : 'No'}
        </Tag>
      ),
    },
    {
      title: 'Last Login',
      dataIndex: 'last_login',
      render: (_, record) =>
        record.last_login ? dayjs(record.last_login).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: 'Actions',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="edit"
          onClick={() => {
            setEditingAdmin(record);
            setEditModalOpen(true);
          }}
        >
          Edit
        </a>,
        <a key="delete" style={{ color: 'red' }} onClick={() => handleDelete(record)}>
          Delete
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.AdminUser>
        headerTitle="Admin Users"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        search={false}
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            onClick={() => setCreateModalOpen(true)}
          >
            New Admin
          </Button>,
        ]}
        request={async () => {
          const res = await getAdmins();
          return {
            data: res.data ?? [],
            success: res.success,
          };
        }}
      />

      <ModalForm
        title="New Admin"
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onFinish={async (values) => {
          try {
            await createAdmin(values as any);
            message.success('Admin created');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error?.message ?? 'Failed to create admin');
            return false;
          }
        }}
      >
        <ProFormText
          name="email"
          label="Email"
          rules={[
            { required: true, message: 'Please enter email' },
            { type: 'email', message: 'Please enter a valid email' },
          ]}
        />
        <ProFormText.Password
          name="password"
          label="Password"
          rules={[{ required: true, message: 'Please enter password' }]}
        />
        <ProFormText
          name="name"
          label="Name"
          rules={[{ required: true, message: 'Please enter name' }]}
        />
        <ProFormSelect
          name="role"
          label="Role"
          rules={[{ required: true, message: 'Please select a role' }]}
          options={[
            { label: 'Super Admin', value: 'super_admin' },
            { label: 'Operator', value: 'operator' },
            { label: 'Viewer', value: 'viewer' },
          ]}
        />
      </ModalForm>

      <ModalForm
        title="Edit Admin"
        open={editModalOpen}
        onOpenChange={setEditModalOpen}
        initialValues={editingAdmin ? { name: editingAdmin.name, role: editingAdmin.role } : {}}
        modalProps={{ destroyOnClose: true }}
        onFinish={async (values) => {
          if (!editingAdmin) return false;
          try {
            await updateAdmin(editingAdmin.id, values as any);
            message.success('Admin updated');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error?.message ?? 'Failed to update admin');
            return false;
          }
        }}
      >
        <ProFormText
          name="name"
          label="Name"
          rules={[{ required: true, message: 'Please enter name' }]}
        />
        <ProFormSelect
          name="role"
          label="Role"
          rules={[{ required: true, message: 'Please select a role' }]}
          options={[
            { label: 'Super Admin', value: 'super_admin' },
            { label: 'Operator', value: 'operator' },
            { label: 'Viewer', value: 'viewer' },
          ]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default AdminUsers;
