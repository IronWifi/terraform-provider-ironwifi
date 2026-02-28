resource "ironwifi_user" "employee" {
  username   = "john.doe@example.com"
  password   = "SecureP@ssw0rd"
  email      = "john.doe@example.com"
  firstname  = "John"
  lastname   = "Doe"
  user_type  = "e"
  authsource = "local"
}
