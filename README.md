# Vault Store

[![Tests Status](https://github.com/dracory/vaultstore/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/vaultstore/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/vaultstore)](https://goreportcard.com/report/github.com/dracory/vaultstore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/vaultstore)](https://pkg.go.dev/github.com/dracory/vaultstore)

Vault - a secure value storage (data-at-rest) implementation for Go.

## Scope

VaultStore is specifically designed as a data store component for securely storing and retrieving secrets. It is **not** an API or a complete secrets management system. Features such as user management, access control, and API endpoints are intentionally beyond the scope of this project.

VaultStore is meant to be integrated into your application as a library, providing the data storage layer for your secrets management needs. The application using VaultStore is responsible for implementing any additional layers such as API endpoints, user management, or access control if needed.

## Documentation

- [Overview](/docs/overview.md) - General overview of the VaultStore library
- [Usage Guide](/docs/usage_guide.md) - Examples of how to use VaultStore
- [Technical Reference](/docs/technical_reference.md) - Detailed technical information
- [Query Interface](/docs/query_interface.md) - Documentation for the flexible query interface
- [Data Stores](/docs/data_stores.md) - Information about the data store implementation

## Features

- Secure storage of sensitive data
- Token-based access to secrets
- Password protection for stored values
- Password rotation
- Flexible query interface for retrieving records
- Soft delete functionality for data recovery
- Support for multiple database backends

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at [https://www.gnu.org/licenses/agpl-3.0.en.html](https://www.gnu.org/licenses/agpl-3.0.txt)

For commercial use, please use my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Installation
```
go get -u github.com/dracory/vaultstore
```

## Technical Details

For database schema, record structure, and other technical information, please see the [Technical Reference](/docs/technical_reference.md).

## Setup

```golang
vault, err := NewStore(NewStoreOptions{
	VaultTableName:     "my_vault",
	DB:                 databaseInstance,
	AutomigrateEnabled: true,
})

```

## Usage

Here are some basic examples of using VaultStore. For comprehensive documentation, see the [Usage Guide](/docs/usage_guide.md).

```golang
// Create a token
token, err := vault.TokenCreate("my_value", "my_password", 20)
// token: "tk_abc123def456..."

// Check if a token exists
exists, err := vault.TokenExists(token)
// exists: true

// Read a value using a token
value, err := vault.TokenRead(token, "my_password")
// value: "my_value"

// Update a token's value
err := vault.TokenUpdate(token, "new_value", "my_password")

// Read multiple tokens at once (more efficient than individual calls)
ctx := context.Background()
tokens := []string{"token1", "token2", "token3"}
tokenValues, err := vault.TokensRead(ctx, tokens, "my_password")
// tokenValues: map[string]string{"token1": "value1", "token2": "value2", "token3": "value3"}

// Resolve multiple tokens with keys (convenience method)
keyTokenMap := map[string]string{
    "api_key":    "token1_here",
    "db_config":  "token2_here", 
    "auth_token": "token3_here",
}
resolvedMap, err := vault.TokensReadToResolvedMap(ctx, keyTokenMap, "my_password")
// resolvedMap: map[string]string{"api_key": "api_value", "db_config": "db_string", "auth_token": "auth_secret"}

// Bulk rekey all records with old password to new password
count, err := vault.BulkRekey(ctx, "old_password", "new_password")
// count: 5 (number of records rekeyed)

// Hard delete a token
err := vault.TokenDelete(token)

// Soft delete a token
err := vault.TokenSoftDelete(token)
```

## Changelog

For a detailed version history and changes, please see the [Changelog](/docs/changelog.md).
