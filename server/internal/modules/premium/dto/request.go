package dto

type PremiumHistoryRequest struct {
	Pair     string `form:"pair" binding:"omitempty"`
	Interval string `form:"interval" binding:"omitempty,oneof=1m 5m 15m 1h 4h 1d"`
	From     string `form:"from" binding:"omitempty"`
	To       string `form:"to" binding:"omitempty"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=1000"`
}

type WSSubscribeMessage struct {
	Action string   `json:"action"` // subscribe, unsubscribe
	Pairs  []string `json:"pairs"`
}
