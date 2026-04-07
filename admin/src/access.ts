export default function access(initialState: { currentAdmin?: API.AdminUser } | undefined) {
  const { currentAdmin } = initialState ?? {};
  return {
    isSuperAdmin: currentAdmin?.role === 'super_admin',
    isOperator: currentAdmin?.role === 'super_admin' || currentAdmin?.role === 'operator',
    isViewer: !!currentAdmin,
  };
}
