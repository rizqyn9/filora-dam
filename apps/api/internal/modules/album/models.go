package album

import "time"

type Album struct {
	ID           int64     `json:"id"`
	GalleryID    int64     `json:"gallery_id"`
	OwnerID      int64     `json:"owner_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	CoverAssetID *string   `json:"cover_asset_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Member struct {
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAlbumInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateAlbumInput struct {
	Name         string  `json:"name" validate:"required,min=1,max=255"`
	Description  *string `json:"description" validate:"omitempty,max=1000"`
	CoverAssetID *string `json:"cover_asset_id" validate:"omitempty,uuid"`
}

type AddMemberInput struct {
	UserID int64  `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=editor viewer"`
}

type AddAssetInput struct {
	AssetID   string `json:"asset_id" validate:"required,uuid"`
	SortOrder int32  `json:"sort_order"`
}
