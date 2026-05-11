import { Card, Tag, Button, Modal, Form, Input, Timeline, App } from 'antd';
import { useState } from 'react';
import { useRequest } from '@umijs/max';
import {
  changeUserInvestmentType,
  getUserProductChanges,
  type UserProductChangeLog,
} from '@/services/api';

interface Props {
  userId: string;
  currentType: string;
  canEdit?: boolean;
  onChanged: () => void;
}

const LABEL: Record<string, { text: string; color: string }> = {
  arbitrage: { text: '套利投资', color: 'blue' },
  trading: { text: '交易投资', color: 'gold' },
};

export default function InvestmentTypeCard({ userId, currentType, canEdit = true, onChanged }: Props) {
  const { message } = App.useApp();
  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm<{ reason: string }>();

  const { data: history, refresh: refreshHistory } = useRequest(
    async () => {
      const res = await getUserProductChanges(userId);
      return res.data ?? [];
    },
    { refreshDeps: [userId] },
  );

  const normalizedCurrent = currentType || 'arbitrage';
  const target: 'arbitrage' | 'trading' =
    normalizedCurrent === 'trading' ? 'arbitrage' : 'trading';
  const meta = LABEL[normalizedCurrent] ?? { text: normalizedCurrent, color: 'default' };

  const handleOk = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      await changeUserInvestmentType(userId, target, values.reason);
      message.success('已更新');
      form.resetFields();
      setModalOpen(false);
      onChanged();
      refreshHistory();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? e?.message ?? '操作失败');
    } finally {
      setSubmitting(false);
    }
  };

  const logs: UserProductChangeLog[] = (history as UserProductChangeLog[] | undefined) ?? [];
  const items = logs.map((l) => ({
    children: (
      <div>
        <div>
          <Tag color={LABEL[l.from_type]?.color}>
            {LABEL[l.from_type]?.text ?? l.from_type}
          </Tag>
          →
          <Tag color={LABEL[l.to_type]?.color}>
            {LABEL[l.to_type]?.text ?? l.to_type}
          </Tag>
        </div>
        <div style={{ color: '#888', fontSize: 12 }}>
          {l.admin_email} · {new Date(l.created_at).toLocaleString()}
        </div>
        <div style={{ marginTop: 4 }}>原因：{l.reason}</div>
      </div>
    ),
  }));

  return (
    <Card
      title="投资产品类型"
      style={{ marginBottom: 16 }}
      extra={canEdit ? (
        <Button type="primary" onClick={() => setModalOpen(true)}>
          切换为 {LABEL[target].text}
        </Button>
      ) : null}
    >
      <div style={{ marginBottom: 16 }}>
        当前类型：<Tag color={meta.color}>{meta.text}</Tag>
      </div>
      <h4 style={{ marginTop: 16 }}>变更历史</h4>
      {items.length > 0 ? (
        <Timeline items={items} />
      ) : (
        <span style={{ color: '#888' }}>暂无变更记录</span>
      )}

      <Modal
        title={`切换为 ${LABEL[target].text}`}
        open={modalOpen}
        onOk={handleOk}
        onCancel={() => {
          form.resetFields();
          setModalOpen(false);
        }}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="reason"
            label="变更原因"
            rules={[
              { required: true, message: '请填写变更原因' },
              { min: 5, max: 500, message: '变更原因需 5-500 字符' },
            ]}
          >
            <Input.TextArea rows={3} placeholder="例：参与交易投资内测" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
