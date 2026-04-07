import { LogoutOutlined, UserOutlined } from '@ant-design/icons';
import type { Settings as LayoutSettings } from '@ant-design/pro-components';
import type { RequestConfig, RunTimeLayoutConfig } from '@umijs/max';
import { history, useModel } from '@umijs/max';
import type { MenuProps } from 'antd';
import { Spin } from 'antd';
import React from 'react';
import { flushSync } from 'react-dom';
import HeaderDropdown from '@/components/HeaderDropdown';
import defaultSettings from '../config/defaultSettings';
import { errorConfig } from './requestErrorConfig';
import { getMe } from '@/services/api';
import '@ant-design/v5-patch-for-react-19';

const loginPath = '/user/login';

export async function getInitialState(): Promise<{
  settings?: Partial<LayoutSettings>;
  currentAdmin?: API.AdminUser;
}> {
  const token = localStorage.getItem('token');
  if (token && history.location.pathname !== loginPath) {
    try {
      const res = await getMe({ skipErrorHandler: true });
      return {
        currentAdmin: res.data,
        settings: defaultSettings as Partial<LayoutSettings>,
      };
    } catch (_error) {
      localStorage.removeItem('token');
      localStorage.removeItem('admin');
      history.push(loginPath);
    }
  }
  return {
    settings: defaultSettings as Partial<LayoutSettings>,
  };
}

const AvatarName: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const { currentAdmin } = initialState || {};
  return <span className="anticon">{currentAdmin?.name}</span>;
};

const AvatarDropdown: React.FC<{ children?: React.ReactNode }> = ({ children }) => {
  const { setInitialState } = useModel('@@initialState');

  const onMenuClick: MenuProps['onClick'] = (event) => {
    const { key } = event;
    if (key === 'logout') {
      flushSync(() => {
        setInitialState((s) => ({ ...s, currentAdmin: undefined }));
      });
      localStorage.removeItem('token');
      localStorage.removeItem('admin');
      history.replace({ pathname: loginPath });
      return;
    }
    if (key === 'profile') {
      history.push('/profile');
    }
  };

  return (
    <HeaderDropdown
      menu={{
        selectedKeys: [],
        onClick: onMenuClick,
        items: [
          { key: 'profile', icon: <UserOutlined />, label: '个人设置' },
          { type: 'divider' as const },
          { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
        ],
      }}
    >
      {children}
    </HeaderDropdown>
  );
};

export const layout: RunTimeLayoutConfig = ({ initialState }) => {
  return {
    avatarProps: {
      icon: <UserOutlined />,
      title: <AvatarName />,
      render: (_: unknown, avatarChildren: React.ReactNode) => {
        return <AvatarDropdown>{avatarChildren}</AvatarDropdown>;
      },
    },
    onPageChange: () => {
      const { location } = history;
      if (!initialState?.currentAdmin && location.pathname !== loginPath) {
        history.push(loginPath);
      }
    },
    menuHeaderRender: undefined,
    ...initialState?.settings,
  };
};

export const request: RequestConfig = {
  baseURL: '/api/v1/admin',
  ...errorConfig,
};
