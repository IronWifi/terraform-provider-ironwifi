# ironwifi_devices

Lists IronWiFi devices with optional name filtering.

## Example Usage

```hcl
data "ironwifi_devices" "all" {}

data "ironwifi_devices" "printers" {
  name_filter = "printer"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter devices by name (substring match).

## Attribute Reference

- `devices` - List of devices. Each device contains:
  - `id` - Device ID (UUID).
  - `username` - Device identifier (MAC address).
  - `email` - Associated email address.
  - `firstname` - First name.
  - `lastname` - Last name.
  - `status` - Device status.
