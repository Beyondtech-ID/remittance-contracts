// Package remittance is the shared wire contract between core-api and any
// remittance backbone service it talks to (topremit-backbone today, more
// backbones/providers later).
//
// # Why this exists
//
// Before this module, core-api hand-mirrored a backbone's request/response
// structs by copy-pasting them into its own package
// (internal/remittance_backbone/dto.go), field by field, with a comment
// telling the next person to keep both copies in sync by hand. That drifted:
// core-api's copy of SubmitRequest and the backbone's real
// SubmitRemittanceRequest ended up with completely different fields, and
// core-api kept calling a path (/v1.0/remittance/submit) the backbone had
// already renamed to /v1.0/remittance/transactions. Both sides compiled
// fine; only a live call would have failed.
//
// This module is the fix: one Go module both core-api and every backbone
// import, so the request/response shapes, the HMAC signing scheme, and the
// path strings can never silently diverge between the two sides of a
// service-to-service call.
//
// # Adopting this in core-api
//
// Replace internal/remittance_backbone's hand-written Config/Client/DTOs
// with this package's Config/Client and the DTOs defined here. Core-api
// stops owning any backbone-shaped struct at all — it only imports types
// from this module.
//
// # Adopting this in a new backbone
//
// A new backbone implementing this contract (whether it's a second
// integration inside topremit-backbone's own partner_client abstraction, or
// an entirely separate service talking to a different provider) binds its
// HTTP handlers to the same request/response structs and the same Sign/
// Verify functions defined here, so any core-api client built against this
// module works against it unchanged — core only needs a new Config{BaseURL,
// ServiceID, secrets} pointed at the new backbone, no new mapping code.
//
// # Versioning policy — how "add a provider without breaking anything" works
//
// This module is additive-only within v1:
//   - Adding a new optional field (with omitempty) to an existing struct is
//     allowed and is NOT a breaking change — both the Go struct and the JSON
//     wire format tolerate unknown/missing fields.
//   - Adding a new DTO, a new Client method, or a new Code/status constant is
//     allowed.
//   - Removing a field, renaming a field's JSON tag, changing a field's type,
//     or changing what a Code/status constant means is NEVER allowed in v1.
//     That requires a new major version — a new module path
//     (.../remittance-contracts/v2) per standard Go module semantics — so
//     existing importers keep compiling against v1 until they choose to
//     migrate.
//
// # The Extra escape hatch
//
// Several request/response structs carry an `Extra map[string]any` field.
// It exists for provider-specific data that doesn't fit the fields already
// modeled here (today those fields are shaped by Topremit's API, since
// Topremit is the only integrated provider) — a new provider's peculiar
// field goes into Extra instead of requiring a change to the shared struct.
// If a field in Extra turns out to be broadly useful across providers, a
// later v1 release can promote it to a named optional field (additive,
// non-breaking); Extra itself never goes away.
//
// # What this module deliberately does not cover
//
// Data-reference/browsing endpoints (countries, payers, purposes,
// relationships, fund-sources, list-senders/recipients/transactions) are out
// of scope for v1: today they pass partner-shaped data through almost
// unchanged, change rarely, and carry much lower drift risk than the
// money-movement path. They can be added as new, purely additive DTOs and
// Client methods in a later v1 release without touching anything here.
package remittance
