package folder

import "github.com/google/uuid"

type CreateFolderRequest struct {
	SpaceID  uuid.UUID  `json:"space_id" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"` // nullable = root level
	Name     string     `json:"name" validate:"required,min=1,max=255"`
}

type RenameFolderRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type MoveFolderRequest struct {
	ParentID *uuid.UUID `json:"parent_id"` // nullable = move to root
}

type BreadcrumbItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
