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
      title: '删除管理员',
      content: `确认删除管理员"${record.name}"？`,
      okType: 'danger',
      onOk: async () => {
        try {
          await deleteAdmin(record.id);
          message.success('已删除');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error?.message ?? '删除管理员失败');
        }
      },
    });
  };

  const columns: ProColumns<API.AdminUser>[] = [
    { title: '姓名', dataIndex: 'name' },
    { title: '邮箱', dataIndex: 'email' },
    {
      title: '角色',
      dataIndex: 'role',
      render: (_, record) => {
        const roleLabelMap: Record<string, string> = {
          super_admin: '超级管理员',
          operator: '操作员',
          viewer: '观察者',
        };
        return <Tag color={roleColorMap[record.role] ?? 'default'}>{roleLabelMap[record.role] ?? record.role}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      render: (_, record) => (
        <Tag color={record.is_active ? 'green' : 'red'}>
          {record.is_active ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'last_login',
      render: (_, record) =>
        record.last_login ? dayjs(record.last_login).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="edit"
          onClick={() => {
            setEditingAdmin(record);
            setEditModalOpen(true);
          }}
        >
          编辑
        </a>,
        <a key="delete" style={{ color: 'red' }} onClick={() => handleDelete(record)}>
          删除
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.AdminUser>
        headerTitle="管理员管理"
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
            新增管理员
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
        title="创建管理员"
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onFinish={async (values) => {
          try {
            await createAdmin(values as any);
            message.success('已创建');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error?.message ?? '创建管理员失败');
            return false;
          }
        }}
      >
        <ProFormText
          name="email"
          label="邮箱"
          rules={[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱' },
          ]}
        />
        <ProFormText.Password
          name="password"
          label="密码"
          rules={[{ required: true, message: '请输入密码' }]}
        />
        <ProFormText
          name="name"
          label="姓名"
          rules={[{ required: true, message: '请输入姓名' }]}
        />
        <ProFormSelect
          name="role"
          label="角色"
          rules={[{ required: true, message: '请选择角色' }]}
          options={[
            { label: '超级管理员', value: 'super_admin' },
            { label: '操作员', value: 'operator' },
            { label: '观察者', value: 'viewer' },
          ]}
        />
      </ModalForm>

      <ModalForm
        title="编辑管理员"
        open={editModalOpen}
        onOpenChange={setEditModalOpen}
        initialValues={editingAdmin ? { name: editingAdmin.name, role: editingAdmin.role } : {}}
        modalProps={{ destroyOnClose: true }}
        onFinish={async (values) => {
          if (!editingAdmin) return false;
          try {
            await updateAdmin(editingAdmin.id, values as any);
            message.success('已更新');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error?.message ?? '更新管理员失败');
            return false;
          }
        }}
      >
        <ProFormText
          name="name"
          label="姓名"
          rules={[{ required: true, message: '请输入姓名' }]}
        />
        <ProFormSelect
          name="role"
          label="角色"
          rules={[{ required: true, message: '请选择角色' }]}
          options={[
            { label: '超级管理员', value: 'super_admin' },
            { label: '操作员', value: 'operator' },
            { label: '观察者', value: 'viewer' },
          ]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default AdminUsers;
