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
      title: '重置密码',
      content: `确认重置密码：${record.email}？`,
      onOk: async () => {
        try {
          const res = await resetUserPassword(record.id);
          if (res.success && res.data) {
            modal.success({
              title: '密码已重置',
              content: `新密码：${res.data.temporary_password}`,
            });
          }
        } catch (error: any) {
          message.error(error?.message ?? '重置密码失败');
        }
      },
    });
  };

  const columns: ProColumns<API.UserListItem>[] = [
    {
      title: '邮箱',
      dataIndex: 'email',
      copyable: true,
    },
    {
      title: '姓名',
      dataIndex: 'full_name',
      search: false,
    },
    {
      title: '余额',
      dataIndex: 'balance',
      search: false,
      render: (_, record) => `$${record.balance}`,
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      search: false,
      render: (_, record) => dayjs(record.created_at).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => [
        <a key="detail" onClick={() => history.push(`/users/${record.id}`)}>
          详情
        </a>,
        <a key="reset" onClick={() => handleResetPassword(record)}>
          重置密码
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.UserListItem>
        headerTitle="用户管理"
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
