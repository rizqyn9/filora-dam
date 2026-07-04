package adapters

import "testing"

func TestValidateCredentials(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		raw     string
		wantErr bool
	}{
		{"cloudinary ok", "cloudinary", `{"cloud_name":"c","api_key":"k","api_secret":"s"}`, false},
		{"cloudinary missing", "cloudinary", `{"cloud_name":"c"}`, true},
		{"imagekit ok", "imagekit", `{"public_key":"p","private_key":"pk","url_endpoint":"https://ik"}`, false},
		{"r2 ok", "r2", `{"access_key_id":"a","secret_access_key":"s","bucket_name":"b","endpoint":"https://e"}`, false},
		{"gcs ok", "gcs", `{"bucket_name":"b","service_account_key":"{}"}`, false},
		{"gcs missing", "gcs", `{"bucket_name":"b"}`, true},
		{"unknown type", "dropbox", `{}`, true},
		{"bad json", "cloudinary", `not-json`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCredentials(tc.typ, []byte(tc.raw))
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateCredentials(%s) err=%v, wantErr=%v", tc.typ, err, tc.wantErr)
			}
		})
	}
}

func TestNewAdapterValid(t *testing.T) {
	a, err := NewAdapter("cloudinary", []byte(`{"cloud_name":"c","api_key":"k","api_secret":"s"}`))
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}
	if a == nil {
		t.Fatal("expected adapter, got nil")
	}
}
