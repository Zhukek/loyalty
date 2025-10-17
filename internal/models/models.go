package models

import "time"

type User struct {
	UserPublic
	Pass string `json:"password"`
}

type UserPublic struct {
	Id  int    `json:"id"`
	Log string `json:"login"`
}

type OrderStatus string

const (
	OrderNew        OrderStatus = "NEW"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number   int         `json:"number"`
	Status   OrderStatus `json:"status"`
	Accrual  int         `json:"accrual,omitempty"`
	Uploaded time.Time   `json:"uploaded_at"`
	UserID   int         `json:"-"`
}
