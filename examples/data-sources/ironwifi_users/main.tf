data "ironwifi_users" "admins" {
  username_filter = "admin"
}

output "admin_users" {
  value = data.ironwifi_users.admins.items
}
