package models

import "time"

type User struct {
	UserPublic
	Pass    string  `json:"password"`
	Balance float64 `json:"balance"`
}

type UserPublic struct {
	Id  int    `json:"id"`
	Log string `json:"login"`
}

type OrderStatus string

const (
	OrderNew        OrderStatus = "NEW"
	OrderRegistered OrderStatus = "REGISTERED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number   string      `json:"number"`
	Status   OrderStatus `json:"status"`
	Accrual  float64     `json:"accrual,omitempty"`
	Uploaded time.Time   `json:"uploaded_at"`
	UserID   int         `json:"-"`
}

type AccrualOrder struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float64     `json:"accrual"`
}

type Withdraw struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
