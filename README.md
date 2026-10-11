# Terraform Provider for Apache Superset (`terraform-provider-superset`)

[![Go Version](https://img.shields.io/badge/go-1.27.1-blue.svg)](https://go.dev/)
[![Terraform Version](https://img.shields.io/badge/terraform-1.16.5-purple.svg)](https://www.terraform.io/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

The open-source **Terraform Provider for Apache Superset** is built with **Go** and HashiCorp's **Terraform Plugin Framework**. It enables managing Apache Superset Role-Based Access Control (RBAC), Row-Level Security (RLS), and users as Infrastructure as Code (IaC).

---

## 🚀 Key Features

- **Role & Permission Management (`superset_role`)**: Create and configure roles with granular permissions via Permission View Menus (PVM).
- **Row-Level Security (`superset_row_level_security`)**: Manage row-level filters using flexible SQL clauses (`clause`), associated datasets (`tables`), and assigned roles (`roles`).
- **Permission Data Source (`data.superset_permission`)**: Look up Permission View Menu IDs by `(permission_name, view_menu_name)` pairs such as `can_read` on `Dashboard`, `all_datasource_access`, etc.
- **User Management & SSO/UPN Support (`superset_user`)**: Support account pre-provisioning using **UPN (User Principal Name)** for Single Sign-On users (Azure AD, GitLab, Vault, Keycloak). Simply declare `upn` and `roles`, and the provider takes care of user setup.
- **Automated REST API & CSRF Authentication**: JWT authentication via `/api/v1/security/login`, with automatic retrieval and injection of `X-CSRFToken` headers for state-changing operations (POST/PUT/DELETE).

---

## 📁 Project Structure

```text
terraform-provider-superset/
├── internal/
│   ├── client/                  # Superset REST API Client (Auth, Users, Roles, PVM, RLS, CSRF)
│   │   ├── client.go
│   │   ├── client_test.go       # Mock HTTP server unit tests
│   │   ├── models.go            # JSON data models
│   │   ├── permissions.go       # Permission View Menus API queries
│   │   ├── rls.go               # Row-Level Security CRUD API
│   │   ├── roles.go             # Roles CRUD API
│   │   └── users.go             # Users CRUD API
│   └── provider/                # Terraform Plugin Framework implementation
│       ├── provider.go          # Provider configuration
│       ├── provider_test.go     # Provider schema tests
│       ├── resource_role.go     # superset_role resource (with permissions)
│       ├── resource_row_level_security.go # superset_row_level_security resource
│       ├── resource_user.go     # superset_user resource (with UPN SSO support)
│       ├── data_source_permission.go # superset_permission data source
│       ├── data_source_role.go  # superset_role data source
│       └── data_source_user.go  # superset_user data source
├── examples/
│   ├── dev.tfrc                 # dev_overrides configuration for local Terraform CLI testing
│   └── main.tf                  # Complete end-to-end example for RBAC, RLS, and UPN SSO
├── Makefile                     # Build, test, format, install targets
├── main.go                      # Entry point for the Terraform Provider plugin
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE                      # Apache 2.0 License
└── README.md
```

---

## 🛠 Prerequisites

- **Go**: `1.27.1` (verify with `go version`)
- **Terraform**: `1.16.5` (verify with `terraform version`)

Both tools can be installed or placed in `~/.local/bin` (included in your `$PATH`).

---

## 🔨 Build & Test Guide

### 1. Build the Provider
```bash
cd /home/xtruong/Desktop/open-source/terraform-provider-superset
make build
# or: go build -o terraform-provider-superset .
```

### 2. Run Unit Tests
```bash
make test
# or: go test -v ./...
```

---

## 🔐 SSO Deep Dive: How Superset Resolves UPN for Access Control

When users or services log into Superset via Identity Providers (IdP) like **Azure AD / Microsoft Entra ID**, **GitLab**, **HashiCorp Vault**, or **Keycloak**:

1. **How It Works**:
   - During SSO login via OAuth2/OIDC, the IdP returns a payload containing identity claims:
     - **Azure AD (Entra ID)**: provides `userPrincipalName` or `upn` (e.g., `user@company.com`).
     - **GitLab**: provides `preferred_username` or `email`.
     - **Vault / Keycloak**: provides `upn`, `preferred_username`, or `email`.
   - In `superset_config.py`, Flask-AppBuilder's (FAB) `CustomSecurityManager` maps this claim to the Superset `username` field.

2. **Integration with Terraform**:
   - DevOps and Platform Engineers pre-provision users in Terraform specifying only `upn`:
     ```hcl
     resource "superset_user" "analyst" {
       upn   = "analyst.user@company.com"
       roles = [superset_role.finance_analyst.id]
     }
     ```
   - The Terraform Provider automatically:
     - Sets `username = upn` (`analyst.user@company.com`).
     - Sets `email = upn`.
     - Generates a cryptographically secure random temporary password (since authentication is handled by the SSO token rather than local DB passwords).
     - Pre-assigns the designated roles and RLS policies to the user.
   - When the user logs in for the first time via Azure AD, GitLab, or Vault, Superset matches the user account where `username == upn` and applies all roles and permissions managed by Terraform without manual onboarding.

3. **Sample Configuration in `superset_config.py`**:
```python
from superset.security import SupersetSecurityManager

class CustomSsoSecurityManager(SupersetSecurityManager):
    def oauth_user_info(self, provider, response=None):
        if provider == 'azure':
            data = self.appbuilder.sm.oauth_remotes[provider].get('me').json()
            # UPN is the primary identity attribute in Azure AD
            upn = data.get('userPrincipalName') or data.get('upn') or data.get('mail')
            return {
                'username': upn,  # Matches the 'upn' defined in Terraform
                'email': data.get('mail') or upn,
                'first_name': data.get('givenName') or upn.split('@')[0],
                'last_name': data.get('surname') or 'SSO',
            }
        elif provider in ('gitlab', 'vault', 'keycloak'):
            data = self.appbuilder.sm.oauth_remotes[provider].get('user').json()
            upn = data.get('preferred_username') or data.get('email')
            return {
                'username': upn,
                'email': data.get('email') or upn,
                'first_name': data.get('name', 'SSO User'),
                'last_name': 'SSO',
            }
        return {}

CUSTOM_SECURITY_MANAGER = CustomSsoSecurityManager
```

---

## 💡 Complete HCL Configuration Example (`examples/main.tf`)

```hcl
terraform {
  required_providers {
    superset = {
      source = "xtruong/superset"
    }
  }
}

provider "superset" {
  base_url = "http://localhost:8088"
  username = "admin"
  password = "your_admin_password"
  insecure = true
}

# 1. Query specific permissions
data "superset_permission" "read_dashboard" {
  permission_name = "can_read"
  view_menu_name  = "Dashboard"
}

data "superset_permission" "all_datasource" {
  permission_name = "all_datasource_access"
  view_menu_name  = "all_datasource_access"
}

# 2. Create role and assign granular permissions
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"
  permissions = [
    data.superset_permission.read_dashboard.id,
    data.superset_permission.all_datasource.id,
  ]
}

# 3. Define Row-Level Security (RLS)
resource "superset_row_level_security" "vn_finance_data" {
  name        = "VN_Finance_Only"
  filter_type = "Regular"
  clause      = "country = 'VN' AND department = 'finance'"
  tables      = [1] # Dataset ID
  roles       = [superset_role.finance_analyst.id]
}

# 4. Pre-provision SSO User using UPN
resource "superset_user" "sso_analyst" {
  upn   = "analyst.user@company.com"
  roles = [superset_role.finance_analyst.id]
}
```

---

## 📄 License
This project is licensed under the [Apache License 2.0](LICENSE).
