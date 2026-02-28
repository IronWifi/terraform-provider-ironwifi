resource "ironwifi_certificate" "employee_cert" {
  user_id      = ironwifi_user.employee.id
  cn           = "john.doe"
  subject      = "CN=john.doe,O=Example Corp"
  validity     = 365
  distribution = "email"
  hash         = "sha2"
}
