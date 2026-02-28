resource "ironwifi_voucher" "guest_pass" {
  template_name    = "Guest WiFi Pass"
  voucher_format   = "alphanumeric"
  voucher_length   = 8
  voucher_quantity = 100
  voucher_devices  = 3
  voucher_duration = "24h"
  group_id         = ironwifi_group.guests.id
}
