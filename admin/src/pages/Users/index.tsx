import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App, Button, Tag, Typography } from 'antd';
import React, { useRef, useState } from 'react';
import { getUsers, resetUserPassword } from '@/services/api';
import dayjs from 'dayjs';
import BulkTagModal from './BulkTagModal';

const Users: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const { message, modal } = App.useApp();
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [bulkOpen, setBulkOpen] = useState(false);

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
      title: '投资类型',
      dataIndex: 'investment_type',
      width: 100,
      valueEnum: {
        arbitrage: { text: '套利', status: 'Default' },
        trading: { text: '交易', status: 'Processing' },
      },
      render: (_, record) => {
        const t = (record as any).investment_type || 'arbitrage';
        if (t === 'trading') return <Tag color="gold">交易</Tag>;
        return <Tag color="blue">套利</Tag>;
      },
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
        rowSelection={{
          selectedRowKeys: selectedKeys,
          onChange: (keys) => setSelectedKeys(keys as string[]),
          preserveSelectedRowKeys: false,
        }}
        toolBarRender={() => [
          <Button
            key="bulk-tag"
            type="primary"
            disabled={selectedKeys.length === 0}
            onClick={() => setBulkOpen(true)}
          >
            批量打标 {selectedKeys.length > 0 ? `(${selectedKeys.length})` : ''}
          </Button>,
        ]}
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
      <BulkTagModal
        open={bulkOpen}
        selectedUserIds={selectedKeys}
        onClose={() => setBulkOpen(false)}
        onDone={() => {
          setBulkOpen(false);
          setSelectedKeys([]);
          actionRef.current?.reload();
        }}
      />
    </PageContainer>
  );
};

export default Users;
