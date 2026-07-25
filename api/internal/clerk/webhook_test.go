package clerk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func sign(secret, id, ts string, body []byte) string {
	raw := secret[len("whsec_"):]
	key, _ := base64.StdEncoding.DecodeString(raw)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + ts + "." + string(body)))
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifySvix(t *testing.T) {
	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("super-secret-key"))
	body := []byte(`{"type":"user.created","data":{"id":"user_1"}}`)
	h := SvixHeaders{ID: "msg_1", Timestamp: "1700000000"}
	h.Signature = sign(secret, h.ID, h.Timestamp, body)

	if err := VerifySvix(secret, h, body); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}

	// tampered body must fail
	if err := VerifySvix(secret, h, []byte(`{"type":"user.deleted"}`)); err == nil {
		t.Fatal("expected failure on tampered body")
	}

	// missing headers must fail
	if err := VerifySvix(secret, SvixHeaders{}, body); err == nil {
		t.Fatal("expected failure on missing headers")
	}
}

func TestParseWebhook(t *testing.T) {
	body := []byte(`{
		"type": "user.created",
		"data": {
			"id": "user_123",
			"primary_email_address_id": "idn_2",
			"email_addresses": [
				{"id": "idn_1", "email_address": "old@example.com"},
				{"id": "idn_2", "email_address": "primary@example.com"}
			],
			"first_name": "Ada",
			"last_name": "Lovelace",
			"image_url": "https://img/avatar.png"
		}
	}`)

	evtType, ident, err := ParseWebhook(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if evtType != "user.created" {
		t.Fatalf("type = %q", evtType)
	}
	if ident.ClerkUserID != "user_123" {
		t.Fatalf("clerk id = %q", ident.ClerkUserID)
	}
	if ident.Email != "primary@example.com" {
		t.Fatalf("email = %q, want primary", ident.Email)
	}
	if ident.Name != "Ada Lovelace" {
		t.Fatalf("name = %q", ident.Name)
	}
	if ident.AvatarURL == nil || *ident.AvatarURL != "https://img/avatar.png" {
		t.Fatalf("avatar = %v", ident.AvatarURL)
	}
}
