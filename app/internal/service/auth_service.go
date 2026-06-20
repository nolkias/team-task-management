package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"teamtask/internal/domain"
	"teamtask/internal/jwtutil"
	"teamtask/internal/repository"
)

type AuthService struct {
	users  repository.UserRepository
	issuer *jwtutil.Issuer
}

func NewAuthService(users repository.UserRepository, issuer *jwtutil.Issuer) *AuthService {
	return &AuthService{users: users, issuer: issuer}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	u := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	}
	return s.users.Create(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.issuer.Generate(u.ID)
}
