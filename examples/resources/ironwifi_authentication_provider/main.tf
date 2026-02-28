resource "ironwifi_authentication_provider" "ldap" {
  name   = "Corporate LDAP"
  type   = "ldap"
  status = "enabled"
  configuration = jsonencode({
    basedn     = "dc=example,dc=com"
    bind       = "cn=admin,dc=example,dc=com"
    dbpassword = "ldap-password"
  })
}
