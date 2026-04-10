import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { Tag } from 'antd';
import React, { useRef } from 'react';
import dayjs from 'dayjs';
import { getAuditLogs } from '@/services/api';

const actionMap: Record<string, { text: string; color: string }> = {
  adjust_balance: { text: '调整余额', color: 'gold' },
  disable_user: { text: '禁用用户', color: 'red' },
  enable_user: { text: '启用用户', color: 'green' },
  reset_password: { text: '重置密码', color: 'orange' },
  reset_2fa: { text: '重置2FA', color: 'cyan' },
  delete_trade: { text: '删除交易', color: 'volcano' },
  create_admin: { text: '创建管理员', color: 'blue' },
  delete_admin: { text: '删除管理员', color: 'magenta' },
};

const targetTypeMap: Record<string, string> = {
  user: '用户',
  trade: '交易',
  admin: '管理员',
};

const AuditLogsPage: React.FC = () => {
  const actionRef = useRef<ActionType>();

  const columns: ProColumns<API.AuditLog>[] = [
    {
      title: '时间',
      dataIndex: 'created_at',
      valueType: 'dateTime',
      search: false,
      width: 180,
      render: (_, record) => dayjs(record.created_at).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作人',
      dataIndex: 'admin_email',
      ellipsis: true,
      copyable: true,
    },
    {
      title: '操作类型',
      dataIndex: 'action',
      valueType: 'select',
      valueEnum: Object.fromEntries(
        Object.entries(actionMap).map(([value, item]) => [value, { text: item.text }]),
      ),
      render: (_, record) => {
        const item = actionMap[record.action] ?? { text: record.action, color: 'default' };
        return <Tag color={item.color}>{item.text}</Tag>;
      },
    },
    {
      title: '目标',
      dataIndex: 'target_type',
      search: false,
      render: (_, record) => targetTypeMap[record.target_type] ?? record.target_type,
    },
    {
      title: '目标ID',
      dataIndex: 'target_id',
      search: false,
      ellipsis: true,
      copyable: true,
    },
    {
      title: '详情',
      dataIndex: 'detail',
      search: false,
      ellipsis: true,
    },
    {
      title: 'IP',
      dataIndex: 'ip_address',
      search: false,
      width: 150,
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
      <ProTable<API.AuditLog>
        headerTitle="管理员审计日志"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        search={{ labelWidth: 'auto' }}
        request={async (params) => {
          const dateRange = params.dateRange as string[] | undefined;
          const res = await getAuditLogs({
            page: params.current ?? 1,
            limit: params.pageSize ?? 20,
            action: params.action as string | undefined,
            admin_id: params.admin_id as string | undefined,
            date_from: dateRange?.[0] ? dayjs(dateRange[0]).format('YYYY-MM-DD') : '',
            date_to: dateRange?.[1] ? dayjs(dateRange[1]).format('YYYY-MM-DD') : '',
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

export default AuditLogsPage;
