package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	now := time.Now().UTC()

	// Act.
	raw, err := GenerateAccessToken("secret", 42, "admin", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	claims, err := ParseAccessToken("secret", raw)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" || claims.ID == "" || claims.ExpiresAt == nil {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestParseAccessTokenRejectsWrongSecret(t *testing.T) {
	raw, err := GenerateAccessToken("secret", 42, "user", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	// Act.
	_, parseErr := ParseAccessToken("wrong", raw)

	// Assert.
	if parseErr == nil {
		t.Fatal("expected parse error")
	}
}

func TestHashTokenValidation(t *testing.T) {
	tests := []struct {
		name    string
		pepper  string
		token   string
		purpose string
		want    string
	}{
		{name: "empty token", pepper: "pepper", token: "", purpose: "refresh", want: "empty token"},
		{name: "empty pepper", pepper: "", token: "token", purpose: "refresh", want: "empty token pepper"},
		{name: "empty purpose", pepper: "pepper", token: "token", purpose: "", want: "empty token purpose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			_, err := HashToken(tt.pepper, tt.token, tt.purpose)

			// Assert.
			if err == nil || err.Error() != tt.want {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestTokenHelpers(t *testing.T) {
	// Act.
	refreshHashA, err := HashRefreshToken("pepper", "token")
	if err != nil {
		t.Fatal(err)
	}
	refreshHashB, err := HashRefreshToken("pepper", "token")
	if err != nil {
		t.Fatal(err)
	}
	accessHash, err := HashAccessToken("pepper", "token")
	if err != nil {
		t.Fatal(err)
	}
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	csrfToken, err := GenerateCSRFToken()
	if err != nil {
		t.Fatal(err)
	}

	// Assert.
	if refreshHashA != refreshHashB || refreshHashA == accessHash {
		t.Fatalf("unexpected hashes: %s %s %s", refreshHashA, refreshHashB, accessHash)
	}
	if len(refreshToken) < 32 || len(csrfToken) < 32 {
		t.Fatalf("unexpected token lengths: %d %d", len(refreshToken), len(csrfToken))
	}
}

func TestParseAccessTokenRejectsInvalidMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			ID:        "id",
		},
	})
	raw, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	// Act.
	_, parseErr := ParseAccessToken("secret", raw)

	// Assert.
	if parseErr == nil || !strings.Contains(parseErr.Error(), "invalid signing method") {
		t.Fatalf("unexpected error: %v", parseErr)
	}
}
