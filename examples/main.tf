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
  password = "admin_password"
  insecure = true
}

# ==============================================================================
# FEATURE: QUERY EXISTING PERMISSIONS (Permission View Menus - PVM)
# ==============================================================================
# Query permission to read Dashboards
data "superset_permission" "can_read_dashboard" {
  permission_name = "can_read"
  view_menu_name  = "Dashboard"
}

# Query permission to access all datasources (or specify schema/database)
data "superset_permission" "all_datasource_access" {
  permission_name = "all_datasource_access"
  view_menu_name  = "all_datasource_access"
}

# Query default Gamma role in Superset
data "superset_role" "gamma" {
  name = "Gamma"
}

# ==============================================================================
# FEATURE: CREATE ROLES AND ASSIGN GRANULAR PERMISSIONS
# ==============================================================================
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"

  # Assign list of permission IDs directly to Role
  permissions = [
    data.superset_permission.can_read_dashboard.id,
    data.superset_permission.all_datasource_access.id,
  ]
}

resource "superset_role" "data_engineer" {
  name = "Data_Engineer"
  permissions = [
    data.superset_permission.can_read_dashboard.id,
  ]
}

# ==============================================================================
# FEATURE: ROW-LEVEL SECURITY (RLS) - DATA ACCESS CONTROL BY ROW
# ==============================================================================
# RLS rule: Only allow Finance_Analyst role to view rows for Vietnam
resource "superset_row_level_security" "vn_finance_filter" {
  name        = "VN_Finance_Data_Only"
  description = "Filter revenue data for Vietnam finance analysts"
  filter_type = "Regular" # Regular: applies to specified roles; Base: applies to all except specified roles
  clause      = "country = 'VN' AND department = 'finance'"

  # Dataset / Table ID in Superset
  tables = [1]

  # Associate this RLS rule with Finance_Analyst role
  roles = [
    superset_role.finance_analyst.id,
  ]
}

# ==============================================================================
# USER MANAGEMENT WITH SSO / UPN (Azure AD, GitLab, Vault, Keycloak)
# ==============================================================================
# Identity Mapping (Account Pre-linking) using UPN:
# 1. username = UPN (matches claim sent by IdP upon user login)
# 2. email = UPN
# 3. Password: User authenticates 100% via SSO IdP Token, bypassing DB local passwords.
#    (Provider handles schema requirements internally for Superset REST API).
# 4. Pre-assign Roles & RLS so permissions apply immediately upon first SSO login.
resource "superset_user" "sso_analyst" {
  upn = "analyst.user@company.com"

  roles = [
    data.superset_role.gamma.id,
    superset_role.finance_analyst.id,
  ]
}

resource "superset_user" "sso_engineer" {
  upn = "alex.data@company.com"

  roles = [
    superset_role.data_engineer.id,
  ]
}

# Traditional Local Database User
resource "superset_user" "local_admin_backup" {
  username   = "backup_admin"
  first_name = "Backup"
  last_name  = "Admin"
  email      = "backup@company.com"
  password   = "VeryStrongPassword123!"
  active     = true

  roles = [
    data.superset_role.gamma.id,
  ]
}

# Outputs
output "sso_user_id" {
  value = superset_user.sso_analyst.id
}

output "rls_rule_id" {
  value = superset_row_level_security.vn_finance_filter.id
}
