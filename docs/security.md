# Security Overview

This document summarizes the security properties, assumptions, and risks of the `vaultstore` library, with concrete recommendations for safe use and future hardening.

## Scope and Threat Model

- **Library only**
  - `vaultstore` is a storage component, not a full secrets-management system.
  - It does not implement authentication, authorization, rate limiting, or network-facing APIs.
  - The calling application is responsible for all access control and exposure boundaries.

- **Primary goal**
  - Provide *data-at-rest protection* for secret values stored in a database table via a token + password model.

- **Out of scope**
  - Protection against:
    - A fully compromised application host/runtime.
    - A fully compromised database + application simultaneously.
    - Client-side security, transport security (TLS), or key distribution.

## High-Level Design

- Secrets are stored as records in a single table (configurable name).
- Each record includes:
  - `id` (primary key)
  - `vault_token` (unique, random-ish token, typically with `tk_` prefix)
  - `vault_value` (encrypted/encoded string)
  - timestamps (`created_at`, `updated_at`, `soft_deleted_at`)
- Secrets are written and accessed via the `StoreInterface` using:
  - `TokenCreate`, `TokenCreateCustom`, `TokenUpdate`
  - `TokenRead`, `TokensRead`, `TokenExists`, `TokenDelete`, `TokenSoftDelete`

## Encryption and Encoding Model

- **Password-derived key**
  - Calling code provides a `password` for `TokenCreate`/`TokenUpdate` and must use the same password for `TokenRead`/`TokensRead`.
  - Uses **Argon2id** key derivation function with configurable parameters via `CryptoConfig`.
  - Default parameters: 3 iterations, 64MB memory, parallelism of 4.

- **Encryption construction (AES-256-GCM)**
  - Uses standard AES-256-GCM authenticated encryption.
  - Each encryption generates a unique 12-byte nonce.
  - Ciphertext format: `base64(nonce || ciphertext || tag)`.
  - Built-in authentication tag provides integrity and authenticity guarantees.

- **Randomness**
  - Token generation uses `crypto/rand` for cryptographically secure randomness.
  - Salt generation uses `crypto/rand`.
  - All cryptographic operations use secure random sources.

### Security Assessment of the Crypto Model

- **Standard crypto construction**
  - Uses well-reviewed AES-256-GCM authenticated encryption.
  - Argon2id is the current recommended password hashing algorithm (OWASP, IETF).
  - Construction provides both confidentiality and authenticity.

- **Password derivation**
  - Argon2id with tunable parameters (iterations, memory, parallelism).
  - Per-record salt generated via `crypto/rand`.
  - Memory-hard function resists GPU/ASIC attacks.

- **Integrity / authenticity**
  - AES-GCM provides built-in authentication tags.
  - Tampering with ciphertext is detected during decryption.
  - No reliance on ad-hoc integrity checks.

## Data Store and SQL Layer

- Uses `github.com/doug-martin/goqu/v9` to build queries and `github.com/dracory/database` helpers for execution.
- Queries are constructed with bound parameters, which helps prevent classical SQL injection, assuming upstream `database` helpers are safe.
- Table creation uses a builder (`github.com/dracory/sb`) and creates:
  - Unique constraint on `vault_token`.
  - `vault_value` as `LONGTEXT` (large payloads allowed).

### Risks and Considerations

- **Soft delete semantics**
  - Soft-deleted records remain in the database with `soft_deleted_at` set.
  - Callers must ensure that these records are not exposed via custom queries.

- **Token enumeration**
  - Tokens are random and prefixed with `tk_`, but:
    - The token length is configurable; too-short tokens reduce the search space.
    - If tokens are ever exposed in predictable patterns or logs, enumeration / guessing risk increases.

- **Logging and debug mode**
  - When `debugEnabled` is true, SQL strings and some errors are logged via `log.Println`.
  - Depending on calling code and log configuration, this may leak table and schema information, and, via other layers, potentially sensitive context.

## Application Responsibilities

To use `vaultstore` safely, the integrating application must handle:

- **Authentication & authorization**
  - Restrict which principals can call token operations.
  - Enforce RBAC or other policies for creating, reading, updating, and deleting secrets.

- **Rate limiting & abuse protection**
  - Implement rate limiting / throttling on operations, especially those involving password verification (`TokenRead`, `TokensRead`).
  - Consider IP-based and principal-based rate limits to reduce brute-force attempts.

- **Transport security**
  - Ensure all traffic to any API wrapping `vaultstore` uses TLS and secure client configuration.

- **Secrets and password management**
  - Decide how user passwords or application-level passwords are provisioned, rotated, and stored.
  - Avoid logging passwords, tokens, or decrypted secret values.

## Recommendations for Hardening

### Short-Term (Library Configuration / Usage)

- **Minimum token length**
  - Enforce sufficiently large token lengths at call sites (e.g., 32+ characters total) to reduce guessing risk.

- **Disable debug in production**
  - Ensure `EnableDebug(false)` in all production environments.

- **Reduce error leakage**
  - Consider mapping detailed internal errors to a generic error (e.g., `"invalid token or password"`) at the application boundary.

- **Strong passwords / application keys**
  - Require high-entropy passwords or use randomly generated application keys instead of user-chosen passwords when possible.

### Medium-Term (Library Changes)

- **CryptoConfig tuning**
  - Adjust Argon2id parameters based on your security requirements and performance constraints.
  - Use `HighSecurityCryptoConfig()` for maximum protection or `LightweightCryptoConfig()` for resource-constrained environments.

- **Add integrity protection**
  - AES-GCM already provides built-in authentication; ensure it's used consistently.

- **Clarify threat model in docs**
  - Document explicitly that:
    - `vaultstore` protects against database compromise without password knowledge.
    - It does **not** protect against compromise of the application runtime or leakage of passwords / keys.

### Long-Term (Ecosystem and Operational)

- **Security review and tests**
  - Commission formal cryptographic review once migrated to standard primitives.
  - Add property-based and fuzz tests around encode/decode and token operations.

- **Monitoring and observability**
  - Provide hooks or guidance for emitting structured security logs (e.g., token operations, failed reads) without leaking sensitive data.

## Known Limitations

- AES-256-GCM encryption with Argon2id KDF provides strong security but may require tuning for specific compliance requirements (e.g., PCI-DSS, ISO 27001).
- Built-in key rotation is now available via `BulkRekey()` when `PasswordIdentityEnabled` is true; otherwise rotation requires `TokenRead` + `TokenUpdate`.
- The library assumes a trustworthy database driver and the `github.com/dracory/database` / `github.com/dracory/sb` helpers are secure against injection and misconfiguration.

---

This document should evolve along with the library. When making security-relevant code changes (crypto, token format, schema, or error handling), update this file to reflect the current design and threat model.
