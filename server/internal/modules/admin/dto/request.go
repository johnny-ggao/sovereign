package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=super_admin operator viewer"`
}

type UpdateAdminRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role" binding:"omitempty,oneof=super_admin operator viewer"`
	IsActive *bool  `json:"is_active"`
}

type UserListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
	Search string `form:"search"`
	Status string `form:"status"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Language string `json:"language"`
}

type AdjustBalanceRequest struct {
	Currency string `json:"currency" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

type InvestmentListQuery struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	Search    string `form:"search"`
	Status    string `form:"status"`
	SortBy    string `form:"sort_by,default=created_at"`
	SortOrder string `form:"sort_order,default=desc"`
}

type TradeListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Pair     string `form:"pair"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

type TransactionListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Type     string `form:"type"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}
