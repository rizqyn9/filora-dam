package auth

import "testing"

func TestDecide(t *testing.T) {
	tests := []struct {
		name        string
		perms       []Permission
		resource    string
		action      string
		wantAllowed bool
		wantScope   Scope
	}{
		{
			name:        "no permissions denies",
			perms:       nil,
			resource:    "asset",
			action:      "read",
			wantAllowed: false,
		},
		{
			name:        "exact match own",
			perms:       []Permission{{"asset", "read", ScopeOwn}},
			resource:    "asset",
			action:      "read",
			wantAllowed: true,
			wantScope:   ScopeOwn,
		},
		{
			name:        "superuser wildcard all",
			perms:       []Permission{{"*", "*", ScopeAll}},
			resource:    "gallery",
			action:      "delete",
			wantAllowed: true,
			wantScope:   ScopeAll,
		},
		{
			name:        "resource wildcard",
			perms:       []Permission{{"asset", "*", ScopeAll}},
			resource:    "asset",
			action:      "download",
			wantAllowed: true,
			wantScope:   ScopeAll,
		},
		{
			name:        "widest scope wins (all beats own)",
			perms:       []Permission{{"asset", "read", ScopeOwn}, {"*", "*", ScopeAll}},
			resource:    "asset",
			action:      "read",
			wantAllowed: true,
			wantScope:   ScopeAll,
		},
		{
			name:        "action mismatch denies",
			perms:       []Permission{{"asset", "read", ScopeAll}},
			resource:    "asset",
			action:      "delete",
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.perms, tt.resource, tt.action)
			if got.Allowed != tt.wantAllowed {
				t.Fatalf("Allowed = %v, want %v", got.Allowed, tt.wantAllowed)
			}
			if got.Allowed && got.Scope != tt.wantScope {
				t.Fatalf("Scope = %v, want %v", got.Scope, tt.wantScope)
			}
		})
	}
}
