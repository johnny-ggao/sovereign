export default function access(initialState: { currentAdmin?: API.AdminUser } | undefined) {
  const { currentAdmin } = initialState ?? {};
  const role = currentAdmin?.role;
  return {
    isSuperAdmin: role === 'super_admin',
    isOperator: role === 'super_admin' || role === 'operator',
    isViewer: !!currentAdmin,
    canReviewWithdrawals: role === 'super_admin' || role === 'operator',
  };
}
