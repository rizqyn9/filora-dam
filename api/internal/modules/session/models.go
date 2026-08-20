package session

type CreateSessionRequest struct {
	Label string `json:"label" validate:"max=255"`
}

type SessionResponse struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Token string `json:"token,omitempty"` // only returned on create
}
