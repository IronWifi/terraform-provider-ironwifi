resource "ironwifi_org_unit" "engineering" {
  name        = "Engineering"
  description = "Engineering department"
}

resource "ironwifi_org_unit" "frontend" {
  name        = "Frontend Team"
  description = "Frontend engineering team"
  parent_id   = ironwifi_org_unit.engineering.id
}
