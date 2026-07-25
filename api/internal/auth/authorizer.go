package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Scope is how far a permission grant reaches.
type Scope string

const (
	ScopeOwn Scope = "own"
	ScopeAll Scope = "all"
)

// Permission is a single effective grant for a user.
type Permission struct {
	Resource string
	Action   string
	Scope    Scope
}

// Decision is the result of an authorization check.
type Decision struct {
	Allowed bool
	Scope   Scope
}

// Decide evaluates a permission set against a requested (resource, action),
// honoring '*' wildcards and choosing the widest matching scope (all > own).
// Pure function — unit-testable without a database.
func Decide(perms []Permission, resource, action string) Decision {
	d := Decision{}
	for _, p := range perms {
		resMatch := p.Resource == resource || p.Resource == "*"
		actMatch := p.Action == action || p.Action == "*"
		if !resMatch || !actMatch {
			continue
		}
		d.Allowed = true
		if p.Scope == ScopeAll {
			d.Scope = ScopeAll
		} else if d.Scope != ScopeAll {
			d.Scope = ScopeOwn
		}
	}
	return d
}

// Authorizer resolves a user's global RBAC permissions.
type Authorizer struct {
	q *db.Queries
}

func NewAuthorizer(pool *pgxpool.Pool) *Authorizer {
	return &Authorizer{q: db.New(pool)}
}

// Authorize returns the decision for (resource, action) for the given user.
func (a *Authorizer) Authorize(ctx context.Context, userID int64, resource, action string) (Decision, error) {
	rows, err := a.q.GetUserPermissions(ctx, userID)
	if err != nil {
		return Decision{}, err
	}
	perms := make([]Permission, 0, len(rows))
	for _, r := range rows {
		perms = append(perms, Permission{
			Resource: r.Resource,
			Action:   r.Action,
			Scope:    Scope(r.Scope),
		})
	}
	return Decide(perms, resource, action), nil
}

// Require returns a Forbidden AppError unless the user has the permission.
func (a *Authorizer) Require(ctx context.Context, userID int64, resource, action string) error {
	d, err := a.Authorize(ctx, userID, resource, action)
	if err != nil {
		return err
	}
	if !d.Allowed {
		return lib.ErrForbidden("insufficient permission: " + resource + ":" + action)
	}
	return nil
}
