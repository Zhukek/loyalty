package repository

import (
	"context"

	model "github.com/Zhukek/loyalty/internal/models"
)

type Repository interface {
	Close()
	Ping(ctx context.Context) error
	CreateUser(login string, hashed_pass string, ctx context.Context) (*model.User, error)
}
