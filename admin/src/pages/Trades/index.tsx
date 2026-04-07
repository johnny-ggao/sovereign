import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { Card, Col, Row, Statistic, Tag } from 'antd';
import React, { useEffect, useState } from 'react';
import { getTrades, getTradeStats } from '@/services/api';
import dayjs from 'dayjs';

const pnlColor = (value: string): string =>
  parseFloat(value) >= 0 ? '#3f8600' : '#cf1322';

const formatPnl = (value: string): string =>
  parseFloat(value) >= 0 ? `+${value}` : value;

const TradesPage: React.FC = () => {
  const [stats, setStats] = useState<API.TradeStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  useEffect(() => {
    getTradeStats()
      .then((res) => {
        if (res.success && res.data) {
          setStats(res.data);
        }
      })
      .finally(() => setStatsLoading(false));
  }, []);

  const columns: ProColumns<API.TradeListItem>[] = [
    {
      title: '交易对',
      dataIndex: 'pair',
      copyable: true,
    },
    {
      title: '买入交易所',
      dataIndex: 'buy_exchange',
      hideInSearch: true,
      render: (_, r) => <Tag color="green">{r.buy_exchange}</Tag>,
    },
    {
      title: '卖出交易所',
      dataIndex: 'sell_exchange',
      hideInSearch: true,
      render: (_, r) => <Tag color="red">{r.sell_exchange}</Tag>,
    },
    {
      title: '买入价',
      dataIndex: 'buy_price',
      hideInSearch: true,
    },
    {
      title: '卖出价',
      dataIndex: 'sell_price',
      hideInSearch: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      hideInSearch: true,
    },
    {
      title: '溢价率',
      dataIndex: 'premium_pct',
      hideInSearch: true,
      render: (_, r) => `${r.premium_pct}%`,
    },
    {
      title: '盈亏',
      dataIndex: 'pnl',
      hideInSearch: true,
      render: (_, r) => (
        <span style={{ color: pnlColor(r.pnl), fontWeight: 600 }}>
          {formatPnl(r.pnl)}
        </span>
      ),
    },
    {
      title: '手续费',
      dataIndex: 'fee',
      hideInSearch: true,
    },
    {
      title: '执行时间',
      dataIndex: 'executed_at',
      hideInSearch: true,
      render: (_, r) => dayjs(r.executed_at).format('YYYY-MM-DD HH:mm:ss'),
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
              title="今日盈亏"
              value={stats?.pnl_1d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_1d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_1d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_1d ?? '0'} USDT
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="7日盈亏"
              value={stats?.pnl_7d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_7d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_7d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_7d ?? '0'} USDT
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="30日盈亏"
              value={stats?.pnl_30d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_30d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_30d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_30d ?? '0'} USDT
            </div>
          </Card>
        </Col>
      </Row>

      <ProTable<API.TradeListItem>
        headerTitle="套利交易历史"
        columns={columns}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        request={async (params) => {
          const dateRange = params.dateRange;
          const res = await getTrades({
            page: params.current ?? 1,
            limit: params.pageSize ?? 20,
            pair: params.pair ?? '',
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

export default TradesPage;
