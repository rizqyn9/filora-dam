package gallery

import (
	"testing"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

func TestRoleRank(t *testing.T) {
	if !(roleRank(db.MemberRoleOwner) > roleRank(db.MemberRoleEditor)) {
		t.Fatal("owner should outrank editor")
	}
	if !(roleRank(db.MemberRoleEditor) > roleRank(db.MemberRoleViewer)) {
		t.Fatal("editor should outrank viewer")
	}
	if roleRank(db.MemberRoleViewer) <= 0 {
		t.Fatal("viewer should have positive rank")
	}
	if roleRank(db.MemberRole("nonsense")) != 0 {
		t.Fatal("unknown role should rank 0")
	}
}
