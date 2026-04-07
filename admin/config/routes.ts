export default [
  {
    path: '/user',
    layout: false,
    routes: [
      { name: 'login', path: '/user/login', component: './Login' },
    ],
  },
  { path: '/dashboard', name: 'Dashboard', icon: 'dashboard', component: './Dashboard' },
  { path: '/users', name: 'Users', icon: 'team', component: './Users' },
  { path: '/users/:id', component: './UserDetail' },
  { path: '/admin-users', name: 'Admins', icon: 'crown', component: './AdminUsers', access: 'isSuperAdmin' },
  { path: '/profile', name: 'Profile', icon: 'user', component: './Profile' },
  { path: '/', redirect: '/dashboard' },
  { path: '*', layout: false, component: './404' },
];
