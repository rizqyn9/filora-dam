// Package clerk adapts Clerk (https://clerk.com) to Filora: verifying session
// tokens and verifying/parsing Clerk (Svix) webhooks. Neutral identity types
// live in internal/auth.
package clerk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
)

// SvixHeaders are the signature headers Clerk sends with each webhook.
type SvixHeaders struct {
	ID        string
	Timestamp string
	Signature string
}

// VerifySvix verifies a Clerk webhook signature (Svix scheme):
// base64(HMAC_SHA256(secretKey, "<id>.<timestamp>.<body>")).
func VerifySvix(secret string, h SvixHeaders, body []byte) error {
	if h.ID == "" || h.Timestamp == "" || h.Signature == "" {
		return errors.New("missing svix headers")
	}

	raw := strings.TrimPrefix(secret, "whsec_")
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("decode signing secret: %w", err)
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(h.ID + "." + h.Timestamp + "." + string(body)))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Signature header is a space-separated list of "v1,<sig>" entries.
	for _, entry := range strings.Fields(h.Signature) {
		sig := entry
		if i := strings.IndexByte(entry, ','); i >= 0 {
			sig = entry[i+1:]
		}
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return nil
		}
	}
	return errors.New("no matching signature")
}

// clerk webhook payload shapes (only the fields we use).
type webhookEnvelope struct {
	Type string      `json:"type"`
	Data webhookData `json:"data"`
}

type webhookData struct {
	ID                    string       `json:"id"`
	EmailAddresses        []clerkEmail `json:"email_addresses"`
	PrimaryEmailAddressID *string      `json:"primary_email_address_id"`
	FirstName             *string      `json:"first_name"`
	LastName              *string      `json:"last_name"`
	ImageURL              *string      `json:"image_url"`
}

type clerkEmail struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
}

// ParseWebhook extracts the event type and a ClerkIdentity from a webhook body.
func ParseWebhook(body []byte) (string, auth.ClerkIdentity, error) {
	var e webhookEnvelope
	if err := json.Unmarshal(body, &e); err != nil {
		return "", auth.ClerkIdentity{}, err
	}

	ident := auth.ClerkIdentity{
		ClerkUserID: e.Data.ID,
		Email:       primaryEmail(e.Data),
		Name:        fullName(e.Data),
		AvatarURL:   e.Data.ImageURL,
	}
	return e.Type, ident, nil
}

func primaryEmail(d webhookData) string {
	if d.PrimaryEmailAddressID != nil {
		for _, e := range d.EmailAddresses {
			if e.ID == *d.PrimaryEmailAddressID {
				return e.EmailAddress
			}
		}
	}
	if len(d.EmailAddresses) > 0 {
		return d.EmailAddresses[0].EmailAddress
	}
	return ""
}

func fullName(d webhookData) string {
	parts := make([]string, 0, 2)
	if d.FirstName != nil && *d.FirstName != "" {
		parts = append(parts, *d.FirstName)
	}
	if d.LastName != nil && *d.LastName != "" {
		parts = append(parts, *d.LastName)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
