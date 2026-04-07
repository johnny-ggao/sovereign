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
  { title: '币种', dataIndex: 'currency', key: 'currency' },
  { title: '可用', dataIndex: 'available', key: 'available' },
  { title: '运营中', dataIndex: 'in_operation', key: 'in_operation' },
  { title: '冻结', dataIndex: 'frozen', key: 'frozen' },
  { title: '收益', dataIndex: 'earnings', key: 'earnings' },
  { title: '合计', dataIndex: 'total', key: 'total' },
];

const transactionColumns: ColumnsType<API.TransactionInfo> = [
  { title: '类型', dataIndex: 'type', key: 'type', render: (t: string) => <Tag>{t}</Tag> },
  { title: '币种', dataIndex: 'currency', key: 'currency' },
  { title: '网络', dataIndex: 'network', key: 'network' },
  { title: '金额', dataIndex: 'amount', key: 'amount' },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (s: string) => {
      const colorMap: Record<string, string> = { completed: 'green', pending: 'orange', failed: 'red' };
      return <Tag color={colorMap[s] ?? 'default'}>{s}</Tag>;
    },
  },
  { title: '交易哈希', dataIndex: 'tx_hash', key: 'tx_hash', ellipsis: true },
  { title: '日期', dataIndex: 'created_at', key: 'created_at' },
];

const investmentColumns: ColumnsType<API.InvestmentInfo> = [
  { title: '金额', dataIndex: 'amount', key: 'amount' },
  { title: '币种', dataIndex: 'currency', key: 'currency' },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (s: string) => {
      const colorMap: Record<string, string> = { active: 'green', settled: 'blue', cancelled: 'red' };
      return <Tag color={colorMap[s] ?? 'default'}>{s}</Tag>;
    },
  },
  { title: '净收益', dataIndex: 'net_return', key: 'net_return' },
  { title: '开始日期', dataIndex: 'start_date', key: 'start_date' },
];

const settlementColumns: ColumnsType<API.SettlementInfo> = [
  { title: '周期', dataIndex: 'period', key: 'period' },
  { title: '净收益', dataIndex: 'net_return', key: 'net_return' },
  { title: '费率', dataIndex: 'fee_rate', key: 'fee_rate' },
  { title: '结算时间', dataIndex: 'settled_at', key: 'settled_at' },
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
      title: '重置密码',
      content: `确认重置密码：${user?.email}？`,
      onOk: async () => {
        try {
          const res = await resetUserPassword(id);
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

  const handleAdjustBalance = async () => {
    if (!id) return;
    setAdjustLoading(true);
    try {
      const res = await adjustBalance(id, { amount: adjustAmount, reason: adjustReason });
      if (res.success) {
        message.success('余额已调整');
        setAdjustModalOpen(false);
        setAdjustAmount('');
        setAdjustReason('');
        fetchUser();
      }
    } catch (error: any) {
      message.error(error?.message ?? '调整余额失败');
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
            <Button onClick={handleResetPassword}>重置密码</Button>
            <Button type="primary" onClick={() => setAdjustModalOpen(true)}>
              调整余额
            </Button>
          </Space>
        ) : null
      }
    >
      {user && (
        <>
          <Card style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, lg: 3 }}>
              <Descriptions.Item label="邮箱">{user.email}</Descriptions.Item>
              <Descriptions.Item label="姓名">{user.full_name}</Descriptions.Item>
              <Descriptions.Item label="电话">{user.phone}</Descriptions.Item>
              <Descriptions.Item label="语言">{user.language}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={user.is_active ? 'green' : 'red'}>
                  {user.is_active ? '活跃' : '停用'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="注册时间">
                {dayjs(user.created_at).format('YYYY-MM-DD HH:mm')}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Tabs
            items={[
              {
                key: 'wallets',
                label: '钱包',
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
                label: '交易记录',
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
                label: '投资记录',
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
                label: '结算记录',
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
        title="调整余额 (USDT)"
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
          <label>金额（如 100 或 -50）：</label>
          <Input
            value={adjustAmount}
            onChange={(e) => setAdjustAmount(e.target.value)}
            placeholder="如 100 或 -50"
            style={{ marginTop: 8 }}
          />
        </div>
        <div>
          <label>原因：</label>
          <Input.TextArea
            value={adjustReason}
            onChange={(e) => setAdjustReason(e.target.value)}
            placeholder="调整原因"
            style={{ marginTop: 8 }}
            rows={3}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default UserDetailPage;
