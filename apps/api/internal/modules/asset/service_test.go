package asset

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildKey(t *testing.T) {
	id := uuid.MustParse("018f0000-0000-7000-8000-000000000abc")
	key := buildKey(42, id, "Vacation Photo.JPG")
	if !strings.HasPrefix(key, "galleries/42/") {
		t.Fatalf("key %q missing gallery prefix", key)
	}
	if !strings.HasSuffix(key, ".JPG") {
		t.Fatalf("key %q should keep extension", key)
	}
	if !strings.Contains(key, id.String()) {
		t.Fatalf("key %q should contain asset id", key)
	}
}

func TestRank(t *testing.T) {
	if !(rank("owner") > rank("editor") && rank("editor") > rank("viewer") && rank("viewer") > 0) {
		t.Fatal("role rank ordering is wrong")
	}
	if rank("nope") != 0 {
		t.Fatal("unknown role should be 0")
	}
}
