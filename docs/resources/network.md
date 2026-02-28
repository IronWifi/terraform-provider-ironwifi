# ironwifi_network

Manages an IronWiFi network (RADIUS NAS).

## Example Usage

```hcl
resource "ironwifi_network" "office" {
  name   = "Office-WiFi"
  region = "us-east1"
  ipv6   = false
  coa    = true
}
```

## Argument Reference

- `name` - (Required) Network name (NAS identifier).
- `region` - (Optional) RADIUS region. Defaults to `""`.
- `auth_port` - (Optional) RADIUS authentication port. Defaults to `1812`.
- `acct_port` - (Optional) RADIUS accounting port. Defaults to `1813`.
- `ipv6` - (Optional) Enable IPv6 support. Defaults to `false`.
- `unknown_users` - (Optional) Action for unknown users: `reject` or `accept`. Defaults to `reject`.
- `open_roaming` - (Optional) Enable OpenRoaming. Defaults to `false`.
- `eduroam` - (Optional) Enable eduroam federation. Defaults to `false`.
- `coa` - (Optional) Enable Change of Authorization. Defaults to `false`.
- `radsec` - (Optional) Enable RadSec. Defaults to `false`.

## Attribute Reference

- `id` - Network ID (UUID).
- `secret` - RADIUS shared secret (sensitive).
- `primary_ip` - Primary RADIUS server IP.
- `backup_ip` - Backup RADIUS server IP.

## Import

```bash
terraform import ironwifi_network.example <network-id>
```
