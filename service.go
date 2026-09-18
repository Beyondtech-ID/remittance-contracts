package remittance

// Service* are Topremit's payment-method codes — the value that goes in
// QuoteRequest.Service, CreateRecipientRequest.ServiceID, and
// GetRates'/Client.GetRates' serviceId. Plain string constants (not a
// distinct named type) so they drop into those existing string fields
// without a conversion. Not every corridor supports every service — some
// corridors only exist under one specific service (e.g. China only quotes
// under ServiceWeChatPay/ServiceAlipay, never ServiceBankAccount; confirmed
// live 2026-09-18, "This country is currently unavailable" was really "not
// under this service").
const (
	ServiceBankAccount = "BankAccount" // Bank Transfer
	ServiceBiFAST      = "BiFAST"
	ServiceLLG         = "LLG"
	ServiceRTGS        = "RTGS"
	ServiceRTOL        = "RTOL"
	ServiceUnionPay    = "UnionPay"
	ServiceEWallet     = "eWallet"
	ServiceCashPickup  = "CashPickup"
	ServiceAlipay      = "Alipay"
	ServiceWeChatPay   = "WeChatPay"
)
