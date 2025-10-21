package repository

import (
	"context"

	"github.com/Zhukek/loyalty/internal/models"
)

type Repository interface {
	Close()
	Ping(ctx context.Context) error
	CreateUser(login string, hashed_pass string, ctx context.Context) (*models.UserPublic, error)
	GetUserByName(login string, ctx context.Context) (*models.User, error)
	CreateOrder(number string, userId int, status models.OrderStatus, ctx context.Context) error
	GetOrderByNum(number string, ctx context.Context) (*models.Order, error)
	GetUserOrders(userID int, ctx context.Context) ([]models.Order, error)
	GetProcessingOrders(ctx context.Context) ([]models.Order, error)
}
