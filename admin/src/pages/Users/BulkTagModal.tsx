import { Modal, Form, Select, Input, App } from 'antd';
import { useState } from 'react';
import { bulkChangeUserInvestmentType } from '@/services/api';

interface Props {
  open: boolean;
  selectedUserIds: string[];
  onClose: () => void;
  onDone: () => void;
}

export default function BulkTagModal({ open, selectedUserIds, onClose, onDone }: Props) {
  const { message } = App.useApp();
  const [form] = Form.useForm<{ type: 'arbitrage' | 'trading'; reason: string }>();
  const [submitting, setSubmitting] = useState(false);

  const handleOk = async () => {
    const values = await form.validateFields();
    setSubmitting(true);
    try {
      const res = await bulkChangeUserInvestmentType(selectedUserIds, values.type, values.reason);
      const data = (res as any)?.data ?? res;
      const succeeded = data?.succeeded ?? [];
      const failed = data?.failed ?? [];
      if (failed.length === 0) {
        message.success(`已成功打标 ${succeeded.length} 个用户`);
      } else {
        message.warning(`成功 ${succeeded.length}，失败 ${failed.length}（详见日志）`);
        console.warn('bulk-tag failures:', failed);
      }
      form.resetFields();
      onDone();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? e?.message ?? '操作失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={`批量打标 (${selectedUserIds.length} 个用户)`}
      open={open}
      onOk={handleOk}
      onCancel={() => { form.resetFields(); onClose(); }}
      confirmLoading={submitting}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item name="type" label="目标投资类型" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 'arbitrage', label: '套利投资' },
              { value: 'trading', label: '交易投资' },
            ]}
          />
        </Form.Item>
        <Form.Item
          name="reason"
          label="变更原因"
          rules={[
            { required: true, message: '请填写变更原因' },
            { min: 5, max: 500, message: '变更原因需 5-500 字符' },
          ]}
        >
          <Input.TextArea rows={3} placeholder="例：批次拉入交易投资内测" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
