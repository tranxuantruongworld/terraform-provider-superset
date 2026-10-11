# Pre-provision SSO User via UPN
resource "superset_user" "sso_user" {
  upn   = "user@company.com"
  roles = [superset_role.analyst.id]
}

# Local Database User
resource "superset_user" "local_admin" {
  username   = "local_admin"
  first_name = "Local"
  last_name  = "Admin"
  email      = "local_admin@company.com"
  password   = "SuperSecretPassword123!"
  active     = true
  roles      = [superset_role.analyst.id]
}
