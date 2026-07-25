package account

import (
	"context"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
)

// ProcessWebhook records a Clerk webhook delivery idempotently and applies it.
// Returns handled=false when the event was already processed (duplicate).
func (s *Service) ProcessWebhook(ctx context.Context, eventID, eventType string, payload []byte, ident auth.ClerkIdentity) (bool, error) {
	id, inserted, err := s.repo.InsertWebhookEvent(ctx, eventID, eventType, payload)
	if err != nil {
		return false, err
	}
	if !inserted {
		return false, nil // duplicate delivery, already handled
	}

	var derr error
	switch eventType {
	case "user.created", "user.updated":
		if ident.ClerkUserID != "" && ident.Email != "" {
			_, derr = s.SyncFromClerk(ctx, ident)
		}
	case "user.deleted":
		if ident.ClerkUserID != "" {
			derr = s.repo.DeactivateByClerkID(ctx, ident.ClerkUserID)
		}
	default:
		// Unhandled event types are recorded and acknowledged.
	}

	if derr != nil {
		_ = s.repo.MarkWebhookFailed(ctx, id, derr.Error())
		return true, derr
	}
	_ = s.repo.MarkWebhookProcessed(ctx, id)
	return true, nil
}
