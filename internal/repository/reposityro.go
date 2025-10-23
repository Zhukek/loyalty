package repository

import (
	"context"

	"github.com/Zhukek/loyalty/internal/models"
)

type Repository interface {
	Close()
	Ping(ctx context.Context) error
	CreateUser(login string, hashedPass string, ctx context.Context) (*models.UserPublic, error)
	GetUserByName(login string, ctx context.Context) (*models.User, error)
	CreateOrder(number string, userID int, status models.OrderStatus, ctx context.Context) error
	GetOrderByNum(number string, ctx context.Context) (*models.Order, error)
	GetUserOrders(userID int, ctx context.Context) ([]models.Order, error)
	GetProcessingOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrder(number string, status models.OrderStatus, accrual *float64, ctx context.Context) error
	UpdateOrderAndBalance(userID int, number string, status models.OrderStatus, accrual *float64, ctx context.Context) error
	GetWithdraws(userID int, ctx context.Context) ([]models.Withdraw, error)
	GetUserBalance(userID int, ctx context.Context) (*models.Balance, error)
	MakeWithdraw(userID int, withdraw float64, orderNum string, ctx context.Context) error
}
