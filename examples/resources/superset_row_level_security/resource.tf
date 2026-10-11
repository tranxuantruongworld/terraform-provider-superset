resource "superset_row_level_security" "vn_filter" {
  name        = "VN_Sales_Only"
  description = "Filter sales data for Vietnam analysts"
  filter_type = "Regular"
  clause      = "country = 'VN'"
  tables      = [1]
  roles       = [superset_role.analyst.id]
}
