package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	totpStepSeconds = int64(30)
	totpDigits      = 6
	totpWindow      = int64(1)
)

// GenerateTOTPCodeAt returns a 6-digit RFC 6238 TOTP code using HMAC-SHA1.
func GenerateTOTPCodeAt(secret string, at time.Time) (string, error) {
	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return "", err
	}
	counter := uint64(at.UTC().Unix() / totpStepSeconds)
	return hotpCode(key, counter), nil
}

func VerifyTOTPCode(secret, code string, at time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}

	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return false
	}
	counter := at.UTC().Unix() / totpStepSeconds
	for offset := -totpWindow; offset <= totpWindow; offset++ {
		if counter+offset < 0 {
			continue
		}
		want := hotpCode(key, uint64(counter+offset))
		if hmac.Equal([]byte(want), []byte(code)) {
			return true
		}
	}
	return false
}

func decodeTOTPSecret(secret string) ([]byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(secret))
	normalized = strings.ReplaceAll(normalized, " ", "")
	if normalized == "" {
		return nil, fmt.Errorf("totp secret vazio")
	}
	if padding := len(normalized) % 8; padding != 0 {
		normalized += strings.Repeat("=", 8-padding)
	}
	return base32.StdEncoding.DecodeString(normalized)
}

func hotpCode(key []byte, counter uint64) string {
	var payload [8]byte
	binary.BigEndian.PutUint64(payload[:], counter)

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(payload[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", binCode%1000000)
}
