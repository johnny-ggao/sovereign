import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { Tag } from 'antd';
import React from 'react';
import { getInvestments } from '@/services/api';
import dayjs from 'dayjs';

const statusColors: Record<string, string> = {
  active: 'blue',
  stopping: 'orange',
  redeemed: 'default',
};

const statusLabels: Record<string, string> = {
  active: '运行中',
  stopping: '停止中',
  redeemed: '已赎回',
};

const columns: ProColumns<API.InvestmentListItem>[] = [
  {
    title: '用户邮箱',
    dataIndex: 'user_email',
    render: (_, record) => (
      <a onClick={() => history.push(`/users/${record.user_id}`)}>{record.user_email}</a>
    ),
    hideInSearch: true,
  },
  {
    title: '搜索',
    dataIndex: 'search',
    hideInTable: true,
    fieldProps: { placeholder: '邮箱 / 用户ID' },
  },
  {
    title: '投资金额',
    dataIndex: 'amount',
    hideInSearch: true,
    sorter: true,
    render: (_, r) => `${r.amount} ${r.currency}`,
  },
  {
    title: '状态',
    dataIndex: 'status',
    valueType: 'select',
    valueEnum: {
      active: { text: '运行中', status: 'Processing' },
      stopping: { text: '停止中', status: 'Warning' },
      redeemed: { text: '已赎回', status: 'Default' },
    },
    render: (_, r) => (
      <Tag color={statusColors[r.status] ?? 'default'}>{statusLabels[r.status] ?? r.status}</Tag>
    ),
  },
  {
    title: '净收益',
    dataIndex: 'net_return',
    hideInSearch: true,
  },
  {
    title: '开始日期',
    dataIndex: 'start_date',
    hideInSearch: true,
    render: (_, r) => dayjs(r.start_date).format('YYYY-MM-DD'),
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    hideInSearch: true,
    render: (_, r) => dayjs(r.created_at).format('YYYY-MM-DD'),
  },
];

const InvestmentsPage: React.FC = () => {
  return (
    <PageContainer>
      <ProTable<API.InvestmentListItem>
        headerTitle="投资管理"
        columns={columns}
        rowKey="id"
        request={async (params, sort) => {
          const sortBy = Object.keys(sort ?? {})[0];
          const sortOrder = sortBy
            ? sort[sortBy] === 'ascend'
              ? 'asc'
              : 'desc'
            : undefined;
          const res = await getInvestments({
            page: params.current ?? 1,
            limit: params.pageSize ?? 20,
            search: params.search ?? '',
            status: params.status ?? '',
            sort_by: sortBy ?? 'created_at',
            sort_order: sortOrder ?? 'desc',
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

export default InvestmentsPage;
