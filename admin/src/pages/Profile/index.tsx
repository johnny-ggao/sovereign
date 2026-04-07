import { PageContainer } from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { App, Button, Card, Descriptions, Input, Modal, Tag } from 'antd';
import React, { useState } from 'react';
import dayjs from 'dayjs';
import { changePassword } from '@/services/api';

const Profile: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const { message } = App.useApp();
  const admin = initialState?.currentAdmin;

  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleChangePassword = async () => {
    if (!oldPassword || !newPassword) {
      message.error('请填写所有字段');
      return;
    }
    setLoading(true);
    try {
      const res = await changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });
      if (res.success) {
        message.success('密码已修改');
        setPasswordModalOpen(false);
        setOldPassword('');
        setNewPassword('');
      }
    } catch (error: any) {
      message.error(error?.message ?? '密码修改失败');
    } finally {
      setLoading(false);
    }
  };

  if (!admin) {
    return null;
  }

  const roleColorMap: Record<string, string> = {
    super_admin: 'red',
    operator: 'blue',
    viewer: 'green',
  };

  return (
    <PageContainer>
      <Card
        extra={
          <Button type="primary" onClick={() => setPasswordModalOpen(true)}>
            修改密码
          </Button>
        }
      >
        <Descriptions column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="姓名">{admin.name}</Descriptions.Item>
          <Descriptions.Item label="邮箱">{admin.email}</Descriptions.Item>
          <Descriptions.Item label="角色">
            <Tag color={roleColorMap[admin.role] ?? 'default'}>{admin.role}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="最后登录">
            {admin.last_login
              ? dayjs(admin.last_login).format('YYYY-MM-DD HH:mm')
              : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Modal
        title="修改密码"
        open={passwordModalOpen}
        onOk={handleChangePassword}
        onCancel={() => {
          setPasswordModalOpen(false);
          setOldPassword('');
          setNewPassword('');
        }}
        confirmLoading={loading}
        okButtonProps={{ disabled: !oldPassword || !newPassword }}
      >
        <div style={{ marginBottom: 16 }}>
          <label>旧密码：</label>
          <Input.Password
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            placeholder="请输入旧密码"
            style={{ marginTop: 8 }}
          />
        </div>
        <div>
          <label>新密码：</label>
          <Input.Password
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="请输入新密码"
            style={{ marginTop: 8 }}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default Profile;
