package space

import "time"

// Space is the domain view of a space (top-level container).
type Space struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	IsDefault    bool      `json:"is_default"`
	StorageQuota int64     `json:"storage_quota"`
	StorageUsed  int64     `json:"storage_used"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Member is a space member with basic user info.
type Member struct {
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Invitation is a pending space invitation (token included for the owner to share).
type Invitation struct {
	ID        int64      `json:"id"`
	SpaceID   int64      `json:"space_id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Token     string     `json:"token"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// --- inputs ---

type CreateSpaceInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateSpaceInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type InviteInput struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=editor viewer"`
}

type UpdateMemberRoleInput struct {
	Role string `json:"role" validate:"required,oneof=owner editor viewer"`
}

type AcceptInput struct {
	Token string `json:"token" validate:"required"`
}
