package session

import (
	"time"

	"github.com/google/uuid"
)

// Session is the domain view of a CLI session (raw token never stored/returned
// except once at issue time).
type Session struct {
	ID         uuid.UUID  `json:"id"`
	UserID     int64      `json:"user_id"`
	Label      *string    `json:"label,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IssueInput is the request to create a new CLI session.
type IssueInput struct {
	Label *string `json:"label" validate:"omitempty,max=255"`
}

// IssueResult returns the raw token exactly once, plus session metadata.
type IssueResult struct {
	Token   string  `json:"token"`
	Session Session `json:"session"`
}
