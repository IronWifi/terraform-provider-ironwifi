resource "ironwifi_network" "office" {
  name   = "Office-WiFi"
  region = "us-east1"
  ipv6   = false
  coa    = true
}

output "network_secret" {
  value     = ironwifi_network.office.secret
  sensitive = true
}

output "primary_ip" {
  value = ironwifi_network.office.primary_ip
}
