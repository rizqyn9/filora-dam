package tag

import "time"

type Tag struct {
	ID        int64     `json:"id"`
	GalleryID int64     `json:"gallery_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTagInput struct {
	Name string `json:"name" validate:"required,min=1,max=64"`
}

type UpdateTagInput struct {
	Name string `json:"name" validate:"required,min=1,max=64"`
}

type AttachInput struct {
	AssetID string `json:"asset_id" validate:"required,uuid"`
}
