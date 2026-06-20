package service

import (
	"context"
	"errors"
)

var ErrEmailDeliveryFailed = errors.New("email delivery failed")

type EmailService interface {
	SendInvite(ctx context.Context, toEmail, teamName string) error
}

type MockEmailService struct {
	ShouldFail bool
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendInvite(ctx context.Context, toEmail, teamName string) error {
	if m.ShouldFail {
		return ErrEmailDeliveryFailed
	}
	return nil
}
