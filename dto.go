package remittance

// PartyPayload carries KYC identity for either the sender or the
// beneficiary of a remittance order. Fields cover what Topremit's own
// sender/beneficiary APIs require today (see the field comments); a
// provider that needs something not modeled here should put it in Extra
// rather than wait for a new named field.
type PartyPayload struct {
	Name           string `json:"name"`
	IdentityType   string `json:"identityType,omitempty"` // e.g. NIK, PASSPORT
	IdentityNumber string `json:"identityNumber"`
	DateOfBirth    string `json:"dateOfBirth,omitempty"` // YYYY-MM-DD
	Nationality    string `json:"nationality"`
	Address        string `json:"address,omitempty"`
	Phone          string `json:"phone,omitempty"`
	BankCode       string `json:"bankCode,omitempty"`
	AccountNumber  string `json:"accountNumber,omitempty"`
	Province       string `json:"province,omitempty"`
	City           string `json:"city,omitempty"`
	PostalCode     string `json:"postalCode,omitempty"`
	IDCountry      string `json:"idCountry,omitempty"`
	IDExpiryDate   string `json:"idExpiryDate,omitempty"`
	CountryOfBirth string `json:"countryOfBirth,omitempty"`
	Occupation     string `json:"occupation,omitempty"`
	// DateOfEstablished/EstablishedCountry: required when Segment is
	// "BUSINESS" instead of "PERSONAL".
	DateOfEstablished  string         `json:"dateOfEstablished,omitempty"`
	EstablishedCountry string         `json:"establishedCountry,omitempty"`
	Extra              map[string]any `json:"extra,omitempty"`
}

// SubmitRequest is the body of POST /v1.0/remittance/transactions. The
// caller is expected to have already obtained quotationId/recipientId (and
// senderId, if the provider needs one) from this contract's Quote/
// CreateSender/CreateRecipient calls — Submit executes against a price the
// caller already saw and approved, it does not recompute pricing.
type SubmitRequest struct {
	QuotationID string `json:"quotationId"`
	RecipientID string `json:"recipientId"`
	SenderID    string `json:"senderId,omitempty"`
	// ClientTransactionID doubles as the idempotency key: resubmitting the
	// same value replays the existing order instead of creating a new one.
	ClientTransactionID string         `json:"clientTransactionId"`
	PurposeCode         string         `json:"purposeCode"`
	FundSource          string         `json:"fundSource"`
	Description         string         `json:"description,omitempty"`
	BranchID            *string        `json:"branchId,omitempty"`
	Extra               map[string]any `json:"extra,omitempty"`
}

// SubmitResponse is returned synchronously by Submit. Execution against the
// provider is asynchronous, so Status here is typically
// OrderStatusProcessing — the terminal outcome arrives later via
// StatusCallback or a StatusCheck poll.
type SubmitResponse struct {
	ClientTransactionID   string         `json:"clientTransactionId"`
	BackboneRef           string         `json:"backboneRef"`
	Status                OrderStatus    `json:"status"`
	PartnerTransactionRef string         `json:"partnerTransactionRef,omitempty"`
	Extra                 map[string]any `json:"extra,omitempty"`
}

// StatusResponse is returned by GET /v1.0/remittance/status/:partnerRef —
// the poll-based fallback lane for when a StatusCallback is delayed or lost.
// PartnerRef is the same value as SubmitRequest.ClientTransactionID — the
// backbone's own idempotency key, just addressed under a different field
// name once it comes back out of the backbone (matches
// topremit-backbone's actual wire shape; not renamed here to stay in sync
// with it).
type StatusResponse struct {
	PartnerRef  string         `json:"partnerRef"`
	BackboneRef string         `json:"backboneRef"`
	Status      OrderStatus    `json:"status"`
	UpdatedAt   string         `json:"updatedAt"`
	Extra       map[string]any `json:"extra,omitempty"`
}

// QuoteRequest is the body of POST /v1.0/remittance/quote. Amount defaults
// to 1 (of Currency) if omitted — this is a corridor-availability +
// indicative-rate probe, not a price lock. Note Service here and
// QuoteResponse.ServiceID are the same concept (a payment-method code) —
// named differently because that's how topremit-backbone's real request
// vs. response fields are actually spelled; not unified here to avoid
// drifting from its wire shape.
type QuoteRequest struct {
	Country        string         `json:"country"`
	Service        string         `json:"service"`
	Currency       string         `json:"currency"`
	RoutingChannel string         `json:"routingChannel,omitempty"` // e.g. INSTANT, REGULAR
	Amount         float64        `json:"amount,omitempty"`
	IsSendAmount   bool           `json:"isSendAmount"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// MoneyAmount is a source/destination money value. Amount stays a string
// (matches how Topremit itself represents it) rather than being normalized
// to a numeric type.
type MoneyAmount struct {
	Country  string `json:"country,omitempty"` // only meaningful on a destination amount
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// Fee is a fee amount attached to a quote or rate.
type Fee struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// QuoteResponse is the body of a successful Quote call.
type QuoteResponse struct {
	RefID          string         `json:"refId"`
	ExpirationDate string         `json:"expirationDate"`
	ServiceID      string         `json:"serviceId"`
	RoutingChannel string         `json:"routingChannel,omitempty"`
	FxRate         float64        `json:"fxRate"`
	Source         MoneyAmount    `json:"source"`
	Destination    MoneyAmount    `json:"destination"`
	Fee            Fee            `json:"fee"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// RateFee is one routing channel's fee for a rate lookup — INSTANT and
// REGULAR (and any other channel a provider supports) commonly charge
// different fees for the same corridor, so RateResponse carries one
// RateFee per channel rather than a single fee.
type RateFee struct {
	RoutingChannel string  `json:"routingChannel"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
}

// RateResponse is returned by
// GET /v1.0/remittance/rates?originCountry=&targetCountry= — a cached,
// reference-only quotation. It is NOT a frozen/lockable price: Submit still
// requires the caller to have obtained a fresh QuotationID via Quote.
// FxRate applies across every channel; Fee does not, so it stays a list
// keyed by RoutingChannel.
type RateResponse struct {
	OriginCountry string `json:"originCountry"`
	TargetCountry string `json:"targetCountry"`
	// ServiceID is the payment method this rate was probed under (see the
	// ServiceX constants) — echoed back so the caller can tell which
	// service a given fee list belongs to without having to remember what
	// it asked for.
	ServiceID      string         `json:"serviceId,omitempty"`
	Currency       string         `json:"currency"`
	FxRate         float64        `json:"fxRate"`
	ExpirationDate string         `json:"expirationDate"`
	Fee            []RateFee      `json:"fee"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// CreateSenderRequest is the body of POST /v1.0/remittance/sender. Sender
// identity is generally "create once, reuse" — a backbone is expected to
// cache and return the same PartnerSenderID for a repeated identical
// identity rather than creating a duplicate upstream.
type CreateSenderRequest struct {
	Party      PartyPayload   `json:"party"`
	Segment    string         `json:"segment,omitempty"` // e.g. PERSONAL, BUSINESS — default PERSONAL
	FundSource string         `json:"fundSource"`
	Extra      map[string]any `json:"extra,omitempty"`
}

type CreateSenderResponse struct {
	PartnerSenderID string         `json:"partnerSenderId"`
	Extra           map[string]any `json:"extra,omitempty"`
}

// RecipientBeneficiary carries the beneficiary side of a recipient record.
type RecipientBeneficiary struct {
	FullName      string         `json:"fullName"`
	IDType        string         `json:"idType,omitempty"`
	IDNumber      string         `json:"idNumber,omitempty"`
	ContactNumber string         `json:"contactNumber,omitempty"`
	Address       string         `json:"address,omitempty"`
	City          string         `json:"city,omitempty"`
	PostalCode    string         `json:"postalCode,omitempty"`
	Email         string         `json:"email,omitempty"`
	Province      string         `json:"province,omitempty"`
	Regency       string         `json:"regency,omitempty"`
	Relationship  string         `json:"relationship,omitempty"`
	Segment       string         `json:"segment,omitempty"`
	DateOfBirth   string         `json:"dateOfBirth,omitempty"`
	Gender        string         `json:"gender,omitempty"`
	Nationality   string         `json:"nationality,omitempty"`
	IDCountry     string         `json:"idCountry,omitempty"`
	IDExpiryDate  string         `json:"idExpiryDate,omitempty"`
	Extra         map[string]any `json:"extra,omitempty"`
}

// CreditParty carries the payout-destination side of a recipient record.
// Shaped around a bank-account payout today (Topremit is the only
// integrated provider): AccountNumber is optional here, not because it's
// unimportant, but because a provider whose payout identifier isn't
// literally a bank account number (a mobile-wallet number, a cash-pickup
// code, ...) should put that identifier in Extra instead of forcing it
// into AccountNumber. A backbone integrating a bank-account-based
// provider still treats AccountNumber as required on its own side.
type CreditParty struct {
	AccountNumber  string         `json:"accountNumber,omitempty"`
	BankProvince   string         `json:"bankProvince,omitempty"`
	BankCity       string         `json:"bankCity,omitempty"`
	BankBranchName string         `json:"bankBranchName,omitempty"`
	BankBranchCode string         `json:"bankBranchCode,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// CreateRecipientRequest is the body of POST /v1.0/remittance/recipient.
// PayerID selects which payer/bank network receives the money — required
// for a provider (like Topremit) that models payout through a discrete
// payer selection step; optional here because a provider without that
// concept (routing payout by CreditParty/ServiceID alone) simply omits it.
// A backbone integrating a payer-based provider still requires it on its
// own side — it must never guess which payer/bank receives the money.
type CreateRecipientRequest struct {
	PayerID                    string               `json:"payerId,omitempty"`
	DestinationCurrencyISOCode string               `json:"destinationCurrencyIsoCode"`
	DestinationCountryISOCode  string               `json:"destinationCountryIsoCode"`
	ServiceID                  string               `json:"serviceId"`
	Beneficiary                RecipientBeneficiary `json:"beneficiary"`
	CreditParty                CreditParty          `json:"creditParty"`
	Extra                      map[string]any       `json:"extra,omitempty"`
}

type CreateRecipientResponse struct {
	ID                         string               `json:"id"`
	PayerID                    string               `json:"payerId,omitempty"`
	DestinationCurrencyISOCode string               `json:"destinationCurrencyIsoCode"`
	DestinationCountryISOCode  string               `json:"destinationCountryIsoCode"`
	ServiceID                  string               `json:"serviceId"`
	Beneficiary                RecipientBeneficiary `json:"beneficiary"`
	CreditParty                CreditParty          `json:"creditParty"`
	Extra                      map[string]any       `json:"extra,omitempty"`
}

// UpdateRecipientRequest is the body of PUT /v1.0/remittance/recipient/:id.
// See CreateRecipientRequest for why PayerID is optional here.
type UpdateRecipientRequest struct {
	PayerID                    string               `json:"payerId,omitempty"`
	DestinationCurrencyISOCode string               `json:"destinationCurrencyIsoCode"`
	DestinationCountryISOCode  string               `json:"destinationCountryIsoCode"`
	ServiceID                  string               `json:"serviceId"`
	Beneficiary                RecipientBeneficiary `json:"beneficiary"`
	CreditParty                CreditParty          `json:"creditParty"`
	Extra                      map[string]any       `json:"extra,omitempty"`
}

// VerifyRecipientAccountRequest is the body of
// POST /v1.0/remittance/recipient/verify-account.
type VerifyRecipientAccountRequest struct {
	Country string `json:"country"`
	Type    string `json:"type"` // e.g. BANK_ACCOUNT, E_WALLET
	// PayerID: see CreateRecipientRequest for why this is optional.
	PayerID     string `json:"payerId,omitempty"`
	CreditParty struct {
		AccountNumber string `json:"accountNumber,omitempty"`
	} `json:"creditParty"`
	Beneficiary struct {
		FullName string `json:"fullName,omitempty"`
	} `json:"beneficiary,omitempty"`
	Extra map[string]any `json:"extra,omitempty"`
}

type VerifyRecipientAccountResponse struct {
	Status      string `json:"status"`
	Beneficiary struct {
		FullName string `json:"fullName"`
	} `json:"beneficiary"`
	Extra map[string]any `json:"extra,omitempty"`
}

// WalletBalanceEntry is one entry of the array returned by
// GET /v1.0/remittance/wallet/balance — the caller's prefund balance held
// with the provider, per currency. Balance stays a string, matching how
// Topremit itself represents it.
type WalletBalanceEntry struct {
	Balance  string         `json:"balance"`
	Currency string         `json:"currency"`
	Extra    map[string]any `json:"extra,omitempty"`
}

type WalletBalanceResponse []WalletBalanceEntry

// StatusCallback is the body a backbone POSTs to core-api's
// POST /v1.0/backbone/callback-remittance-status, signed with this same
// contract's Sign function using the backbone as signer.
type StatusCallback struct {
	PartnerRef    string         `json:"partnerRef"`
	BackboneRef   string         `json:"backboneRef"`
	Status        OrderStatus    `json:"status"`
	FailureReason string         `json:"failureReason,omitempty"`
	Extra         map[string]any `json:"extra,omitempty"`
}
