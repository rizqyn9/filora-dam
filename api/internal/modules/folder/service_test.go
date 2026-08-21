package folder

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type fakeFolderRepo struct {
	moved     bool
	ancestors []db.GetFolderAncestorsRow
}

func (r *fakeFolderRepo) GetByID(_ context.Context, _ uuid.UUID) (*db.Folder, error) { return nil, nil }

func (r *fakeFolderRepo) ListByParent(_ context.Context, _, _ uuid.UUID) ([]db.Folder, error) {
	return nil, nil
}

func (r *fakeFolderRepo) ListRoot(_ context.Context, _ uuid.UUID) ([]db.Folder, error) {
	return nil, nil
}

func (r *fakeFolderRepo) Create(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string) (*db.Folder, error) {
	return nil, nil
}
func (r *fakeFolderRepo) Rename(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (r *fakeFolderRepo) Move(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	r.moved = true
	return nil
}
func (r *fakeFolderRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *fakeFolderRepo) Restore(_ context.Context, _ uuid.UUID) error    { return nil }
func (r *fakeFolderRepo) GetAncestors(_ context.Context, _ uuid.UUID) ([]db.GetFolderAncestorsRow, error) {
	return r.ancestors, nil
}

func TestMoveFolder_IntoSelf_Rejected(t *testing.T) {
	folderID := uuid.New()
	repo := &fakeFolderRepo{}
	svc := NewService(repo)

	err := svc.MoveFolder(context.Background(), folderID, &folderID)
	if err == nil {
		t.Fatal("expected error when moving folder into itself")
	}
	if !strings.Contains(err.Error(), "cannot move folder into itself") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.moved {
		t.Fatal("repo.Move should not have been called")
	}
}

func TestMoveFolder_IntoDescendant_Rejected(t *testing.T) {
	folderID := uuid.New()
	targetParent := uuid.New()

	// Simulate: targetParent's ancestors include folderID (making it a descendant)
	repo := &fakeFolderRepo{
		ancestors: []db.GetFolderAncestorsRow{
			{ID: targetParent, Name: "child"},
			{ID: folderID, Name: "self"}, // folderID is ancestor of targetParent → circular
		},
	}
	svc := NewService(repo)

	err := svc.MoveFolder(context.Background(), folderID, &targetParent)
	if err == nil {
		t.Fatal("expected error when moving into descendant")
	}
	if !strings.Contains(err.Error(), "cannot move folder into its own descendant") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.moved {
		t.Fatal("repo.Move should not have been called")
	}
}

func TestMoveFolder_ValidTarget_Succeeds(t *testing.T) {
	folderID := uuid.New()
	targetParent := uuid.New()

	// targetParent's ancestors don't include folderID
	repo := &fakeFolderRepo{
		ancestors: []db.GetFolderAncestorsRow{
			{ID: targetParent, Name: "target"},
			{ID: uuid.New(), Name: "root"},
		},
	}
	svc := NewService(repo)

	err := svc.MoveFolder(context.Background(), folderID, &targetParent)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !repo.moved {
		t.Fatal("expected repo.Move to be called")
	}
}

func TestMoveFolder_ToRoot_Succeeds(t *testing.T) {
	repo := &fakeFolderRepo{}
	svc := NewService(repo)

	err := svc.MoveFolder(context.Background(), uuid.New(), nil) // nil = move to root
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !repo.moved {
		t.Fatal("expected repo.Move to be called")
	}
}
