export default [
  {
    path: '/user',
    layout: false,
    routes: [
      { name: 'login', path: '/user/login', component: './Login' },
    ],
  },
  {
    path: '/change-password',
    component: './ChangePassword',
    layout: false,
  },
  { path: '/dashboard', name: '数据概览', icon: 'dashboard', component: './Dashboard' },
  { path: '/users', name: '用户管理', icon: 'team', component: './Users' },
  { path: '/investments', name: '投资管理', icon: 'fund', component: './Investments' },
  { path: '/trades', name: '套利交易', icon: 'swap', component: './Trades' },
  { path: '/transactions', name: '充提管理', icon: 'transaction', component: './Transactions' },
  { path: '/users/:id', component: './UserDetail' },
  { path: '/admin-users', name: '管理员', icon: 'crown', component: './AdminUsers', access: 'isSuperAdmin' },
  { path: '/audit-logs', name: '审计日志', icon: 'fileSearch', component: './AuditLogs', access: 'isSuperAdmin' },
  { path: '/profile', name: '个人设置', icon: 'user', component: './Profile' },
  { path: '/', redirect: '/dashboard' },
  { path: '*', layout: false, component: './404' },
];
