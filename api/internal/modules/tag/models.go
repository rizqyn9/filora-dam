package tag

import "github.com/google/uuid"

type CreateTagRequest struct {
	SpaceID uuid.UUID `json:"space_id" validate:"required"`
	Name    string    `json:"name" validate:"required,min=1,max=100"`
}

type TagAssetRequest struct {
	SpaceID uuid.UUID `json:"space_id" validate:"required"`
	AssetID uuid.UUID `json:"asset_id" validate:"required"`
	TagID   int64     `json:"tag_id" validate:"required"`
}
