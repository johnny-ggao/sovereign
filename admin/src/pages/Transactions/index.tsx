import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { Card, Col, Row, Statistic, Tag } from 'antd';
import React, { useEffect, useState } from 'react';
import { getTransactions, getTransactionStats } from '@/services/api';
import dayjs from 'dayjs';

const typeLabels: Record<string, string> = {
  deposit: '充值',
  withdraw: '提现',
};

const statusLabels: Record<string, string> = {
  pending: '等待中',
  processing: '处理中',
  confirmed: '已确认',
  failed: '失败',
};

const statusColors: Record<string, string> = {
  pending: 'orange',
  processing: 'blue',
  confirmed: 'green',
  failed: 'red',
};

const TransactionsPage: React.FC = () => {
  const [stats, setStats] = useState<API.TransactionStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  useEffect(() => {
    getTransactionStats()
      .then((res) => {
        if (res.success && res.data) {
          setStats(res.data);
        }
      })
      .finally(() => setStatsLoading(false));
  }, []);

  const columns: ProColumns<API.TransactionListItem>[] = [
    {
      title: '用户邮箱',
      dataIndex: 'user_email',
      hideInSearch: true,
      render: (_, r) => (
        <a onClick={() => history.push(`/users/${r.user_id}`)}>{r.user_email}</a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      valueType: 'select',
      valueEnum: {
        deposit: { text: '充值' },
        withdraw: { text: '提现' },
      },
      render: (_, r) => (
        <Tag color={r.type === 'deposit' ? 'green' : 'red'}>
          {typeLabels[r.type] ?? r.type}
        </Tag>
      ),
    },
    {
      title: '金额',
      dataIndex: 'amount',
      hideInSearch: true,
    },
    {
      title: '币种',
      dataIndex: 'currency',
      hideInSearch: true,
    },
    {
      title: '网络',
      dataIndex: 'network',
      hideInSearch: true,
    },
    {
      title: '地址',
      dataIndex: 'address',
      hideInSearch: true,
      ellipsis: true,
      width: 160,
    },
    {
      title: '交易哈希',
      dataIndex: 'tx_hash',
      hideInSearch: true,
      ellipsis: true,
      width: 160,
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueType: 'select',
      valueEnum: {
        pending: { text: '等待中' },
        processing: { text: '处理中' },
        confirmed: { text: '已确认' },
        failed: { text: '失败' },
      },
      render: (_, r) => (
        <Tag color={statusColors[r.status] ?? 'default'}>
          {statusLabels[r.status] ?? r.status}
        </Tag>
      ),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      hideInSearch: true,
      render: (_, r) => dayjs(r.created_at).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '搜索',
      dataIndex: 'search',
      hideInTable: true,
      fieldProps: {
        placeholder: '邮箱 / 地址 / 交易哈希',
      },
    },
    {
      title: '日期范围',
      dataIndex: 'dateRange',
      valueType: 'dateRange',
      hideInTable: true,
    },
  ];

  return (
    <PageContainer>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="今日充值"
              value={stats?.deposit_1d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#3f8600' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.deposit_count_1d ?? 0} 笔
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="7日充值"
              value={stats?.deposit_7d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#3f8600' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.deposit_count_7d ?? 0} 笔
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="30日充值"
              value={stats?.deposit_30d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#3f8600' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.deposit_count_30d ?? 0} 笔
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="今日提现"
              value={stats?.withdraw_1d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#cf1322' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.withdraw_count_1d ?? 0} 笔
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="7日提现"
              value={stats?.withdraw_7d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#cf1322' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.withdraw_count_7d ?? 0} 笔
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="30日提现"
              value={stats?.withdraw_30d ?? '0.00'}
              suffix="USDT"
              valueStyle={{ color: '#cf1322' }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              {stats?.withdraw_count_30d ?? 0} 笔
            </div>
          </Card>
        </Col>
      </Row>

      <ProTable<API.TransactionListItem>
        headerTitle="充提币记录"
        columns={columns}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        request={async (params) => {
          const dateRange = params.dateRange;
          const res = await getTransactions({
            page: params.current ?? 1,
            limit: params.pageSize ?? 20,
            type: params.type ?? '',
            search: params.search ?? '',
            status: params.status ?? '',
            date_from: dateRange?.[0] ?? '',
            date_to: dateRange?.[1] ?? '',
          });
          return {
            data: res.data ?? [],
            total: res.meta?.total ?? 0,
            success: res.success,
          };
        }}
        pagination={{ defaultPageSize: 20 }}
      />
    </PageContainer>
  );
};

export default TransactionsPage;
