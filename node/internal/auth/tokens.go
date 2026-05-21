package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	AccessTokenTTL  = 8 * time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour
)

var (
	ErrInvalidToken = errors.New("token invalido")
	ErrExpiredToken = errors.New("token expirado")
)

type TokenManager struct {
	secret []byte
	now    func() time.Time
}

type Claims struct {
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Type      string `json:"typ"`
}

func NewTokenManager(secret string, now func() time.Time) TokenManager {
	if now == nil {
		now = time.Now
	}
	return TokenManager{secret: []byte(secret), now: now}
}

func (m TokenManager) GenerateAccessToken(subject string) (string, time.Time, error) {
	if len(m.secret) == 0 || strings.TrimSpace(subject) == "" {
		return "", time.Time{}, ErrInvalidToken
	}
	now := m.now().UTC()
	expiresAt := now.Add(AccessTokenTTL)
	header := jwtHeader{Algorithm: "HS256", Type: "JWT"}
	claims := jwtClaims{
		Subject:   subject,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Type:      "access",
	}

	headerSegment, err := encodeJSONSegment(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimsSegment, err := encodeJSONSegment(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	signingInput := headerSegment + "." + claimsSegment
	signature := sign(signingInput, m.secret)
	return signingInput + "." + signature, expiresAt, nil
}

func (m TokenManager) ValidateAccessToken(token string) (Claims, error) {
	if len(m.secret) == 0 {
		return Claims{}, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	want := sign(signingInput, m.secret)
	if !hmac.Equal([]byte(want), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}

	var header jwtHeader
	if err := decodeJSONSegment(parts[0], &header); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return Claims{}, ErrInvalidToken
	}

	var rawClaims jwtClaims
	if err := decodeJSONSegment(parts[1], &rawClaims); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if rawClaims.Subject == "" || rawClaims.Type != "access" || rawClaims.ExpiresAt == 0 {
		return Claims{}, ErrInvalidToken
	}
	expiresAt := time.Unix(rawClaims.ExpiresAt, 0).UTC()
	if !expiresAt.After(m.now().UTC()) {
		return Claims{}, ErrExpiredToken
	}
	return Claims{
		Subject:   rawClaims.Subject,
		IssuedAt:  time.Unix(rawClaims.IssuedAt, 0).UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func NewRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("gerar refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func encodeJSONSegment(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeJSONSegment(segment string, target any) error {
	payload, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func sign(signingInput string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
