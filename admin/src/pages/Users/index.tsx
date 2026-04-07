import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App, Button, Typography } from 'antd';
import React, { useRef } from 'react';
import { getUsers, resetUserPassword } from '@/services/api';
import dayjs from 'dayjs';

const Users: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const { message, modal } = App.useApp();

  const handleResetPassword = (record: API.UserListItem) => {
    modal.confirm({
      title: 'Reset Password',
      content: `Are you sure you want to reset the password for ${record.email}?`,
      onOk: async () => {
        try {
          const res = await resetUserPassword(record.id);
          if (res.success && res.data) {
            modal.success({
              title: 'Password Reset',
              content: `Temporary password: ${res.data.temporary_password}`,
            });
          }
        } catch (error: any) {
          message.error(error?.message ?? 'Failed to reset password');
        }
      },
    });
  };

  const columns: ProColumns<API.UserListItem>[] = [
    {
      title: 'Email',
      dataIndex: 'email',
      copyable: true,
    },
    {
      title: 'Name',
      dataIndex: 'full_name',
      search: false,
    },
    {
      title: 'Balance',
      dataIndex: 'balance',
      search: false,
      render: (_, record) => `$${record.balance}`,
    },
    {
      title: 'Joined',
      dataIndex: 'created_at',
      search: false,
      render: (_, record) => dayjs(record.created_at).format('YYYY-MM-DD'),
    },
    {
      title: 'Actions',
      valueType: 'option',
      render: (_, record) => [
        <a key="detail" onClick={() => history.push(`/users/${record.id}`)}>
          Detail
        </a>,
        <a key="reset" onClick={() => handleResetPassword(record)}>
          Reset Password
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.UserListItem>
        headerTitle="Users"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async (params) => {
          const res = await getUsers({
            page: params.current,
            per_page: params.pageSize,
            email: params.email,
          });
          return {
            data: res.data ?? [],
            success: res.success,
            total: res.meta?.total ?? 0,
          };
        }}
        pagination={{ defaultPageSize: 20 }}
      />
    </PageContainer>
  );
};

export default Users;
