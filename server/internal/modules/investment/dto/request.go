package dto

type CreateInvestmentRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"omitempty,oneof=USDT"`
}

type RedeemRequest struct {
	InvestmentID string `json:"investment_id" binding:"required,uuid"`
}
