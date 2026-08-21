package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/crypto/bcrypt"

	iauth "github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
	"github.com/rizqyn9/filora-dam/api/internal/telemetry"
)

var tracer = otel.Tracer("filora-api")

const (
	WebIdleTTL = 7 * 24 * time.Hour
	CLIIdleTTL = 90 * 24 * time.Hour
)

func IdleTTL(client db.ClientType) time.Duration {
	if client == db.ClientTypeCli {
		return CLIIdleTTL
	}
	return WebIdleTTL
}

type Service struct {
	queries *db.Queries
	cache   *iauth.TokenCache
	metrics *telemetry.Metrics
}

func NewService(queries *db.Queries, cache *iauth.TokenCache, metrics *telemetry.Metrics) *Service {
	return &Service{queries: queries, cache: cache, metrics: metrics}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	ctx, span := tracer.Start(ctx, "auth.login")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", req.Email), attribute.String("auth.client", req.Client))

	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		s.metrics.AuthLogins.Add(ctx, 1, metric.WithAttributes(
			attribute.String("client", req.Client), attribute.Bool("success", false),
		))
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		s.metrics.AuthLogins.Add(ctx, 1, metric.WithAttributes(
			attribute.String("client", req.Client), attribute.Bool("success", false),
		))
		return nil, fmt.Errorf("invalid email or password")
	}

	token, sess, err := s.createSession(ctx, user.ID, db.ClientType(req.Client), req.Label)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "session creation failed")
		return nil, err
	}

	s.cache.Set(lib.HashToken(token), &iauth.CachedSession{
		SessionID: sess.ID,
		UserID:    user.ID,
		ExpiresAt: sess.ExpiresAt,
	})

	s.metrics.AuthLogins.Add(ctx, 1, metric.WithAttributes(
		attribute.String("client", req.Client), attribute.Bool("success", true),
	))
	slog.InfoContext(ctx, "user logged in", "user_id", user.ID, "client", req.Client)

	return &AuthResponse{
		Token: token,
		User:  UserData{ID: user.ID, Email: user.Email, Name: user.Name},
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	ctx, span := tracer.Start(ctx, "auth.register")
	defer span.End()

	tokenHash := lib.HashToken(req.InviteToken)
	invite, err := s.queries.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		span.SetStatus(codes.Error, "invalid invitation")
		return nil, fmt.Errorf("invalid or expired invitation")
	}

	span.SetAttributes(attribute.String("user.email", invite.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        invite.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create user failed")
		return nil, fmt.Errorf("create user: %w", err)
	}

	_ = s.queries.AssignRole(ctx, db.AssignRoleParams{UserID: user.ID, RoleName: "member"})
	_ = s.queries.AcceptInvitation(ctx, invite.ID)
	_, _ = s.queries.AddSpaceMember(ctx, db.AddSpaceMemberParams{
		SpaceID: invite.SpaceID,
		UserID:  user.ID,
		Role:    invite.Role,
	})

	token, sess, err := s.createSession(ctx, user.ID, db.ClientTypeWeb, "")
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.cache.Set(lib.HashToken(token), &iauth.CachedSession{
		SessionID: sess.ID,
		UserID:    user.ID,
		ExpiresAt: sess.ExpiresAt,
	})

	slog.InfoContext(ctx, "user registered", "user_id", user.ID, "email", invite.Email)

	return &AuthResponse{
		Token: token,
		User:  UserData{ID: user.ID, Email: user.Email, Name: user.Name},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID int64) error {
	return s.queries.RevokeSession(ctx, sessionID)
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, req ChangePasswordRequest) error {
	ctx, span := tracer.Start(ctx, "auth.change_password")
	defer span.End()

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		span.SetStatus(codes.Error, "incorrect password")
		return fmt.Errorf("incorrect current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.queries.UpdatePassword(ctx, db.UpdatePasswordParams{ID: userID, PasswordHash: string(hash)}); err != nil {
		span.RecordError(err)
		return fmt.Errorf("update password: %w", err)
	}

	_ = s.queries.RevokeAllUserSessions(ctx, userID)
	s.cache.DeleteByUserID(userID)

	slog.InfoContext(ctx, "password changed", "user_id", userID)
	return nil
}

func (s *Service) GetMe(ctx context.Context, userID int64) (*UserData, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &UserData{ID: user.ID, Email: user.Email, Name: user.Name}, nil
}

func (s *Service) GetByTokenHash(ctx context.Context, hash string) (*db.Session, error) {
	sess, err := s.queries.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	return &sess, nil
}

func (s *Service) Touch(ctx context.Context, id int64, newExpiry time.Time) error {
	return s.queries.TouchSession(ctx, db.TouchSessionParams{ID: id, ExpiresAt: newExpiry})
}

func (s *Service) createSession(ctx context.Context, userID int64, client db.ClientType, label string) (string, *db.Session, error) {
	raw, err := lib.GenerateToken()
	if err != nil {
		return "", nil, err
	}

	expiresAt := time.Now().Add(IdleTTL(client))
	sess, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: lib.HashToken(raw),
		Client:    client,
		Label:     label,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", nil, fmt.Errorf("create session: %w", err)
	}

	return raw, &sess, nil
}
