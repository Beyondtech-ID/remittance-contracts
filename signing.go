package remittance

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrStaleTimestamp / ErrInvalidSignature are returned by VerifyRequest. A
// caller can match on these with errors.Is instead of re-deriving the
// distinction itself.
var (
	ErrStaleTimestamp   = errors.New("remittance: timestamp outside allowed window")
	ErrInvalidSignature = errors.New("remittance: invalid signature")
)

// Headers is the full set of values a signed request/callback under this
// contract must carry.
type Headers struct {
	ServiceID string
	Timestamp string
	Signature string
}

// Sign computes the HMAC-SHA512 signature this contract uses in both
// directions: core-api signing a request toward a backbone, and a backbone
// signing a StatusCallback toward core-api. Same recipe, opposite
// direction, different secret per direction.
//
//	toSign    = METHOD + ":" + PATH + ":" + SERVICE_ID + ":" + sha256_hex(compact(body)) + ":" + TIMESTAMP
//	signature = base64(HMAC_SHA512(toSign, secretKey))
//
// PATH must be the request path only (no scheme/host, no query string) —
// this mirrors topremit-backbone's own hmac_signature.go exactly.
func Sign(secretKey, method, path, serviceID string, body []byte, timestamp string) (string, error) {
	hashedBody, err := hashBody(body)
	if err != nil {
		return "", fmt.Errorf("compact request body: %w", err)
	}
	toSign := method + ":" + path + ":" + serviceID + ":" + hashedBody + ":" + timestamp
	h := hmac.New(sha512.New, []byte(secretKey))
	if _, err := h.Write([]byte(toSign)); err != nil {
		return "", fmt.Errorf("hash data: %w", err)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// Verify checks sig against every key in secretKeys, in order, using a
// constant-time comparison, returning true on the first match. Pass both a
// current and previous secret during a key-rotation window.
func Verify(secretKeys []string, method, path, serviceID string, body []byte, timestamp, sig string) (bool, error) {
	for _, key := range secretKeys {
		expected, err := Sign(key, method, path, serviceID, body, timestamp)
		if err != nil {
			return false, err
		}
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return true, nil
		}
	}
	return false, nil
}

// BuildHeaders signs the request and returns the full header set to send.
func BuildHeaders(secretKey, serviceID, method, path string, body []byte, timestamp string) (Headers, error) {
	sig, err := Sign(secretKey, method, path, serviceID, body, timestamp)
	if err != nil {
		return Headers{}, err
	}
	return Headers{ServiceID: serviceID, Timestamp: timestamp, Signature: sig}, nil
}

// VerifyRequest is the one-call server-side check: parses timestamp,
// rejects it if outside window (window <= 0 disables the check), then
// verifies sig against secretKeys. A backbone's HTTP middleware (whatever
// framework it uses) wraps this — reading the request body/headers and
// mapping the returned error to its own HTTP response is still the
// framework-specific part this function deliberately stays out of.
func VerifyRequest(secretKeys []string, window time.Duration, method, path, serviceID string, body []byte, timestamp, sig string) error {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("%w: malformed timestamp: %v", ErrInvalidSignature, err)
	}
	if window > 0 {
		skew := time.Since(ts)
		if skew < 0 {
			skew = -skew
		}
		if skew > window {
			return ErrStaleTimestamp
		}
	}
	ok, err := Verify(secretKeys, method, path, serviceID, body, timestamp, sig)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidSignature
	}
	return nil
}

// hashBody compacts the JSON body and returns its lower-hex SHA-256 digest.
func hashBody(body []byte) (string, error) {
	if len(body) == 0 {
		return "", nil
	}
	dst := &bytes.Buffer{}
	if err := json.Compact(dst, body); err != nil {
		return "", err
	}
	return strings.ToLower(fmt.Sprintf("%x", sha256.Sum256(dst.Bytes()))), nil
}
