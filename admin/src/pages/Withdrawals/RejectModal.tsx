import { Form, Input, Modal, message } from 'antd';
import React from 'react';
import { rejectWithdrawal } from '@/services/api';

interface Props {
  txId: string | null;
  onClose: () => void;
  onDone: () => void;
}

const RejectModal: React.FC<Props> = ({ txId, onClose, onDone }) => {
  const [form] = Form.useForm();
  const open = !!txId;

  const handleOk = async () => {
    const values = await form.validateFields();
    if (!txId) return;
    try {
      await rejectWithdrawal(txId, values.reason);
      message.success('已拒绝并退回余额');
      form.resetFields();
      onDone();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  return (
    <Modal
      title="拒绝提现"
      open={open}
      onOk={handleOk}
      onCancel={() => {
        form.resetFields();
        onClose();
      }}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="reason"
          label="拒绝原因（将展示给用户）"
          rules={[{ required: true, min: 2, max: 500 }]}
        >
          <Input.TextArea rows={4} placeholder="例：白名单地址不匹配请重新申请" />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default RejectModal;
