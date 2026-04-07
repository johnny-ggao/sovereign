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
        flushSync(() => {
          setInitialState((s) => ({
            ...s,
            currentAdmin: res.data!.admin,
          }));
        });
        message.success('Login successful');
        const urlParams = new URL(window.location.href).searchParams;
        window.location.href = urlParams.get('redirect') || '/dashboard';
        return;
      }
      setErrorMessage(res.error?.message ?? 'Login failed');
    } catch (error: any) {
      const errMsg = error?.info?.message ?? error?.message ?? 'Login failed, please try again.';
      setErrorMessage(errMsg);
    }
  };

  return (
    <div className={styles.container}>
      <Helmet>
        <title>Login - Sovereign Admin</title>
      </Helmet>
      <div style={{ flex: '1', padding: '32px 0' }}>
        <LoginForm
          contentStyle={{ minWidth: 280, maxWidth: '75vw' }}
          logo={<img alt="logo" src="/logo.svg" />}
          title="Sovereign Admin"
          subTitle="Admin Management Portal"
          onFinish={handleSubmit}
        >
          {errorMessage && <LoginMessage content={errorMessage} />}
          <ProFormText
            name="email"
            fieldProps={{
              size: 'large',
              prefix: <UserOutlined />,
            }}
            placeholder="Email"
            rules={[
              { required: true, message: 'Please enter your email' },
              { type: 'email', message: 'Please enter a valid email' },
            ]}
          />
          <ProFormText.Password
            name="password"
            fieldProps={{
              size: 'large',
              prefix: <LockOutlined />,
            }}
            placeholder="Password"
            rules={[{ required: true, message: 'Please enter your password' }]}
          />
        </LoginForm>
      </div>
    </div>
  );
};

export default LoginPage;
