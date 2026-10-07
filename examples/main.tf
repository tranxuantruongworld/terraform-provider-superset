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

# 1. Truy vấn role có sẵn trong Superset (ví dụ role Gamma cho người xem báo cáo)
data "superset_role" "gamma" {
  name = "Gamma"
}

# 2. Định nghĩa các Role mới (nhóm quyền) theo chuẩn RBAC
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"
}

resource "superset_role" "data_engineer" {
  name = "Data_Engineer"
}

# 3. Tạo User và phân quyền (gán Roles) trực tiếp qua Terraform
resource "superset_user" "analyst_user" {
  username   = "john.analyst"
  first_name = "John"
  last_name  = "Doe"
  email      = "john.analyst@company.com"
  password   = "SuperSecret123!"
  active     = true

  roles = [
    data.superset_role.gamma.id,
    superset_role.finance_analyst.id,
  ]
}

resource "superset_user" "engineer_user" {
  username   = "alice.engineer"
  first_name = "Alice"
  last_name  = "Smith"
  email      = "alice.engineer@company.com"
  password   = "SuperSecret456!"
  active     = true

  roles = [
    superset_role.data_engineer.id,
  ]
}

# Outputs
output "analyst_user_id" {
  value = superset_user.analyst_user.id
}

output "engineer_user_id" {
  value = superset_user.engineer_user.id
}
