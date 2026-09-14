package remittance

import (
	"testing"
	"time"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	body := []byte(`{"foo":"bar"}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	sig, err := Sign("secret-1", "POST", "/v1.0/remittance/quote", "core-api", body, timestamp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok, err := Verify([]string{"secret-1"}, "POST", "/v1.0/remittance/quote", "core-api", body, timestamp, sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("expected signature to verify against the same secret")
	}
}

func TestVerifyTriesEachSecretInOrder(t *testing.T) {
	body := []byte(`{}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	sig, err := Sign("previous-secret", "GET", "/v1.0/remittance/wallet/balance", "core-api", body, timestamp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok, err := Verify([]string{"current-secret", "previous-secret"}, "GET", "/v1.0/remittance/wallet/balance", "core-api", body, timestamp, sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("expected signature signed with the previous (rotated) secret to still verify")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	body := []byte(`{}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	sig, err := Sign("secret-1", "POST", "/v1.0/remittance/quote", "core-api", body, timestamp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok, err := Verify([]string{"wrong-secret"}, "POST", "/v1.0/remittance/quote", "core-api", body, timestamp, sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Fatal("expected signature signed with a different secret to fail verification")
	}
}

func TestVerifyRequestRejectsStaleTimestamp(t *testing.T) {
	body := []byte(`{}`)
	staleTimestamp := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	sig, err := Sign("secret-1", "POST", "/v1.0/remittance/quote", "core-api", body, staleTimestamp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	err = VerifyRequest([]string{"secret-1"}, 5*time.Minute, "POST", "/v1.0/remittance/quote", "core-api", body, staleTimestamp, sig)
	if err != ErrStaleTimestamp {
		t.Fatalf("expected ErrStaleTimestamp, got %v", err)
	}
}

func TestVerifyRequestRejectsInvalidSignature(t *testing.T) {
	body := []byte(`{}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	err := VerifyRequest([]string{"secret-1"}, 5*time.Minute, "POST", "/v1.0/remittance/quote", "core-api", body, timestamp, "not-a-real-signature")
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifyRequestAccepts(t *testing.T) {
	body := []byte(`{"amount":100}`)
	timestamp := time.Now().UTC().Format(time.RFC3339)

	sig, err := Sign("secret-1", "POST", "/v1.0/remittance/quote", "core-api", body, timestamp)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if err := VerifyRequest([]string{"secret-1"}, 5*time.Minute, "POST", "/v1.0/remittance/quote", "core-api", body, timestamp, sig); err != nil {
		t.Fatalf("expected VerifyRequest to accept a valid, fresh request, got %v", err)
	}
}
