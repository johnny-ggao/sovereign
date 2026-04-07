import { PageContainer } from '@ant-design/pro-components';
import { Card, Col, Row, Statistic, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';
import { getDashboardStats } from '@/services/api';

const transactionColumns: ColumnsType<API.TransactionInfo> = [
  { title: 'Type', dataIndex: 'type', key: 'type', render: (t: string) => <Tag>{t}</Tag> },
  { title: 'Currency', dataIndex: 'currency', key: 'currency' },
  { title: 'Amount', dataIndex: 'amount', key: 'amount' },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    render: (s: string) => {
      const colorMap: Record<string, string> = {
        completed: 'green',
        pending: 'orange',
        failed: 'red',
      };
      return <Tag color={colorMap[s] ?? 'default'}>{s}</Tag>;
    },
  },
  { title: 'Time', dataIndex: 'created_at', key: 'created_at' },
];

const trendColumns = [
  { title: 'Date', dataIndex: 'date', key: 'date' },
  { title: 'Users', dataIndex: 'count', key: 'count' },
];

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<API.DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getDashboardStats()
      .then((res) => {
        if (res.success && res.data) {
          setStats(res.data);
        }
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="Total Users" value={stats?.total_users ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="New Users Today" value={stats?.new_users_today ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="Total Invested" value={stats?.total_invested ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="Total Deposits" value={stats?.total_deposits ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="Total Withdrawals" value={stats?.total_withdrawals ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="Active Investments" value={stats?.active_investments ?? 0} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="User Trend (Last 7 Days)" loading={loading}>
            <Table
              columns={trendColumns}
              dataSource={stats?.user_trend ?? []}
              rowKey="date"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Recent Transactions" loading={loading}>
            <Table
              columns={transactionColumns}
              dataSource={stats?.recent_transactions ?? []}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Dashboard;
