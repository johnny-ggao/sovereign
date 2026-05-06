import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { Button, Popconfirm, Space, Tabs, Tag, Tooltip, message } from 'antd';
import React, { useRef, useState } from 'react';
import {
  approveWithdrawal,
  listWithdrawals,
  retryWithdrawal,
  type WithdrawReviewItem,
} from '@/services/api';
import RejectModal from './RejectModal';

const STATUS_LABEL: Record<string, { color: string; text: string }> = {
  pending_review: { color: 'gold', text: '待审核' },
  submit_failed: { color: 'red', text: '提交失败' },
  submitted: { color: 'blue', text: '已提交' },
  rejected: { color: 'default', text: '已拒绝' },
  cancelled: { color: 'default', text: '已取消' },
};

const WithdrawalsPage: React.FC = () => {
  const [tab, setTab] = useState<string>('pending_review');
  const [rejectId, setRejectId] = useState<string | null>(null);
  const actionRef = useRef<ActionType>();

  const onApprove = async (id: string) => {
    try {
      await approveWithdrawal(id);
      message.success('已批准并提交至 Cobo');
      actionRef.current?.reload();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  const onRetry = async (id: string) => {
    try {
      await retryWithdrawal(id);
      message.success('已重新提交');
      actionRef.current?.reload();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '重试失败');
    }
  };

  const columns: ProColumns<WithdrawReviewItem>[] = [
    { title: '用户', dataIndex: 'user_email', width: 200, ellipsis: true },
    { title: '币种', dataIndex: 'currency', width: 80 },
    { title: '网络', dataIndex: 'network', width: 80 },
    { title: '金额', dataIndex: 'amount', width: 120 },
    {
      title: '目标地址',
      dataIndex: 'address',
      ellipsis: true,
      copyable: true,
    },
    {
      title: '状态',
      dataIndex: 'review_status',
      width: 100,
      render: (_, r) => {
        const meta = STATUS_LABEL[r.review_status] || { color: 'default', text: r.review_status };
        return <Tag color={meta.color}>{meta.text}</Tag>;
      },
    },
    {
      title: '提交尝试',
      dataIndex: 'submit_attempts',
      width: 100,
      render: (v, r) =>
        r.last_submit_error ? (
          <Tooltip title={r.last_submit_error}>
            <span>{v as number}</span>
          </Tooltip>
        ) : (
          (v as number)
        ),
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      valueType: 'dateTime',
      width: 170,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 220,
      fixed: 'right',
      render: (_, r) => (
        <Space>
          {r.review_status === 'pending_review' && (
            <>
              <Popconfirm
                title={`确定批准并提交 ${r.amount} ${r.currency}？`}
                onConfirm={() => onApprove(r.id)}
              >
                <Button size="small" type="primary">
                  批准
                </Button>
              </Popconfirm>
              <Button size="small" danger onClick={() => setRejectId(r.id)}>
                拒绝
              </Button>
            </>
          )}
          {r.review_status === 'submit_failed' && (
            <>
              <Popconfirm title="确认重新提交至 Cobo？" onConfirm={() => onRetry(r.id)}>
                <Button size="small" type="primary">
                  重试
                </Button>
              </Popconfirm>
              <Button size="small" danger onClick={() => setRejectId(r.id)}>
                拒绝退款
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <Tabs
        activeKey={tab}
        onChange={(k) => {
          setTab(k);
          actionRef.current?.reload();
        }}
        items={[
          { key: 'pending_review', label: '待审核' },
          { key: 'submit_failed', label: '提交失败' },
          { key: '', label: '历史' },
        ]}
      />
      <ProTable<WithdrawReviewItem>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        request={async ({ current, pageSize }) => {
          const res = await listWithdrawals({
            review_status: tab || undefined,
            page: current,
            limit: pageSize,
          });
          return {
            data: res.data ?? [],
            total: res.meta?.total ?? 0,
            success: res.success,
          };
        }}
        columns={columns}
        scroll={{ x: 1200 }}
        pagination={{ defaultPageSize: 20 }}
      />
      <RejectModal
        txId={rejectId}
        onClose={() => setRejectId(null)}
        onDone={() => {
          setRejectId(null);
          actionRef.current?.reload();
        }}
      />
    </PageContainer>
  );
};

export default WithdrawalsPage;
