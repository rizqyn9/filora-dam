package asset

import (
	"io"

	"github.com/google/uuid"
)

// UploadInput groups all parameters for the upload flow.
type UploadInput struct {
	SpaceID     uuid.UUID
	FolderID    *uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	Body        io.Reader
}

type CreateReferenceRequest struct {
	AssetID  uuid.UUID  `json:"asset_id" validate:"required"`
	SpaceID  uuid.UUID  `json:"space_id" validate:"required"`
	FolderID *uuid.UUID `json:"folder_id"`
}

type ListAssetsParams struct {
	SpaceID  uuid.UUID
	FolderID *uuid.UUID
	Limit    int32
	Offset   int32
}
