import { LockOutlined } from '@ant-design/icons';
import { ProFormText } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App, Button, Card, Form } from 'antd';
import React, { useState } from 'react';
import { changePassword } from '@/services/api';

const ChangePasswordPage: React.FC = () => {
  const [form] = Form.useForm();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (values.new_password !== values.confirm_password) {
        message.error('两次输入的密码不一致');
        return;
      }
      setLoading(true);
      await changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      localStorage.setItem('must_change_password', 'false');
      message.success('密码修改成功');
      history.push('/dashboard');
    } catch (error: any) {
      const errMsg =
        error?.info?.message ?? error?.message ?? '密码修改失败';
      message.error(errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        background: '#f0f2f5',
      }}
    >
      <Card title="首次登录请修改密码" style={{ width: 400 }}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <ProFormText.Password
            name="old_password"
            label="当前密码"
            fieldProps={{ size: 'large', prefix: <LockOutlined /> }}
            rules={[{ required: true, message: '请输入当前密码' }]}
          />
          <ProFormText.Password
            name="new_password"
            label="新密码"
            fieldProps={{ size: 'large', prefix: <LockOutlined /> }}
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6位' },
            ]}
          />
          <ProFormText.Password
            name="confirm_password"
            label="确认新密码"
            fieldProps={{ size: 'large', prefix: <LockOutlined /> }}
            rules={[
              { required: true, message: '请确认新密码' },
              { min: 6, message: '密码至少6位' },
            ]}
          />
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            size="large"
          >
            确认修改
          </Button>
        </Form>
      </Card>
    </div>
  );
};

export default ChangePasswordPage;
