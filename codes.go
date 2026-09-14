package remittance

// Code identifies a specific error case, in the same "SERVICE-CASE" shape
// topremit-backbone's errorContract.json keys use (e.g. "RM-10007"). It is
// a plain string type, not a closed enum on purpose: a new backbone is free
// to define its own additional Code constants (under its own service
// prefix) in its own package without needing a change here or any risk of
// colliding with the codes below.
type Code string

const (
	// CodeInvalidRequest: generic malformed-request error, no
	// service-specific code applies.
	CodeInvalidRequest Code = "10001"

	// CodeLoginFailed / CodeUserNotFound: carried over from the shared
	// error-contract convention core-api and its backbones both use;
	// not specific to remittance but part of the same code space.
	CodeLoginFailed  Code = "AU-10001"
	CodeUserNotFound Code = "UM-10001"

	// RM-1000x: remittance-domain error codes, exactly as defined in
	// topremit-backbone's errorContract.json.
	CodeRateNotFound            Code = "RM-10001"
	CodeBranchIDRequired        Code = "RM-10002"
	CodeCorridorNotSupported    Code = "RM-10003"
	CodeInvalidIdempotencyKey   Code = "RM-10004"
	CodePartnerExecutionFailed  Code = "RM-10005"
	CodePartnerTimeout          Code = "RM-10006"
	CodeInvalidSignature        Code = "RM-10007"
	CodeStaleTimestamp          Code = "RM-10008"
	CodeDuplicateSubmission     Code = "RM-10009"
	CodeKYCPayloadInvalid       Code = "RM-10010"
	CodeWebhookNotConfigured    Code = "RM-10011"
	CodeWebhookProcessingFailed Code = "RM-10012"
	CodePartnerResourceNotFound Code = "RM-10013"
)
