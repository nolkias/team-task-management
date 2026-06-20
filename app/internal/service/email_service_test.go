package service

import (
	"context"
	"errors"
	"testing"
)

func TestMockEmailService_SendInvite(t *testing.T) {
	svc := NewMockEmailService()
	if err := svc.SendInvite(context.Background(), "user@example.com", "Engineering"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc.ShouldFail = true
	if err := svc.SendInvite(context.Background(), "user@example.com", "Engineering"); !errors.Is(err, ErrEmailDeliveryFailed) {
		t.Fatalf("expected ErrEmailDeliveryFailed, got %v", err)
	}
}
