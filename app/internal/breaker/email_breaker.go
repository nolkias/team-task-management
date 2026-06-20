package breaker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"

	"teamtask/internal/service"
)

type EmailServiceBreaker struct {
	inner   service.EmailService
	breaker *gobreaker.CircuitBreaker
}

func NewEmailServiceBreaker(inner service.EmailService) *EmailServiceBreaker {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "email_service",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})
	return &EmailServiceBreaker{inner: inner, breaker: cb}
}

func (b *EmailServiceBreaker) SendInvite(ctx context.Context, toEmail, teamName string) error {
	_, err := b.breaker.Execute(func() (interface{}, error) {
		return nil, b.inner.SendInvite(ctx, toEmail, teamName)
	})
	return err
}
