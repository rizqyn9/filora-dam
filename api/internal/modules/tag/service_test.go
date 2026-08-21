package tag

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type fakeTagRepo struct {
	tags   map[int64]*db.Tag
	tagged []struct {
		assetID uuid.UUID
		tagID   int64
	}
}

func newFakeTagRepo() *fakeTagRepo {
	return &fakeTagRepo{tags: make(map[int64]*db.Tag)}
}

func (r *fakeTagRepo) GetByID(_ context.Context, id int64) (*db.Tag, error) {
	if t, ok := r.tags[id]; ok {
		return t, nil
	}
	return nil, errNotFound
}

func (r *fakeTagRepo) ListBySpace(_ context.Context, _ uuid.UUID) ([]db.Tag, error) { return nil, nil }

func (r *fakeTagRepo) Create(_ context.Context, _ uuid.UUID, _ string) (*db.Tag, error) {
	return nil, nil
}

func (r *fakeTagRepo) Delete(_ context.Context, _ int64) error { return nil }

func (r *fakeTagRepo) AddAssetTag(_ context.Context, assetID uuid.UUID, tagID int64) error {
	r.tagged = append(r.tagged, struct {
		assetID uuid.UUID
		tagID   int64
	}{assetID, tagID})
	return nil
}
func (r *fakeTagRepo) RemoveAssetTag(_ context.Context, _ uuid.UUID, _ int64) error { return nil }
func (r *fakeTagRepo) ListByAsset(_ context.Context, _ uuid.UUID) ([]db.Tag, error) { return nil, nil }

var errNotFound = fmt.Errorf("not found")

func TestTagAsset_SameSpace_Succeeds(t *testing.T) {
	spaceID := uuid.New()
	repo := newFakeTagRepo()
	repo.tags[1] = &db.Tag{ID: 1, SpaceID: spaceID, Name: "vacation"}

	svc := NewService(repo)

	err := svc.TagAsset(context.Background(), spaceID, uuid.New(), 1)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(repo.tagged) != 1 {
		t.Fatal("expected tag to be applied")
	}
}

func TestTagAsset_CrossSpace_Rejected(t *testing.T) {
	tagSpace := uuid.New()
	otherSpace := uuid.New()

	repo := newFakeTagRepo()
	repo.tags[1] = &db.Tag{ID: 1, SpaceID: tagSpace, Name: "vacation"}

	svc := NewService(repo)

	err := svc.TagAsset(context.Background(), otherSpace, uuid.New(), 1)
	if err == nil {
		t.Fatal("expected error for cross-space tag")
	}
	if !strings.Contains(err.Error(), "does not belong to space") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.tagged) != 0 {
		t.Fatal("tag should not have been applied")
	}
}

func TestTagAsset_NonexistentTag_Rejected(t *testing.T) {
	repo := newFakeTagRepo()
	svc := NewService(repo)

	err := svc.TagAsset(context.Background(), uuid.New(), uuid.New(), 999)
	if err == nil {
		t.Fatal("expected error for nonexistent tag")
	}
	if !strings.Contains(err.Error(), "tag not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
