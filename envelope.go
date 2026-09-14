package remittance

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// BaseResponse is the wire envelope every endpoint under this contract
// returns, byte-for-byte compatible with topremit-backbone's
// internal/shared/dto.BaseResponse. Any backbone implementing this contract
// MUST return exactly this shape — core-api's Client decodes responses
// assuming it.
type BaseResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	ResponseData    any    `json:"responseData"`
}

// NewBaseResponse builds the envelope using the same
// {httpCode}{serviceCode}{caseCode} concatenation topremit-backbone's
// internal/shared/delivery.ResponseWithCode already uses (e.g. httpCode=200,
// serviceCode="RM", caseCode="00" -> "200RM00"), so any backbone can produce
// a byte-identical envelope without re-deriving the format itself.
func NewBaseResponse(data any, message string, httpCode int, serviceCode, caseCode string) BaseResponse {
	return BaseResponse{
		ResponseCode:    fmt.Sprintf("%d%s%s", httpCode, serviceCode, caseCode),
		ResponseMessage: message,
		ResponseData:    data,
	}
}

// ParseResponseCode splits a ResponseCode back into serviceCode/caseCode.
// ResponseCode has no separators (it's httpCode+serviceCode+caseCode glued
// together), so the real HTTP status the response arrived with — known
// independently from the transport layer, never parsed out of the string —
// is required to know how many leading digits belong to httpCode. What's
// left is split at the first digit: the leading run of non-digit
// characters is serviceCode, the rest is caseCode. This assumes serviceCode
// is always alphabetic ("RM", "AU", "UM", "XX", ...), true for every case
// topremit-backbone's ExtractSnapError produces.
func ParseResponseCode(httpStatus int, responseCode string) (serviceCode, caseCode string, ok bool) {
	prefix := strconv.Itoa(httpStatus)
	if !strings.HasPrefix(responseCode, prefix) {
		return "", "", false
	}
	rest := responseCode[len(prefix):]
	i := 0
	for i < len(rest) && !unicode.IsDigit(rune(rest[i])) {
		i++
	}
	if i == 0 || i == len(rest) {
		return "", "", false
	}
	return rest[:i], rest[i:], true
}
