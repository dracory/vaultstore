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
  - A *custom* derivation function `strongifyPassword` transforms the password using MD5, SHA-1, SHA-256 combinations to produce a long string.

- **Encryption construction (custom)**
  - `encode` pipeline:
    - Base64-encode the plaintext value.
    - Prefix the encoded value with its length and an underscore (`<len>_<b64>`).
    - Pad with pseudo-random characters to a fixed block length (>=128, doubling as needed).
    - Base64-encode the padded string.
    - XOR the result with the `strongifyPassword` output and Base64-encode that.
  - `decode` performs the inverse operations and validates structure (Base64 check, length parsing).

- **Randomness**
  - Token generation uses `crypto/rand` via `randomFromGamma` (good source of entropy) and a restricted gamma of `[a-z0-9]`.
  - Padding within `encode` uses `math/rand/v2` (non-cryptographic PRNG) to generate filler characters; this is not relied on for confidentiality, only obfuscation.

### Security Assessment of the Crypto Model

- **Custom crypto construction**
  - The scheme is *not* a standard, vetted construction (e.g., AES-GCM, XChaCha20-Poly1305, libsodium secretbox).
  - XOR-based schemes are typically fragile and prone to misuse; security relies entirely on correct, non-reused key usage and input handling.

- **Password derivation**
  - `strongifyPassword` combines MD5, SHA-1, and SHA-256, but:
    - It is not a standard password hashing / KDF function (e.g., Argon2, scrypt, PBKDF2).
    - It does not use salt or work factor / iteration count tuned to slow down brute-force attacks.
  - Consequence: Offline brute-force of passwords is easier than with dedicated password KDFs.

- **Integrity / authenticity**
  - There is no explicit MAC / AEAD tag.
  - Structural checks (Base64 validation, length parsing) provide **weak** integrity guarantees and may detect some corruptions or wrong passwords but are *not* cryptographic authenticity.

- **Error messages**
  - Decoding errors sometimes leak implementation details via error prefixes (e.g., `"xor. ..."`, `"base64.1. ..."`, `"vault password incorrect"`).
  - This can slightly aid an attacker performing targeted cryptanalysis or password guessing.

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

- **Replace custom crypto with standard AEAD**
  - Migrate `encode`/`decode` to a well-reviewed construction, such as:
    - AES-256-GCM or ChaCha20-Poly1305 via Go’s `crypto` packages or
    - A high-level library (e.g., `golang.org/x/crypto` primitives) that provides AEAD with nonce handling.
  - Store nonce + ciphertext (and optional associated data) in `vault_value`.

- **Use a standard KDF**
  - Replace `strongifyPassword` with a standard KDF (e.g., Argon2id, scrypt, or PBKDF2 with many iterations and per-record salt).
  - Store salt (and possibly KDF parameters) per record alongside the ciphertext.

- **Add integrity protection**
  - Leverage AEAD’s built-in authentication tag rather than ad-hoc Base64 and length checks.

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

- Custom XOR-based encryption and non-standard password derivation may not meet compliance requirements (e.g., PCI-DSS, ISO 27001) without further justification and review.
- There is no built-in key rotation or re-encryption mechanism; rotation must be implemented by the application using `TokenRead` + `TokenUpdate` with new parameters.
- The library assumes a trustworthy database driver and the `github.com/dracory/database` / `github.com/dracory/sb` helpers are secure against injection and misconfiguration.

---

This document should evolve along with the library. When making security-relevant code changes (crypto, token format, schema, or error handling), update this file to reflect the current design and threat model.
