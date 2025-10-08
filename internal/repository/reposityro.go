package repository

import "context"

type Repository interface {
	Close()
	Ping(ctx context.Context) error
}
