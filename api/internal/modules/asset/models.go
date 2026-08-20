package asset

import "github.com/google/uuid"

type UploadRequest struct {
	SpaceID  uuid.UUID  `json:"space_id" validate:"required"`
	FolderID *uuid.UUID `json:"folder_id"` // nullable = space root
	Name     string     `json:"name" validate:"required,min=1,max=255"`
	// File content comes from multipart form, not JSON body.
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
