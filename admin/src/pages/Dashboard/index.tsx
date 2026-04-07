import { PageContainer } from '@ant-design/pro-components';
import { Card, Col, Row, Statistic, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useState } from 'react';
import { getDashboardStats } from '@/services/api';

const transactionColumns: ColumnsType<API.TransactionInfo> = [
  { title: '类型', dataIndex: 'type', key: 'type', render: (t: string) => <Tag>{t}</Tag> },
  { title: '币种', dataIndex: 'currency', key: 'currency' },
  { title: '金额', dataIndex: 'amount', key: 'amount' },
  {
    title: '状态',
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
  { title: '日期', dataIndex: 'created_at', key: 'created_at' },
];

const trendColumns = [
  { title: '日期', dataIndex: 'date', key: 'date' },
  { title: '用户数', dataIndex: 'count', key: 'count' },
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
            <Statistic title="总用户数" value={stats?.total_users ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="今日新增" value={stats?.new_users_today ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="投资总额" value={stats?.total_invested ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="充值总额" value={stats?.total_deposits ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="提现总额" value={stats?.total_withdrawals ?? '0'} prefix="$" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card loading={loading}>
            <Statistic title="活跃投资" value={stats?.active_investments ?? 0} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="新增用户（近7天）" loading={loading}>
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
          <Card title="近期交易" loading={loading}>
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
