package folder

import "time"

// Folder is the domain view of a folder.
type Folder struct {
	ID        int64     `json:"id"`
	SpaceID   int64     `json:"space_id"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BreadcrumbItem is a single entry in a folder breadcrumb trail.
type BreadcrumbItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// --- inputs ---

type CreateFolderInput struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	ParentID *int64 `json:"parent_id" validate:"omitempty"`
}

type UpdateFolderInput struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type MoveFolderInput struct {
	ParentID *int64 `json:"parent_id"` // NULL = move to root
}
