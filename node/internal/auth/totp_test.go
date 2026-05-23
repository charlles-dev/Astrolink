package auth

import (
	"testing"
	"time"
)

func TestGenerateTOTPCodeAt_RFC6238SHA1Vector(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

	code, err := GenerateTOTPCodeAt(secret, time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateTOTPCodeAt() error = %v", err)
	}

	if code != "287082" {
		t.Fatalf("code = %q, want 287082", code)
	}
}

func TestVerifyTOTPCode_AcceptsAdjacentStepWindow(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	code, err := GenerateTOTPCodeAt(secret, time.Unix(60, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateTOTPCodeAt() error = %v", err)
	}

	if !VerifyTOTPCode(secret, code, time.Unix(89, 0).UTC()) {
		t.Fatal("VerifyTOTPCode() = false, want true for previous step in window")
	}
}

func TestVerifyTOTPCode_RejectsInvalidCode(t *testing.T) {
	if VerifyTOTPCode("JBSWY3DPEHPK3PXP", "000000", time.Unix(60, 0).UTC()) {
		t.Fatal("VerifyTOTPCode() = true, want false")
	}
}
