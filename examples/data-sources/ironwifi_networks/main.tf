data "ironwifi_networks" "all" {}

data "ironwifi_networks" "office" {
  name_filter = "Office"
}

output "all_networks" {
  value = data.ironwifi_networks.all.items
}
