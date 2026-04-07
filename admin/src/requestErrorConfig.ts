import type { RequestOptions } from '@@/plugin-request/request';
import type { RequestConfig } from '@umijs/max';
import { message } from 'antd';

export const errorConfig: RequestConfig = {
  errorConfig: {
    errorThrower: (res) => {
      const { success, error } = res as unknown as API.ApiResponse<unknown>;
      if (!success) {
        const err: any = new Error(error?.message ?? 'Request failed');
        err.name = 'BizError';
        err.info = error;
        throw err;
      }
    },
    errorHandler: (error: any, opts: any) => {
      if (opts?.skipErrorHandler) throw error;

      if (error.name === 'BizError') {
        const errorInfo = error.info as API.ApiResponse<unknown>['error'];
        if (errorInfo) {
          message.error(errorInfo.message);
        }
      } else if (error.response) {
        if (error.response.status === 401) {
          localStorage.removeItem('token');
          localStorage.removeItem('admin');
          window.location.href = '/user/login';
          return;
        }
        message.error(`Request error: ${error.response.status}`);
      } else if (error.request) {
        message.error('Network error, please try again.');
      } else {
        message.error('Request error, please try again.');
      }
    },
  },

  requestInterceptors: [
    (config: RequestOptions) => {
      const token = localStorage.getItem('token');
      if (token) {
        const headers = {
          ...config.headers,
          Authorization: `Bearer ${token}`,
        };
        return { ...config, headers };
      }
      return config;
    },
  ],

  responseInterceptors: [
    (response) => {
      return response;
    },
  ],
};
