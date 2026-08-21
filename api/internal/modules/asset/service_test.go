package asset

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/telemetry"
)

// --- Fakes ---

type fakeRepo struct {
	assets     map[string]*db.Asset // key = checksum
	references []db.AssetReference
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{assets: make(map[string]*db.Asset)}
}

func (r *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*db.Asset, error) {
	for _, a := range r.assets {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errNotFound
}

func (r *fakeRepo) GetByChecksum(_ context.Context, hash string) (*db.Asset, error) {
	if a, ok := r.assets[hash]; ok {
		return a, nil
	}
	return nil, errNotFound
}

func (r *fakeRepo) Create(_ context.Context, params db.CreateAssetParams) (*db.Asset, error) {
	a := &db.Asset{
		ID:               uuid.New(),
		OriginalFilename: params.OriginalFilename,
		Name:             params.Name,
		MimeType:         params.MimeType,
		SizeBytes:        params.SizeBytes,
		ChecksumSha256:   params.ChecksumSha256,
		UploadedBy:       params.UploadedBy,
	}
	r.assets[params.ChecksumSha256] = a
	return a, nil
}

func (r *fakeRepo) CreateReference(_ context.Context, assetID, spaceID uuid.UUID, _ *uuid.UUID) (*db.AssetReference, error) {
	ref := db.AssetReference{ID: int64(len(r.references) + 1), AssetID: assetID, SpaceID: spaceID}
	r.references = append(r.references, ref)
	return &ref, nil
}

func (r *fakeRepo) ListByFolder(_ context.Context, _, _ uuid.UUID, _, _ int32) ([]db.Asset, error) {
	return nil, nil
}

func (r *fakeRepo) ListBySpaceRoot(_ context.Context, _ uuid.UUID, _, _ int32) ([]db.Asset, error) {
	return nil, nil
}
func (r *fakeRepo) SoftDeleteReference(_ context.Context, _ int64) error      { return nil }
func (r *fakeRepo) RestoreReference(_ context.Context, _ int64) error         { return nil }
func (r *fakeRepo) UpdateName(_ context.Context, _ uuid.UUID, _ string) error { return nil }

type fakeStorage struct {
	accounts []db.StorageAccount
}

func (s *fakeStorage) ElectAccount(_ context.Context, _ db.StorageLayer, _ int64) (*db.StorageAccount, error) {
	if len(s.accounts) == 0 {
		return nil, errNotFound
	}
	return &s.accounts[0], nil
}

type fakeUploader struct {
	uploaded []string // keys
}

func (u *fakeUploader) UploadAndRecord(_ context.Context, _ uuid.UUID, _ int64, _ db.StorageLayer, key string, _ io.Reader, _ int64, _ string) (*db.StorageLocation, error) {
	u.uploaded = append(u.uploaded, key)
	return &db.StorageLocation{}, nil
}

type fakeJobCreator struct {
	jobs []uuid.UUID
}

func (j *fakeJobCreator) CreateArchiveJob(_ context.Context, assetID uuid.UUID) error {
	j.jobs = append(j.jobs, assetID)
	return nil
}

type fakeQuota struct {
	used  int64
	quota int64
}

func (q *fakeQuota) CheckQuota(_ context.Context, _ uuid.UUID, additional int64) error {
	if q.quota > 0 && q.used+additional > q.quota {
		return errQuotaExceeded
	}
	return nil
}

func (q *fakeQuota) IncrementUsage(_ context.Context, _ uuid.UUID, bytes int64) error {
	q.used += bytes
	return nil
}

func (q *fakeQuota) DecrementUsage(_ context.Context, _ uuid.UUID, bytes int64) error {
	q.used -= bytes
	return nil
}

var (
	errNotFound      = fmt.Errorf("not found")
	errQuotaExceeded = fmt.Errorf("quota exceeded")
)

// --- Tests ---

func TestUpload_DedupHit_NoPhysicalUpload(t *testing.T) {
	repo := newFakeRepo()
	// Pre-seed an existing asset with known checksum
	existingAsset := &db.Asset{
		ID:             uuid.New(),
		Name:           "existing.jpg",
		SizeBytes:      100,
		ChecksumSha256: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", // sha256("hello")
		MimeType:       "image/jpeg",
		UploadedBy:     1,
	}
	repo.assets[existingAsset.ChecksumSha256] = existingAsset

	uploader := &fakeUploader{}
	jobs := &fakeJobCreator{}
	quota := &fakeQuota{quota: 10000}

	svc := NewService(repo, &fakeStorage{}, uploader, jobs, quota, telemetry.NewMetrics())

	result, err := svc.Upload(context.Background(), 1, UploadInput{
		SpaceID:     uuid.New(),
		Filename:    "new.jpg",
		ContentType: "image/jpeg",
		Size:        5,
		Body:        strings.NewReader("hello"), // sha256 = same as existing
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.ID != existingAsset.ID {
		t.Fatalf("expected dedup to return existing asset %s, got %s", existingAsset.ID, result.ID)
	}
	if len(uploader.uploaded) != 0 {
		t.Fatalf("expected no physical upload on dedup hit, got %d uploads", len(uploader.uploaded))
	}
	if len(jobs.jobs) != 0 {
		t.Fatalf("expected no archive job on dedup hit, got %d", len(jobs.jobs))
	}
	if len(repo.references) != 1 {
		t.Fatalf("expected 1 reference created, got %d", len(repo.references))
	}
}

func TestUpload_NewFile_UploadsAndCreatesArchiveJob(t *testing.T) {
	repo := newFakeRepo()
	account := db.StorageAccount{ID: 1, Provider: db.StorageProviderR2}
	uploader := &fakeUploader{}
	jobs := &fakeJobCreator{}
	quota := &fakeQuota{quota: 10000}

	svc := NewService(repo, &fakeStorage{accounts: []db.StorageAccount{account}}, uploader, jobs, quota, telemetry.NewMetrics())

	result, err := svc.Upload(context.Background(), 1, UploadInput{
		SpaceID:     uuid.New(),
		Filename:    "photo.png",
		ContentType: "image/png",
		Size:        50,
		Body:        strings.NewReader("unique content here"),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected asset, got nil")
	}
	if len(uploader.uploaded) != 1 {
		t.Fatalf("expected 1 physical upload, got %d", len(uploader.uploaded))
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected 1 archive job, got %d", len(jobs.jobs))
	}
	if len(repo.references) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(repo.references))
	}
}

func TestUpload_QuotaExceeded_Rejects(t *testing.T) {
	repo := newFakeRepo()
	quota := &fakeQuota{used: 9990, quota: 10000} // only 10 bytes left

	svc := NewService(repo, &fakeStorage{}, &fakeUploader{}, &fakeJobCreator{}, quota, telemetry.NewMetrics())

	_, err := svc.Upload(context.Background(), 1, UploadInput{
		SpaceID:     uuid.New(),
		Filename:    "big.zip",
		ContentType: "application/zip",
		Size:        100,
		Body:        strings.NewReader(strings.Repeat("x", 100)), // 100 bytes > 10 remaining
	})

	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("expected quota exceeded error, got: %v", err)
	}
}
