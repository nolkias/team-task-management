package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"teamtask/internal/domain"
	"teamtask/internal/jwtutil"
)

type stubUserRepo struct {
	byEmail   map[string]*domain.User
	byID      map[int64]*domain.User
	nextID    int64
	createErr error
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{byEmail: map[string]*domain.User{}, byID: map[int64]*domain.User{}}
}

func (s *stubUserRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	if s.createErr != nil {
		return 0, s.createErr
	}
	if _, exists := s.byEmail[u.Email]; exists {
		return 0, domain.ErrUserAlreadyExists
	}
	s.nextID++
	u.ID = s.nextID
	s.byEmail[u.Email] = u
	s.byID[u.ID] = u
	return u.ID, nil
}

func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := s.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (s *stubUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, ok := s.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		seedErr error
		wantErr error
	}{
		{name: "success", email: "new@example.com"},
		{name: "duplicate email", email: "existing@example.com", wantErr: domain.ErrUserAlreadyExists},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newStubUserRepo()
			if tt.wantErr == domain.ErrUserAlreadyExists {
				repo.byEmail[tt.email] = &domain.User{ID: 1, Email: tt.email}
			}
			issuer := jwtutil.NewIssuer("secret", time.Hour)
			svc := NewAuthService(repo, issuer)

			_, err := svc.Register(context.Background(), tt.email, "password123", "Name")

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{name: "success", email: "user@example.com", password: "correct-password"},
		{name: "wrong password", email: "user@example.com", password: "wrong-password", wantErr: domain.ErrInvalidCredentials},
		{name: "unknown user", email: "missing@example.com", password: "anything", wantErr: domain.ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newStubUserRepo()
			repo.byEmail["user@example.com"] = &domain.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash)}
			issuer := jwtutil.NewIssuer("secret", time.Hour)
			svc := NewAuthService(repo, issuer)

			token, err := svc.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token == "" {
				t.Fatal("expected non-empty token")
			}
		})
	}
}
