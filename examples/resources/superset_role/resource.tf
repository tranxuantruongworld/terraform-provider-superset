resource "superset_role" "analyst" {
  name = "Data_Analyst"
  permissions = [
    data.superset_permission.can_read_dashboard.id,
  ]
}

