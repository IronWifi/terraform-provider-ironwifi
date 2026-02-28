resource "ironwifi_connector" "corp_ldap" {
  name       = "Corporate LDAP"
  type       = "ldap"
  domain     = "example.com"
  basedn     = "dc=example,dc=com"
  bind       = "cn=admin,dc=example,dc=com"
  password   = var.ldap_password
  authsource = "ldap"
  status     = "enabled"
}
