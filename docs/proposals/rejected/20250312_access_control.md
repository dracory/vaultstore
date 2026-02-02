# Access Control and Permissions

## Status: Rejected (Beyond Scope, VaultStore is a storage component, not a full secrets-management system)

## Overview

This proposal suggests implementing a comprehensive access control and permissions system for VaultStore, allowing fine-grained control over who can access and modify specific secrets.

## Current Implementation

Currently, VaultStore has a simple security model where anyone with a token and the correct password can access a secret. There is no built-in way to restrict who can create, read, update, or delete specific secrets.

## Proposed Changes

1. **Role-Based Access Control (RBAC)**: Implement a role-based access control system with predefined roles (e.g., admin, writer, reader).

2. **User Management**: Add user management functionality to create, update, and delete users.

3. **Permission Granularity**: Support permissions at different levels:
   - Global permissions (e.g., can create new secrets)
   - Category permissions (e.g., can access all secrets in a category)
   - Individual secret permissions (e.g., can read but not update a specific secret)

4. **Secret Categorization**: Add the ability to organize secrets into categories or namespaces.

5. **Access Policies**: Implement policy definitions that can be attached to users, roles, or secrets.

## Implementation Details

The implementation would require:

- Creating new tables for users, roles, permissions, and policies
- Extending the context parameter to include authentication information
- Implementing permission checks in all operations
- Adding API methods for managing users, roles, and permissions

Example new structures and methods:

```go
type User struct {
    ID       string
    Username string
    // Other user attributes
}

type Role struct {
    ID   string
    Name string
    // Permissions associated with this role
}

type Permission struct {
    Action   string // create, read, update, delete
    Resource string // global, category:name, secret:id
}

// New method signatures
CreateUser(ctx context.Context, username string, initialPassword string) (User, error)
AssignRoleToUser(ctx context.Context, userID string, roleID string) error
GrantPermission(ctx context.Context, roleID string, permission Permission) error
CheckPermission(ctx context.Context, userID string, action string, resource string) (bool, error)
```

## Benefits

- **Security Enhancement**: Restricts access to secrets based on user identity and permissions
- **Compliance**: Supports the principle of least privilege required by many security standards
- **Multi-User Support**: Enables safe use in team environments
- **Delegation**: Allows administrators to delegate specific permissions to other users

## Risks and Mitigations

- **Complexity**: Significantly increases system complexity. Mitigation: Provide sensible defaults and helper functions.
- **Performance Impact**: Permission checks add overhead. Mitigation: Implement efficient caching of permissions.
- **Migration**: Existing deployments need migration. Mitigation: Provide tools to migrate and assign default permissions.

## Effort Estimation

- Development: 3-4 weeks
- Testing: 2 weeks
- Documentation: 1 week

## Conclusion

Implementing an access control and permissions system would transform VaultStore from a simple secret storage solution to an enterprise-grade secrets management platform suitable for team environments with complex security requirements.
