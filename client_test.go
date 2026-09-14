package remittance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSubmitSendsSignedRequestAndDecodesData(t *testing.T) {
	const secret = "test-secret"
	const serviceID = "core-api-test"

	var gotPath, gotServiceID, gotTimestamp, gotSignature string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotServiceID = r.Header.Get("X-SERVICE-ID")
		gotTimestamp = r.Header.Get("X-TIMESTAMP")
		gotSignature = r.Header.Get("X-SIGNATURE")
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		gotBody = body

		resp := NewBaseResponse(SubmitResponse{
			ClientTransactionID: "ctx-1",
			BackboneRef:         "bb-ref-1",
			Status:              OrderStatusProcessing,
		}, "OK", http.StatusOK, "RM", "00")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(Config{
		BaseURL:          srv.URL,
		ServiceID:        serviceID,
		SecretKeyCurrent: secret,
	})

	req := SubmitRequest{
		QuotationID:         "q-1",
		RecipientID:         "r-1",
		ClientTransactionID: "ctx-1",
		PurposeCode:         "FAMILY_SUPPORT",
		FundSource:          "SALARY",
	}

	resp, err := client.Submit(context.Background(), req)
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if resp.BackboneRef != "bb-ref-1" || resp.Status != OrderStatusProcessing {
		t.Fatalf("unexpected response: %+v", resp)
	}

	if gotPath != PathTransactions {
		t.Fatalf("expected path %s, got %s", PathTransactions, gotPath)
	}
	if gotServiceID != serviceID {
		t.Fatalf("expected X-SERVICE-ID %s, got %s", serviceID, gotServiceID)
	}
	if gotSignature == "" || gotTimestamp == "" {
		t.Fatal("expected signature/timestamp headers to be set")
	}

	ok, err := Verify([]string{secret}, http.MethodPost, PathTransactions, serviceID, gotBody, gotTimestamp, gotSignature)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("server-observed signature does not verify against the secret the client signed with")
	}
}

func TestClientDecodesAPIErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := NewBaseResponse(nil, "Invalid request signature", http.StatusUnauthorized, "RM", "10007")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, ServiceID: "core-api-test", SecretKeyCurrent: "secret"})

	_, err := client.Quote(context.Background(), QuoteRequest{Country: "PHL", Service: "BankAccount", Currency: "PHP"})
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("expected HTTPStatus 401, got %d", apiErr.HTTPStatus)
	}
	if apiErr.Code != CodeInvalidSignature {
		t.Fatalf("expected Code %s, got %s", CodeInvalidSignature, apiErr.Code)
	}
}

func TestClientRetriesOn5xxThenSucceeds(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		resp := NewBaseResponse(WalletBalanceResponse{{Balance: "1000.00", Currency: "IDR"}}, "OK", http.StatusOK, "RM", "00")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, ServiceID: "core-api-test", SecretKeyCurrent: "secret"})

	balances, err := client.GetWalletBalance(context.Background())
	if err != nil {
		t.Fatalf("GetWalletBalance: %v", err)
	}
	if len(balances) != 1 || balances[0].Currency != "IDR" {
		t.Fatalf("unexpected balances: %+v", balances)
	}
	if attempts != 2 {
		t.Fatalf("expected exactly 2 attempts (1 failed 5xx + 1 success), got %d", attempts)
	}
}
