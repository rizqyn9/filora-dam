package storage

import "github.com/rizqyn9/filora-dam/api/internal/modules/storage/adapters"

// Re-export adapter types so the rest of the storage package uses them without direct adapters import everywhere.
type (
	Adapter      = adapters.Adapter
	UploadInput  = adapters.UploadInput
	UploadResult = adapters.UploadResult
)
