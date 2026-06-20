package repository

import (
	"context"

	"teamtask/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}
