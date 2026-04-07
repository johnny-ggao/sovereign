export default [
  {
    path: '/user',
    layout: false,
    routes: [
      { name: 'login', path: '/user/login', component: './Login' },
    ],
  },
  { path: '/dashboard', name: '数据概览', icon: 'dashboard', component: './Dashboard' },
  { path: '/users', name: '用户管理', icon: 'team', component: './Users' },
  { path: '/investments', name: '投资管理', icon: 'fund', component: './Investments' },
  { path: '/users/:id', component: './UserDetail' },
  { path: '/admin-users', name: '管理员', icon: 'crown', component: './AdminUsers', access: 'isSuperAdmin' },
  { path: '/profile', name: '个人设置', icon: 'user', component: './Profile' },
  { path: '/', redirect: '/dashboard' },
  { path: '*', layout: false, component: './404' },
];
