package repository

import (
	"context"

	"github.com/Zhukek/loyalty/internal/models"
)

type Repository interface {
	Close()
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, login string, hashedPass string) (*models.UserPublic, error)
	GetUserByName(ctx context.Context, login string) (*models.User, error)
	CreateOrder(ctx context.Context, number string, userID int, status models.OrderStatus) error
	GetOrderByNum(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
	GetProcessingOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrder(ctx context.Context, number string, status models.OrderStatus, accrual *float64) error
	UpdateOrderAndBalance(ctx context.Context, userID int, number string, status models.OrderStatus, accrual *float64) error
	GetWithdraws(ctx context.Context, userID int) ([]models.Withdraw, error)
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	MakeWithdraw(ctx context.Context, userID int, withdraw float64, orderNum string) error
}
