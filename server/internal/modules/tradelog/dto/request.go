package dto

type CreateTradeRequest struct {
	InvestmentID string `json:"investment_id" binding:"required,uuid"`
	Pair         string `json:"pair" binding:"required"`
	BuyExchange  string `json:"buy_exchange" binding:"required"`
	SellExchange string `json:"sell_exchange" binding:"required"`
	BuyPrice     string `json:"buy_price" binding:"required"`
	SellPrice    string `json:"sell_price" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	PremiumPct   string `json:"premium_pct" binding:"required"`
	PnL          string `json:"pnl" binding:"required"`
	Fee          string `json:"fee"`
	ExecutedAt   string `json:"executed_at" binding:"required"`
}

type BatchCreateTradeRequest struct {
	Trades []CreateTradeRequest `json:"trades" binding:"required,min=1,dive"`
}
