package remittance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Config holds the outbound connection settings toward a backbone
// implementing this contract. SecretKeyPrevious is optional — set it
// during a secret-rotation window so a request signed with the not-yet-
// rotated secret still isn't needed (Client always signs with
// SecretKeyCurrent; SecretKeyPrevious exists here only for symmetry with
// the backbone side, which tries both when verifying inbound requests).
type Config struct {
	BaseURL          string
	ServiceID        string
	SecretKeyCurrent string
	Timeout          time.Duration
	// HTTPClient overrides the client used to send requests. Leave nil to
	// get a private *http.Client{Timeout: Timeout} (default 30s).
	HTTPClient *http.Client
}

// Client is a minimal, dependency-free (stdlib-only) HTTP client for any
// backbone implementing this contract. Switching to a different
// remittance backbone/provider means pointing a new Config at it — the
// calling code and every DTO stay the same.
type Client struct {
	cfg Config
	hc  *http.Client
}

func NewClient(cfg Config) *Client {
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	return &Client{cfg: cfg, hc: hc}
}

const clientMaxAttempts = 3

// do signs and sends one request, retrying up to clientMaxAttempts times on
// network errors or 5xx responses only (never on 4xx — those are the
// caller's fault, retrying won't change the outcome). On success it decodes
// envelope.ResponseData into out (if out is non-nil); on a non-2xx response
// it returns *APIError.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature, err := Sign(c.cfg.SecretKeyCurrent, method, path, c.cfg.ServiceID, bodyBytes, timestamp)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	reqURL := c.cfg.BaseURL + path

	var lastErr error
	for attempt := 1; attempt <= clientMaxAttempts; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, buildErr := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
		if buildErr != nil {
			return fmt.Errorf("build request: %w", buildErr)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-SERVICE-ID", c.cfg.ServiceID)
		req.Header.Set("X-TIMESTAMP", timestamp)
		req.Header.Set("X-SIGNATURE", signature)

		resp, doErr := c.hc.Do(req)
		if doErr != nil {
			lastErr = doErr
			if attempt < clientMaxAttempts {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			return fmt.Errorf("request failed after %d attempts: %w", clientMaxAttempts, lastErr)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}

		if resp.StatusCode >= 500 && attempt < clientMaxAttempts {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		var envelope BaseResponse
		if len(respBody) > 0 {
			if jerr := json.Unmarshal(respBody, &envelope); jerr != nil {
				return fmt.Errorf("decode response envelope (http %d): %w", resp.StatusCode, jerr)
			}
		}

		if resp.StatusCode >= 300 {
			apiErr := &APIError{HTTPStatus: resp.StatusCode, Message: envelope.ResponseMessage}
			if sc, cc, ok := ParseResponseCode(resp.StatusCode, envelope.ResponseCode); ok {
				apiErr.Code = Code(sc + "-" + cc)
			}
			return apiErr
		}

		if out != nil && envelope.ResponseData != nil {
			raw, merr := json.Marshal(envelope.ResponseData)
			if merr != nil {
				return fmt.Errorf("re-marshal response data: %w", merr)
			}
			if uerr := json.Unmarshal(raw, out); uerr != nil {
				return fmt.Errorf("decode response data: %w", uerr)
			}
		}
		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", clientMaxAttempts, lastErr)
}

func (c *Client) Submit(ctx context.Context, req SubmitRequest) (*SubmitResponse, error) {
	var out SubmitResponse
	if err := c.do(ctx, http.MethodPost, PathTransactions, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) StatusCheck(ctx context.Context, partnerRef string) (*StatusResponse, error) {
	var out StatusResponse
	if err := c.do(ctx, http.MethodGet, PathStatusPrefix+partnerRef, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Quote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	var out QuoteResponse
	if err := c.do(ctx, http.MethodPost, PathQuote, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRates looks up a corridor's reference fx rate/fee. serviceID selects
// which Topremit payment method to probe (see the ServiceX constants below)
// — different services aren't just different delivery speeds, some
// corridors only exist under a specific service (e.g. China only quotes
// under ServiceWeChatPay/ServiceAlipay, never ServiceBankAccount). Pass ""
// to let the backbone use its own default.
func (c *Client) GetRates(ctx context.Context, originCountry, targetCountry, serviceID string) (*RateResponse, error) {
	q := url.Values{}
	q.Set("originCountry", originCountry)
	q.Set("targetCountry", targetCountry)
	if serviceID != "" {
		q.Set("serviceId", serviceID)
	}
	var out RateResponse
	if err := c.do(ctx, http.MethodGet, PathRates+"?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSender(ctx context.Context, req CreateSenderRequest) (*CreateSenderResponse, error) {
	var out CreateSenderResponse
	if err := c.do(ctx, http.MethodPost, PathSender, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateRecipient(ctx context.Context, req CreateRecipientRequest) (*CreateRecipientResponse, error) {
	var out CreateRecipientResponse
	if err := c.do(ctx, http.MethodPost, PathRecipient, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateRecipient(ctx context.Context, id string, req UpdateRecipientRequest) (*CreateRecipientResponse, error) {
	var out CreateRecipientResponse
	if err := c.do(ctx, http.MethodPut, PathRecipientPrefix+id, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) VerifyRecipientAccount(ctx context.Context, req VerifyRecipientAccountRequest) (*VerifyRecipientAccountResponse, error) {
	var out VerifyRecipientAccountResponse
	if err := c.do(ctx, http.MethodPost, PathRecipientVerifyAccount, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetWalletBalance(ctx context.Context) (WalletBalanceResponse, error) {
	var out WalletBalanceResponse
	if err := c.do(ctx, http.MethodGet, PathWalletBalance, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
