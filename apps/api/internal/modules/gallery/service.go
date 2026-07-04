package gallery

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

const invitationTTL = 7 * 24 * time.Hour

type Service struct {
	repo  *Repository
	authz *auth.Authorizer
}

func NewService(repo *Repository, authz *auth.Authorizer) *Service {
	return &Service{repo: repo, authz: authz}
}

// Membership returns a user's local role in a gallery (for other modules).
// Returns ErrNotMember if the user is not a member.
func (s *Service) Membership(ctx context.Context, galleryID, userID int64) (string, error) {
	role, err := s.repo.GetMemberRole(ctx, galleryID, userID)
	if err != nil {
		return "", err
	}
	return string(role), nil
}

// RoleOf returns a user's local role in a gallery and whether they are a member.
// Convenient for cross-module access checks (no sentinel error coupling).
func (s *Service) RoleOf(ctx context.Context, galleryID, userID int64) (string, bool, error) {
	role, err := s.repo.GetMemberRole(ctx, galleryID, userID)
	if errors.Is(err, ErrNotMember) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(role), true, nil
}

// QuotaInfo returns a gallery's used and quota bytes (for other modules).
func (s *Service) QuotaInfo(ctx context.Context, galleryID int64) (used, quota int64, err error) {
	g, err := s.repo.GetByID(ctx, galleryID)
	if err != nil {
		return 0, 0, err
	}
	return g.StorageUsed, g.StorageQuota, nil
}

// AddUsed adjusts a gallery's used-bytes counter (delta may be negative).
func (s *Service) AddUsed(ctx context.Context, galleryID, delta int64) error {
	return s.repo.AddUsed(ctx, galleryID, delta)
}

// EnsureDefaultGallery creates the user's default gallery if missing (idempotent).
func (s *Service) EnsureDefaultGallery(ctx context.Context, userID int64) error {
	_, err := s.repo.GetDefault(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrGalleryNotFound) {
		return err
	}
	_, err = s.repo.CreateGalleryWithOwner(ctx, userID, "My Gallery", nil, true)
	return err
}

func (s *Service) Create(ctx context.Context, userID int64, in CreateGalleryInput) (Gallery, error) {
	if err := lib.Validate(in); err != nil {
		return Gallery{}, err
	}
	if err := s.authz.Require(ctx, userID, "gallery", "create"); err != nil {
		return Gallery{}, err
	}
	return s.repo.CreateGalleryWithOwner(ctx, userID, in.Name, in.Description, false)
}

func (s *Service) List(ctx context.Context, userID int64) ([]Gallery, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, galleryID int64) (Gallery, error) {
	if err := s.access(ctx, userID, galleryID, "read", db.MemberRoleViewer); err != nil {
		return Gallery{}, err
	}
	return s.getGallery(ctx, galleryID)
}

func (s *Service) Update(ctx context.Context, userID, galleryID int64, in UpdateGalleryInput) (Gallery, error) {
	if err := lib.Validate(in); err != nil {
		return Gallery{}, err
	}
	if err := s.access(ctx, userID, galleryID, "update", db.MemberRoleEditor); err != nil {
		return Gallery{}, err
	}
	return s.repo.Update(ctx, galleryID, in.Name, in.Description)
}

func (s *Service) Delete(ctx context.Context, userID, galleryID int64) error {
	if err := s.access(ctx, userID, galleryID, "delete", db.MemberRoleOwner); err != nil {
		return err
	}
	g, err := s.getGallery(ctx, galleryID)
	if err != nil {
		return err
	}
	if g.IsDefault {
		return lib.ErrForbidden("the default gallery cannot be deleted")
	}
	return s.repo.Delete(ctx, galleryID)
}

// --- members ---

func (s *Service) ListMembers(ctx context.Context, userID, galleryID int64) ([]Member, error) {
	if err := s.access(ctx, userID, galleryID, "read", db.MemberRoleViewer); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, galleryID)
}

func (s *Service) UpdateMemberRole(ctx context.Context, actorID, galleryID, targetUserID int64, in UpdateMemberRoleInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	if err := s.access(ctx, actorID, galleryID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	ok, err := s.repo.UpdateMemberRole(ctx, galleryID, targetUserID, db.MemberRole(in.Role))
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("member not found")
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, galleryID, targetUserID int64) error {
	if err := s.access(ctx, actorID, galleryID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	g, err := s.getGallery(ctx, galleryID)
	if err != nil {
		return err
	}
	if g.OwnerID == targetUserID {
		return lib.ErrForbidden("the gallery owner cannot be removed")
	}
	ok, err := s.repo.RemoveMember(ctx, galleryID, targetUserID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("member not found")
	}
	return nil
}

// --- invitations ---

func (s *Service) Invite(ctx context.Context, actorID, galleryID int64, in InviteInput) (Invitation, error) {
	if err := lib.Validate(in); err != nil {
		return Invitation{}, err
	}
	if err := s.access(ctx, actorID, galleryID, "invite", db.MemberRoleOwner); err != nil {
		return Invitation{}, err
	}
	token, err := generateToken()
	if err != nil {
		return Invitation{}, lib.ErrInternal("failed to generate invite token").Wrap(err)
	}
	expires := time.Now().Add(invitationTTL)
	return s.repo.CreateInvitation(ctx, galleryID, strings.ToLower(in.Email), db.MemberRole(in.Role), token, &actorID, &expires)
}

func (s *Service) ListInvitations(ctx context.Context, actorID, galleryID int64) ([]Invitation, error) {
	if err := s.access(ctx, actorID, galleryID, "invite", db.MemberRoleOwner); err != nil {
		return nil, err
	}
	return s.repo.ListInvitations(ctx, galleryID)
}

func (s *Service) RevokeInvitation(ctx context.Context, actorID, galleryID, invID int64) error {
	if err := s.access(ctx, actorID, galleryID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	ok, err := s.repo.RevokeInvitation(ctx, invID, galleryID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("invitation not found")
	}
	return nil
}

// Accept consumes an invitation for the current user (email must match).
func (s *Service) Accept(ctx context.Context, userID int64, userEmail string, in AcceptInput) (Gallery, error) {
	if err := lib.Validate(in); err != nil {
		return Gallery{}, err
	}
	inv, err := s.repo.GetInvitationByToken(ctx, in.Token)
	if errors.Is(err, ErrInvitationNotFound) {
		return Gallery{}, lib.ErrNotFound("invitation not found")
	}
	if err != nil {
		return Gallery{}, err
	}
	if inv.Status != "pending" {
		return Gallery{}, lib.ErrConflict("invitation is no longer pending")
	}
	if inv.ExpiresAt != nil && inv.ExpiresAt.Before(time.Now()) {
		return Gallery{}, lib.ErrConflict("invitation has expired")
	}
	if inv.GalleryID == nil {
		return Gallery{}, lib.ErrBadRequest("not a gallery invitation")
	}
	if !strings.EqualFold(inv.Email, userEmail) {
		return Gallery{}, lib.ErrForbidden("invitation was issued to a different email")
	}

	if err := s.repo.AcceptInvitationTx(ctx, inv.ID, *inv.GalleryID, db.MemberRole(inv.Role), &userID, userID); err != nil {
		return Gallery{}, err
	}
	return s.getGallery(ctx, *inv.GalleryID)
}

// --- helpers ---

func (s *Service) getGallery(ctx context.Context, id int64) (Gallery, error) {
	g, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrGalleryNotFound) {
		return Gallery{}, lib.ErrNotFound("gallery not found")
	}
	return g, err
}

// access enforces gallery:<action> globally, then (for own scope) a minimum
// local membership role.
func (s *Service) access(ctx context.Context, userID, galleryID int64, action string, minRole db.MemberRole) error {
	dec, err := s.authz.Authorize(ctx, userID, "gallery", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: gallery:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	role, err := s.repo.GetMemberRole(ctx, galleryID, userID)
	if errors.Is(err, ErrNotMember) {
		return lib.ErrForbidden("you are not a member of this gallery")
	}
	if err != nil {
		return err
	}
	if roleRank(role) < roleRank(minRole) {
		return lib.ErrForbidden("requires " + string(minRole) + " role")
	}
	return nil
}

func roleRank(r db.MemberRole) int {
	switch r {
	case db.MemberRoleOwner:
		return 3
	case db.MemberRoleEditor:
		return 2
	case db.MemberRoleViewer:
		return 1
	}
	return 0
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
