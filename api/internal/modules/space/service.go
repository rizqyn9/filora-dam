package space

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

// Membership returns a user's local role in a space (for other modules).
// Returns ErrNotMember if the user is not a member.
func (s *Service) Membership(ctx context.Context, spaceID, userID int64) (string, error) {
	role, err := s.repo.GetMemberRole(ctx, spaceID, userID)
	if err != nil {
		return "", err
	}
	return string(role), nil
}

// RoleOf returns a user's local role in a space and whether they are a member.
// Convenient for cross-module access checks (no sentinel error coupling).
func (s *Service) RoleOf(ctx context.Context, spaceID, userID int64) (string, bool, error) {
	role, err := s.repo.GetMemberRole(ctx, spaceID, userID)
	if errors.Is(err, ErrNotMember) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(role), true, nil
}

// QuotaInfo returns a space's used and quota bytes (for other modules).
func (s *Service) QuotaInfo(ctx context.Context, spaceID int64) (used, quota int64, err error) {
	sp, err := s.repo.GetByID(ctx, spaceID)
	if err != nil {
		return 0, 0, err
	}
	return sp.StorageUsed, sp.StorageQuota, nil
}

// AddUsed adjusts a space's used-bytes counter (delta may be negative).
func (s *Service) AddUsed(ctx context.Context, spaceID, delta int64) error {
	return s.repo.AddUsed(ctx, spaceID, delta)
}

// EnsureDefaultSpace creates the user's default space if missing (idempotent).
func (s *Service) EnsureDefaultSpace(ctx context.Context, userID int64) error {
	_, err := s.repo.GetDefault(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrSpaceNotFound) {
		return err
	}
	_, err = s.repo.CreateSpaceWithOwner(ctx, userID, "My Space", nil, true)
	return err
}

func (s *Service) Create(ctx context.Context, userID int64, in CreateSpaceInput) (Space, error) {
	if err := lib.Validate(in); err != nil {
		return Space{}, err
	}
	if err := s.authz.Require(ctx, userID, "space", "create"); err != nil {
		return Space{}, err
	}
	return s.repo.CreateSpaceWithOwner(ctx, userID, in.Name, in.Description, false)
}

func (s *Service) List(ctx context.Context, userID int64) ([]Space, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, spaceID int64) (Space, error) {
	if err := s.access(ctx, userID, spaceID, "read", db.MemberRoleViewer); err != nil {
		return Space{}, err
	}
	return s.getSpace(ctx, spaceID)
}

func (s *Service) Update(ctx context.Context, userID, spaceID int64, in UpdateSpaceInput) (Space, error) {
	if err := lib.Validate(in); err != nil {
		return Space{}, err
	}
	if err := s.access(ctx, userID, spaceID, "update", db.MemberRoleEditor); err != nil {
		return Space{}, err
	}
	return s.repo.Update(ctx, spaceID, in.Name, in.Description)
}

func (s *Service) Delete(ctx context.Context, userID, spaceID int64) error {
	if err := s.access(ctx, userID, spaceID, "delete", db.MemberRoleOwner); err != nil {
		return err
	}
	sp, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return err
	}
	if sp.IsDefault {
		return lib.ErrForbidden("the default space cannot be deleted")
	}
	return s.repo.Delete(ctx, spaceID)
}

// --- members ---

func (s *Service) ListMembers(ctx context.Context, userID, spaceID int64) ([]Member, error) {
	if err := s.access(ctx, userID, spaceID, "read", db.MemberRoleViewer); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, spaceID)
}

func (s *Service) UpdateMemberRole(ctx context.Context, actorID, spaceID, targetUserID int64, in UpdateMemberRoleInput) error {
	if err := lib.Validate(in); err != nil {
		return err
	}
	if err := s.access(ctx, actorID, spaceID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	ok, err := s.repo.UpdateMemberRole(ctx, spaceID, targetUserID, db.MemberRole(in.Role))
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("member not found")
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, spaceID, targetUserID int64) error {
	if err := s.access(ctx, actorID, spaceID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	sp, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return err
	}
	if sp.OwnerID == targetUserID {
		return lib.ErrForbidden("the space owner cannot be removed")
	}
	ok, err := s.repo.RemoveMember(ctx, spaceID, targetUserID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("member not found")
	}
	return nil
}

// --- invitations ---

func (s *Service) Invite(ctx context.Context, actorID, spaceID int64, in InviteInput) (Invitation, error) {
	if err := lib.Validate(in); err != nil {
		return Invitation{}, err
	}
	if err := s.access(ctx, actorID, spaceID, "invite", db.MemberRoleOwner); err != nil {
		return Invitation{}, err
	}
	token, err := generateToken()
	if err != nil {
		return Invitation{}, lib.ErrInternal("failed to generate invite token").Wrap(err)
	}
	expires := time.Now().Add(invitationTTL)
	return s.repo.CreateInvitation(ctx, spaceID, strings.ToLower(in.Email), db.MemberRole(in.Role), token, &actorID, &expires)
}

func (s *Service) ListInvitations(ctx context.Context, actorID, spaceID int64) ([]Invitation, error) {
	if err := s.access(ctx, actorID, spaceID, "invite", db.MemberRoleOwner); err != nil {
		return nil, err
	}
	return s.repo.ListInvitations(ctx, spaceID)
}

func (s *Service) RevokeInvitation(ctx context.Context, actorID, spaceID, invID int64) error {
	if err := s.access(ctx, actorID, spaceID, "invite", db.MemberRoleOwner); err != nil {
		return err
	}
	ok, err := s.repo.RevokeInvitation(ctx, invID, spaceID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("invitation not found")
	}
	return nil
}

// Accept consumes an invitation for the current user (email must match).
func (s *Service) Accept(ctx context.Context, userID int64, userEmail string, in AcceptInput) (Space, error) {
	if err := lib.Validate(in); err != nil {
		return Space{}, err
	}
	inv, err := s.repo.GetInvitationByToken(ctx, in.Token)
	if errors.Is(err, ErrInvitationNotFound) {
		return Space{}, lib.ErrNotFound("invitation not found")
	}
	if err != nil {
		return Space{}, err
	}
	if inv.Status != "pending" {
		return Space{}, lib.ErrConflict("invitation is no longer pending")
	}
	if inv.ExpiresAt != nil && inv.ExpiresAt.Before(time.Now()) {
		return Space{}, lib.ErrConflict("invitation has expired")
	}
	if !strings.EqualFold(inv.Email, userEmail) {
		return Space{}, lib.ErrForbidden("invitation was issued to a different email")
	}

	if err := s.repo.AcceptInvitationTx(ctx, inv.ID, inv.SpaceID, db.MemberRole(inv.Role), &userID, userID); err != nil {
		return Space{}, err
	}
	return s.getSpace(ctx, inv.SpaceID)
}

// --- helpers ---

func (s *Service) getSpace(ctx context.Context, id int64) (Space, error) {
	sp, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrSpaceNotFound) {
		return Space{}, lib.ErrNotFound("space not found")
	}
	return sp, err
}

// access enforces space:<action> globally, then (for own scope) a minimum
// local membership role.
func (s *Service) access(ctx context.Context, userID, spaceID int64, action string, minRole db.MemberRole) error {
	dec, err := s.authz.Authorize(ctx, userID, "space", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: space:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	role, err := s.repo.GetMemberRole(ctx, spaceID, userID)
	if errors.Is(err, ErrNotMember) {
		return lib.ErrForbidden("you are not a member of this space")
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
