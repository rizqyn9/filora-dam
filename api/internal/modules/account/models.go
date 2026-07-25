package account

import "time"

// User is the domain view of a user (mirror of a Clerk identity).
type User struct {
	ID          int64      `json:"id"`
	ClerkUserID string     `json:"clerk_user_id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UpdateProfileInput is the payload for updating the current user's profile.
type UpdateProfileInput struct {
	Name      string  `json:"name" validate:"required,min=1,max=255"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}
