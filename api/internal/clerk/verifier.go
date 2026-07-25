package clerk

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
)

// Verifier verifies Clerk session JWTs and resolves the user's identity.
// Construct with NewVerifier only when a Clerk secret key is configured.
type Verifier struct{}

// NewVerifier sets the Clerk API key and returns a token verifier.
func NewVerifier(secretKey string) *Verifier {
	sdk.SetKey(secretKey)
	return &Verifier{}
}

// Verify validates a Clerk session token and returns the mirrored identity.
func (v *Verifier) Verify(ctx context.Context, token string) (auth.ClerkIdentity, error) {
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token})
	if err != nil {
		return auth.ClerkIdentity{}, fmt.Errorf("verify session token: %w", err)
	}
	if claims.Subject == "" {
		return auth.ClerkIdentity{}, fmt.Errorf("token has no subject")
	}

	usr, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return auth.ClerkIdentity{}, fmt.Errorf("fetch clerk user: %w", err)
	}
	return identityFromUser(usr), nil
}

func identityFromUser(u *sdk.User) auth.ClerkIdentity {
	ident := auth.ClerkIdentity{
		ClerkUserID: u.ID,
		AvatarURL:   u.ImageURL,
	}

	if u.PrimaryEmailAddressID != nil {
		for _, e := range u.EmailAddresses {
			if e != nil && e.ID == *u.PrimaryEmailAddressID {
				ident.Email = e.EmailAddress
				break
			}
		}
	}
	if ident.Email == "" && len(u.EmailAddresses) > 0 && u.EmailAddresses[0] != nil {
		ident.Email = u.EmailAddresses[0].EmailAddress
	}

	var parts []string
	if u.FirstName != nil && *u.FirstName != "" {
		parts = append(parts, *u.FirstName)
	}
	if u.LastName != nil && *u.LastName != "" {
		parts = append(parts, *u.LastName)
	}
	ident.Name = strings.TrimSpace(strings.Join(parts, " "))

	return ident
}
