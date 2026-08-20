package storage

type CreateAccountRequest struct {
	Provider string `json:"provider" validate:"required,oneof=cloudinary imagekit r2 gcs"`
	Label    string `json:"label" validate:"required,min=1,max=255"`
	Layer    string `json:"layer" validate:"required,oneof=serving archive"`
	// Credentials is the raw JSON to be encrypted before storage.
	Credentials map[string]string `json:"credentials" validate:"required"`
	QuotaBytes  int64             `json:"quota_bytes"`
}

type UpdateAccountRequest struct {
	Label      string `json:"label" validate:"required,min=1,max=255"`
	IsActive   bool   `json:"is_active"`
	QuotaBytes int64  `json:"quota_bytes"`
}
