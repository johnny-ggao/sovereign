import { jsx as _jsx } from "react/jsx-runtime";
import { PageContainer, ProTable } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App } from 'antd';
import { useRef } from 'react';
import { getUsers, resetUserPassword } from '@/services/api';
import dayjs from 'dayjs';
const Users = () => {
    const actionRef = useRef();
    const { message, modal } = App.useApp();
    const handleResetPassword = (record) => {
        modal.confirm({
            title: '重置密码',
            content: `确认重置密码：${record.email}？`,
            onOk: async () => {
                try {
                    const res = await resetUserPassword(record.id);
                    if (res.success && res.data) {
                        modal.success({
                            title: '密码已重置',
                            content: `新密码：${res.data.temporary_password}`,
                        });
                    }
                }
                catch (error) {
                    message.error(error?.message ?? '重置密码失败');
                }
            },
        });
    };
    const columns = [
        {
            title: '邮箱',
            dataIndex: 'email',
            copyable: true,
        },
        {
            title: '姓名',
            dataIndex: 'full_name',
            search: false,
        },
        {
            title: '余额',
            dataIndex: 'balance',
            search: false,
            render: (_, record) => `$${record.balance}`,
        },
        {
            title: '注册时间',
            dataIndex: 'created_at',
            search: false,
            render: (_, record) => dayjs(record.created_at).format('YYYY-MM-DD'),
        },
        {
            title: '操作',
            valueType: 'option',
            render: (_, record) => [
                _jsx("a", { onClick: () => history.push(`/users/${record.id}`), children: "\u8BE6\u60C5" }, "detail"),
                _jsx("a", { onClick: () => handleResetPassword(record), children: "\u91CD\u7F6E\u5BC6\u7801" }, "reset"),
            ],
        },
    ];
    return (_jsx(PageContainer, { children: _jsx(ProTable, { headerTitle: "\u7528\u6237\u7BA1\u7406", actionRef: actionRef, rowKey: "id", columns: columns, request: async (params) => {
                const res = await getUsers({
                    page: params.current,
                    per_page: params.pageSize,
                    email: params.email,
                });
                return {
                    data: res.data ?? [],
                    success: res.success,
                    total: res.meta?.total ?? 0,
                };
            }, pagination: { defaultPageSize: 20 } }) }));
};
export default Users;
