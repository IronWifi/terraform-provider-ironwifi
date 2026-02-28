# ironwifi_voucher

Manages an IronWiFi voucher template.

## Example Usage

```hcl
resource "ironwifi_voucher" "guest_pass" {
  template_name    = "Guest WiFi Pass"
  voucher_format   = "alphanumeric"
  voucher_length   = 8
  voucher_quantity = 100
  voucher_devices  = 3
  voucher_duration = "24h"
  group_id         = ironwifi_group.guests.id
}
```

## Argument Reference

- `template_name` - (Required) Voucher template name.
- `voucher_format` - (Optional) Voucher code format.
- `voucher_length` - (Optional) Length of voucher codes.
- `voucher_quantity` - (Optional) Number of vouchers to generate.
- `group_id` - (Optional) Group ID to assign voucher users to.
- `orgunit_id` - (Optional) Organizational unit ID.
- `voucher_deletedate` - (Optional) Voucher expiration/delete date.
- `voucher_devices` - (Optional) Number of devices allowed per voucher.
- `voucher_duration` - (Optional) Voucher validity duration.

## Attribute Reference

- `id` - Voucher ID (UUID).

## Import

```bash
terraform import ironwifi_voucher.example <voucher-id>
```
