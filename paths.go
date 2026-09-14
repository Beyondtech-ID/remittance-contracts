package remittance

// Path* constants are the exact HTTP paths this contract's money-movement
// surface uses. Both Client (caller side) and a backbone's own router
// (server side) should reference these constants directly instead of
// re-typing the literal string — that's how a path rename (like
// /v1.0/remittance/submit becoming /v1.0/remittance/transactions once
// already) gets caught at compile time on both sides instead of silently
// drifting.
const (
	PathTransactions           = "/v1.0/remittance/transactions"
	PathStatusPrefix           = "/v1.0/remittance/status/" // + partnerRef
	PathQuote                  = "/v1.0/remittance/quote"
	PathRates                  = "/v1.0/remittance/rates"
	PathSender                 = "/v1.0/remittance/sender"
	PathRecipient              = "/v1.0/remittance/recipient"
	PathRecipientPrefix        = "/v1.0/remittance/recipient/" // + id
	PathRecipientVerifyAccount = "/v1.0/remittance/recipient/verify-account"
	PathWalletBalance          = "/v1.0/remittance/wallet/balance"

	// PathCoreCallback is the path a backbone POSTs StatusCallback to on
	// core-api, relative to core-api's own BaseURL.
	PathCoreCallback = "/v1.0/backbone/callback-remittance-status"
)
