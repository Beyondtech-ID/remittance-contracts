# remittance-contracts

Shared wire contract between core-api and any remittance backbone service
it talks to (topremit-backbone today, more providers/backbones later).

One Go module both sides import — request/response DTOs, the HMAC signing
scheme, canonical error codes, and endpoint paths live here once, instead
of being hand-mirrored (and drifting) between repos.

See the package doc comment in `doc.go` for the full rationale, adoption
guide, and versioning policy. Short version:

```go
client := remittance.NewClient(remittance.Config{
	BaseURL:          "https://topremit-backbone.internal",
	ServiceID:        "core-api",
	SecretKeyCurrent: secret,
})

resp, err := client.Submit(ctx, remittance.SubmitRequest{
	QuotationID:         quotationID,
	RecipientID:         recipientID,
	ClientTransactionID: idempotencyKey,
	PurposeCode:         "FAMILY_SUPPORT",
	FundSource:          "SALARY",
})
```

Adding a new backbone/provider later: point a new `Config` at it. No new
DTOs, no new mapping code, as long as the backbone implements this same
contract (same DTO shapes, same envelope, same signing scheme — see
`doc.go`'s versioning policy for what "same" means and how to extend it
without breaking anyone already on v1).

## Layout

| File | What |
|---|---|
| `dto.go` | Request/response structs |
| `status.go` | `OrderStatus` / `CallbackStatus` enums |
| `codes.go` | Canonical `Code` error-code constants |
| `envelope.go` | `BaseResponse` (SNAP-style envelope) + code parsing |
| `signing.go` | `Sign` / `Verify` / `VerifyRequest` (HMAC-SHA512) |
| `client.go` | Stdlib-only HTTP client implementing the contract |
| `paths.go` | Canonical endpoint path constants |
| `errors.go` | `APIError` returned by `Client` on non-2xx |

## Test

```bash
go build ./...
go vet ./...
go test ./...
```
