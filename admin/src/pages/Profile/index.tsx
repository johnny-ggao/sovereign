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
      message.error('Please fill in all fields');
      return;
    }
    setLoading(true);
    try {
      const res = await changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });
      if (res.success) {
        message.success('Password changed successfully');
        setPasswordModalOpen(false);
        setOldPassword('');
        setNewPassword('');
      }
    } catch (error: any) {
      message.error(error?.message ?? 'Failed to change password');
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
            Change Password
          </Button>
        }
      >
        <Descriptions column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="Name">{admin.name}</Descriptions.Item>
          <Descriptions.Item label="Email">{admin.email}</Descriptions.Item>
          <Descriptions.Item label="Role">
            <Tag color={roleColorMap[admin.role] ?? 'default'}>{admin.role}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Last Login">
            {admin.last_login
              ? dayjs(admin.last_login).format('YYYY-MM-DD HH:mm')
              : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Modal
        title="Change Password"
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
          <label>Current Password:</label>
          <Input.Password
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            placeholder="Enter current password"
            style={{ marginTop: 8 }}
          />
        </div>
        <div>
          <label>New Password:</label>
          <Input.Password
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Enter new password"
            style={{ marginTop: 8 }}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default Profile;
