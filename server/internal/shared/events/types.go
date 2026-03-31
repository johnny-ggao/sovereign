package events

const (
	UserRegistered     = "user.registered"
	UserLoggedIn       = "user.logged_in"
	UserPasswordReset  = "user.password_reset"

	DepositDetected    = "wallet.deposit.detected"
	DepositConfirmed   = "wallet.deposit.confirmed"
	WithdrawRequested  = "wallet.withdraw.requested"
	WithdrawCompleted  = "wallet.withdraw.completed"
	WithdrawFailed     = "wallet.withdraw.failed"

	InvestmentCreated  = "investment.created"
	InvestmentRedeemed = "investment.redeemed"

	SettlementCreated  = "settlement.created"

	PremiumUpdated     = "premium.updated"
)
