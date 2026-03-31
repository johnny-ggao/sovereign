package dto

type GetDepositAddressRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USDT BTC ETH"`
	Network  string `json:"network" binding:"required,oneof=ERC20 TRC20 BEP20"`
}

type WithdrawRequest struct {
	Currency  string `json:"currency" binding:"required,oneof=USDT BTC ETH"`
	Network   string `json:"network" binding:"required,oneof=ERC20 TRC20 BEP20"`
	Address   string `json:"address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	TwoFACode string `json:"two_fa_code" binding:"required,len=6"`
}

type AddWhitelistAddressRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USDT BTC ETH"`
	Network  string `json:"network" binding:"required,oneof=ERC20 TRC20 BEP20"`
	Address  string `json:"address" binding:"required"`
	Label    string `json:"label" binding:"omitempty,max=100"`
}
