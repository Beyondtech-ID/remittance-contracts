package remittance

import "fmt"

// APIError is returned by Client methods for any non-2xx response, decoded
// from the shared BaseResponse envelope. Code is empty if the envelope's
// ResponseCode couldn't be parsed (e.g. the backbone returned a non-JSON or
// non-conforming error body) — callers should still check HTTPStatus/
// Message in that case.
type APIError struct {
	HTTPStatus int
	Code       Code
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("remittance: %d %s: %s", e.HTTPStatus, e.Code, e.Message)
	}
	return fmt.Sprintf("remittance: %d: %s", e.HTTPStatus, e.Message)
}
