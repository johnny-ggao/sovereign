import { PageContainer } from '@ant-design/pro-components';
import { history, useParams, useAccess } from '@umijs/max';
import {
  App,
  Button,
  Card,
  Descriptions,
  Input,
  Modal,
  Space,
  Table,
  Tabs,
  Tag,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';
import dayjs from 'dayjs';
import {
  getUserDetail,
  resetUserPassword,
  adjustBalance,
} from '@/services/api';

const walletColumns: ColumnsType<API.WalletInfo> = [
  { title: 'Currency', dataIndex: 'currency', key: 'currency' },
  { title: 'Available', dataIndex: 'available', key: 'available' },
  { title: 'In Operation', dataIndex: 'in_operation', key: 'in_operation' },
  { title: 'Frozen', dataIndex: 'frozen', key: 'frozen' },
  { title: 'Earnings', dataIndex: 'earnings', key: 'earnings' },
  { title: 'Total', dataIndex: 'total', key: 'total' },
];

const transactionColumns: ColumnsType<API.TransactionInfo> = [
  { title: 'Type', dataIndex: 'type', key: 'type', render: (t: string) => <Tag>{t}</Tag> },
  { title: 'Currency', dataIndex: 'currency', key: 'currency' },
  { title: 'Network', dataIndex: 'network', key: 'network' },
  { title: 'Amount', dataIndex: 'amount', key: 'amount' },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    render: (s: string) => {
      const colorMap: Record<string, string> = { completed: 'green', pending: 'orange', failed: 'red' };
      return <Tag color={colorMap[s] ?? 'default'}>{s}</Tag>;
    },
  },
  { title: 'Tx Hash', dataIndex: 'tx_hash', key: 'tx_hash', ellipsis: true },
  { title: 'Time', dataIndex: 'created_at', key: 'created_at' },
];

const investmentColumns: ColumnsType<API.InvestmentInfo> = [
  { title: 'Amount', dataIndex: 'amount', key: 'amount' },
  { title: 'Currency', dataIndex: 'currency', key: 'currency' },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    render: (s: string) => {
      const colorMap: Record<string, string> = { active: 'green', settled: 'blue', cancelled: 'red' };
      return <Tag color={colorMap[s] ?? 'default'}>{s}</Tag>;
    },
  },
  { title: 'Net Return', dataIndex: 'net_return', key: 'net_return' },
  { title: 'Start Date', dataIndex: 'start_date', key: 'start_date' },
];

const settlementColumns: ColumnsType<API.SettlementInfo> = [
  { title: 'Period', dataIndex: 'period', key: 'period' },
  { title: 'Net Return', dataIndex: 'net_return', key: 'net_return' },
  { title: 'Fee Rate', dataIndex: 'fee_rate', key: 'fee_rate' },
  { title: 'Settled At', dataIndex: 'settled_at', key: 'settled_at' },
];

const UserDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const access = useAccess();
  const { message, modal } = App.useApp();
  const [user, setUser] = useState<API.UserDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [adjustModalOpen, setAdjustModalOpen] = useState(false);
  const [adjustAmount, setAdjustAmount] = useState('');
  const [adjustReason, setAdjustReason] = useState('');
  const [adjustLoading, setAdjustLoading] = useState(false);

  const fetchUser = () => {
    if (!id) return;
    setLoading(true);
    getUserDetail(id)
      .then((res) => {
        if (res.success && res.data) {
          setUser(res.data);
        }
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchUser();
  }, [id]);

  const handleResetPassword = () => {
    if (!id) return;
    modal.confirm({
      title: 'Reset Password',
      content: `Are you sure you want to reset the password for ${user?.email}?`,
      onOk: async () => {
        try {
          const res = await resetUserPassword(id);
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

  const handleAdjustBalance = async () => {
    if (!id) return;
    setAdjustLoading(true);
    try {
      const res = await adjustBalance(id, { amount: adjustAmount, reason: adjustReason });
      if (res.success) {
        message.success('Balance adjusted successfully');
        setAdjustModalOpen(false);
        setAdjustAmount('');
        setAdjustReason('');
        fetchUser();
      }
    } catch (error: any) {
      message.error(error?.message ?? 'Failed to adjust balance');
    } finally {
      setAdjustLoading(false);
    }
  };

  return (
    <PageContainer
      loading={loading}
      onBack={() => history.push('/users')}
      extra={
        access.isOperator ? (
          <Space>
            <Button onClick={handleResetPassword}>Reset Password</Button>
            <Button type="primary" onClick={() => setAdjustModalOpen(true)}>
              Adjust Balance
            </Button>
          </Space>
        ) : null
      }
    >
      {user && (
        <>
          <Card style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, lg: 3 }}>
              <Descriptions.Item label="Email">{user.email}</Descriptions.Item>
              <Descriptions.Item label="Name">{user.full_name}</Descriptions.Item>
              <Descriptions.Item label="Phone">{user.phone}</Descriptions.Item>
              <Descriptions.Item label="Language">{user.language}</Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={user.is_active ? 'green' : 'red'}>
                  {user.is_active ? 'Active' : 'Inactive'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Joined">
                {dayjs(user.created_at).format('YYYY-MM-DD HH:mm')}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Tabs
            items={[
              {
                key: 'wallets',
                label: 'Wallets',
                children: (
                  <Table
                    columns={walletColumns}
                    dataSource={user.wallets}
                    rowKey="currency"
                    pagination={false}
                  />
                ),
              },
              {
                key: 'transactions',
                label: 'Transactions',
                children: (
                  <Table
                    columns={transactionColumns}
                    dataSource={user.recent_transactions}
                    rowKey="id"
                    pagination={{ pageSize: 10 }}
                  />
                ),
              },
              {
                key: 'investments',
                label: 'Investments',
                children: (
                  <Table
                    columns={investmentColumns}
                    dataSource={user.investments}
                    rowKey="id"
                    pagination={{ pageSize: 10 }}
                  />
                ),
              },
              {
                key: 'settlements',
                label: 'Settlements',
                children: (
                  <Table
                    columns={settlementColumns}
                    dataSource={user.recent_settlements}
                    rowKey="id"
                    pagination={{ pageSize: 10 }}
                  />
                ),
              },
            ]}
          />
        </>
      )}

      <Modal
        title="Adjust Balance"
        open={adjustModalOpen}
        onOk={handleAdjustBalance}
        onCancel={() => {
          setAdjustModalOpen(false);
          setAdjustAmount('');
          setAdjustReason('');
        }}
        confirmLoading={adjustLoading}
        okButtonProps={{ disabled: !adjustAmount || !adjustReason }}
      >
        <div style={{ marginBottom: 16 }}>
          <label>Amount (use negative for deduction):</label>
          <Input
            value={adjustAmount}
            onChange={(e) => setAdjustAmount(e.target.value)}
            placeholder="e.g. 100 or -50"
            style={{ marginTop: 8 }}
          />
        </div>
        <div>
          <label>Reason:</label>
          <Input.TextArea
            value={adjustReason}
            onChange={(e) => setAdjustReason(e.target.value)}
            placeholder="Reason for adjustment"
            style={{ marginTop: 8 }}
            rows={3}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default UserDetailPage;
