data "superset_permission" "can_read_dashboard" {
  permission_name = "can_read"
  view_menu_name  = "Dashboard"
}
