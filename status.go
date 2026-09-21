package remittance

// OrderStatus is the coarse, provider-agnostic status of a remittance
// order as seen from core-api's side of the contract. A backbone maps its
// upstream provider's own (often much more granular) status vocabulary
// down to one of these before ever returning it across this boundary —
// core-api never needs to know a provider's raw status strings.
type OrderStatus string

const (
	// OrderStatusProcessing: order accepted, execution against the
	// provider not yet confirmed complete.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusExecuting: a worker has claimed this order and is calling
	// the provider right now. Exposed for observability; core-api should
	// treat it the same as OrderStatusProcessing (still non-terminal).
	OrderStatusExecuting OrderStatus = "EXECUTING"
	// OrderStatusSuccess: terminal, money delivered.
	OrderStatusSuccess OrderStatus = "SUCCESS"
	// OrderStatusFailed: terminal, generic failure (rejected, refunded —
	// anything that isn't specifically OrderStatusCancelled or
	// OrderStatusExpired below).
	OrderStatusFailed OrderStatus = "FAILED"
	// OrderStatusCancelled: terminal — either core's own checker rejected
	// it before submit, or the provider confirmed a cancellation.
	// Core-api is expected to tell the two apart via source/reason on its
	// own side; this value only means "did not go through, cancelled".
	OrderStatusCancelled OrderStatus = "CANCELLED"
	// OrderStatusExpired: terminal — core's approval window lapsed, the
	// quote went invalid before the first submit attempt, or the provider
	// itself reports an expiry per its own contract terms.
	OrderStatusExpired OrderStatus = "EXPIRED"
	// OrderStatusNeedsManualReview: non-terminal, but NOT auto-retried by
	// the backbone — the outcome at the provider is ambiguous (e.g. a
	// submit call failed after the provider may have already accepted it)
	// and requires a human to reconcile before anything else happens to
	// this order.
	OrderStatusNeedsManualReview OrderStatus = "NEEDS_MANUAL_REVIEW"
)

// IsTerminal reports whether this status is a final outcome — no further
// status transition or callback should be expected after one of these.
func (s OrderStatus) IsTerminal() bool {
	switch s {
	case OrderStatusSuccess, OrderStatusFailed, OrderStatusCancelled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// CallbackStatus tracks whether a backbone has successfully delivered a
// StatusCallback to core-api yet.
type CallbackStatus string

const (
	CallbackStatusPending   CallbackStatus = "PENDING"
	CallbackStatusDelivered CallbackStatus = "DELIVERED"
)
