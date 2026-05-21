package auth

import (
	"strings"
	"testing"
	"time"
)

func TestTokenManager_GeraEValidaJWTAssinadoHS256(t *testing.T) {
	manager := NewTokenManager("segredo-de-teste-com-mais-de-32-bytes", func() time.Time {
		return time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	})

	token, expiresAt, err := manager.GenerateAccessToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("token = %q, want JWT", token)
	}
	if !expiresAt.Equal(time.Date(2026, 5, 21, 20, 0, 0, 0, time.UTC)) {
		t.Fatalf("expiresAt = %s", expiresAt)
	}

	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "admin" {
		t.Fatalf("subject = %q", claims.Subject)
	}
}

func TestTokenManager_RejeitaJWTComAssinaturaAlterada(t *testing.T) {
	manager := NewTokenManager("segredo-de-teste-com-mais-de-32-bytes", func() time.Time {
		return time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	})
	token, _, err := manager.GenerateAccessToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := manager.ValidateAccessToken(tampered); err == nil {
		t.Fatal("ValidateAccessToken aceitou token adulterado")
	}
}

func TestRefreshToken_GeraTokenOpacoEHashDeterministico(t *testing.T) {
	token, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || strings.Count(token, ".") == 2 {
		t.Fatalf("refresh token = %q, want opaque token", token)
	}

	hash := HashRefreshToken(token)
	if hash == "" || hash == token {
		t.Fatalf("hash = %q, token = %q", hash, token)
	}
	if HashRefreshToken(token) != hash {
		t.Fatal("HashRefreshToken nao e deterministico")
	}
}
