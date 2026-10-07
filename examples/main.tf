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
# FEATURE 4: TRUY VẤN PERMISSIONS CÓ SẴN (Permission View Menus - PVM)
# ==============================================================================
# Truy vấn quyền đọc Dashboard
data "superset_permission" "can_read_dashboard" {
  permission_name = "can_read"
  view_menu_name   = "Dashboard"
}

# Truy vấn quyền truy cập tất cả datasources (hoặc theo schema/database cụ thể)
data "superset_permission" "all_datasource_access" {
  permission_name = "all_datasource_access"
  view_menu_name   = "all_datasource_access"
}

# Truy vấn role Gamma mặc định của Superset
data "superset_role" "gamma" {
  name = "Gamma"
}

# ==============================================================================
# FEATURE 1: TẠO ROLE VÀ GÁN PERMISSIONS CHI TIẾT
# ==============================================================================
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"

  # Gán danh sách permission IDs trực tiếp vào Role
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
# FEATURE 2: ROW-LEVEL SECURITY (RLS) - PHÂN QUYỀN TRUY CẬP THEO DÒNG DỮ LIỆU
# ==============================================================================
# Luật RLS: Chỉ cho phép người dùng thuộc role Finance_Analyst xem các dòng của Việt Nam
resource "superset_row_level_security" "vn_finance_filter" {
  name        = "VN_Finance_Data_Only"
  description = "Chi xem du lieu doanh thu tai Viet Nam cho Finance Analyst"
  filter_type = "Regular" # Regular: ap dung cho roles duoc chi dinh; Base: ap dung cho tat ca tru roles duoc chi dinh
  clause      = "country = 'VN' AND department = 'finance'"

  # ID cua bang / dataset trong Superset
  tables = [1]

  # Gán luật RLS này vào role Finance_Analyst
  roles = [
    superset_role.finance_analyst.id,
  ]
}

# ==============================================================================
# QUẢN LÝ USER VỚI SSO / UPN (Azure AD, GitLab, Vault, Keycloak)
# ==============================================================================
# Người dùng chỉ cần khai báo UPN, hệ thống tự động:
# 1. Trích xuất username = UPN
# 2. Email = UPN
# 3. Tự tạo mật khẩu tạm thời ngẫu nhiên (vì xác thực ủy quyền qua SSO/OIDC)
# 4. Phân quyền sẵn các Role khi user đăng nhập SSO lần đầu tiên
resource "superset_user" "sso_analyst" {
  upn = "nguyen.van.a@company.com"

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

# User tạo truyền thống (Local Database User)
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
