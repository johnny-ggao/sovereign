import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { LoginForm, ProFormText } from '@ant-design/pro-components';
import { Helmet, useModel } from '@umijs/max';
import { Alert, App } from 'antd';
import { createStyles } from 'antd-style';
import React, { useState } from 'react';
import { flushSync } from 'react-dom';
import { login } from '@/services/api';

const useStyles = createStyles(({ token }) => ({
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100vh',
    overflow: 'auto',
    backgroundImage:
      "url('https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/V-_oS6r-i7wAAAAAAAAAAAAAFl94AQBr')",
    backgroundSize: '100% 100%',
  },
}));

const LoginMessage: React.FC<{ content: string }> = ({ content }) => (
  <Alert style={{ marginBottom: 24 }} message={content} type="error" showIcon />
);

const LoginPage: React.FC = () => {
  const [errorMessage, setErrorMessage] = useState<string>('');
  const { setInitialState } = useModel('@@initialState');
  const { styles } = useStyles();
  const { message } = App.useApp();

  const handleSubmit = async (values: { email: string; password: string }) => {
    try {
      setErrorMessage('');
      const res = await login(values);
      if (res.success && res.data) {
        localStorage.setItem('token', res.data.token);
        localStorage.setItem('admin', JSON.stringify(res.data.admin));
        localStorage.setItem(
          'must_change_password',
          String(res.data.must_change_password),
        );
        flushSync(() => {
          setInitialState((s) => ({
            ...s,
            currentAdmin: res.data!.admin,
          }));
        });
        if (res.data.must_change_password) {
          message.warning('首次登录请修改密码');
          window.location.href = '/change-password';
          return;
        }
        message.success('登录成功');
        const urlParams = new URL(window.location.href).searchParams;
        window.location.href = urlParams.get('redirect') || '/dashboard';
        return;
      }
      setErrorMessage(res.error?.message ?? 'Login failed');
    } catch (error: any) {
      const errMsg =
        error?.info?.message ?? error?.message ?? '登录失败，请重试';
      setErrorMessage(errMsg);
    }
  };

  return (
    <div className={styles.container}>
      <Helmet>
        <title>登录 - Sovereign Admin</title>
      </Helmet>
      <div style={{ flex: '1', padding: '32px 0' }}>
        <LoginForm
          contentStyle={{ minWidth: 280, maxWidth: '75vw' }}
          logo={<img alt="logo" src="/logo.svg" />}
          title="Sovereign Admin"
          subTitle="后台管理系统"
          onFinish={handleSubmit}
        >
          {errorMessage && <LoginMessage content={errorMessage} />}
          <ProFormText
            name="email"
            fieldProps={{
              size: 'large',
              prefix: <UserOutlined />,
            }}
            placeholder="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱' },
            ]}
          />
          <ProFormText.Password
            name="password"
            fieldProps={{
              size: 'large',
              prefix: <LockOutlined />,
            }}
            placeholder="密码"
            rules={[{ required: true, message: '请输入密码' }]}
          />
        </LoginForm>
      </div>
    </div>
  );
};

export default LoginPage;
