package adapters

import "testing"

func TestR2PublicURL(t *testing.T) {
	withURL := newR2Adapter(R2Credentials{
		AccessKeyID: "a", SecretAccessKey: "s", BucketName: "b",
		Endpoint: "https://acc.r2.cloudflarestorage.com", PublicBaseURL: "https://cdn.example.com/",
	})
	if got := withURL.publicURL("galleries/1/x.jpg"); got != "https://cdn.example.com/galleries/1/x.jpg" {
		t.Fatalf("publicURL = %q", got)
	}

	noURL := newR2Adapter(R2Credentials{
		AccessKeyID: "a", SecretAccessKey: "s", BucketName: "b",
		Endpoint: "https://acc.r2.cloudflarestorage.com",
	})
	if got := noURL.publicURL("x"); got != "" {
		t.Fatalf("expected empty public URL, got %q", got)
	}
}

func TestNewAdapterR2Concrete(t *testing.T) {
	a, err := NewAdapter("r2", []byte(`{"access_key_id":"a","secret_access_key":"s","bucket_name":"b","endpoint":"https://e"}`))
	if err != nil {
		t.Fatalf("NewAdapter r2: %v", err)
	}
	if _, ok := a.(*r2Adapter); !ok {
		t.Fatalf("expected *r2Adapter, got %T", a)
	}
}
