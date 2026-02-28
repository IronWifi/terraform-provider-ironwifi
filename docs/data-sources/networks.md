# ironwifi_networks

Lists IronWiFi networks with optional name filtering.

## Example Usage

```hcl
data "ironwifi_networks" "all" {}

data "ironwifi_networks" "office" {
  name_filter = "Office"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter networks by name (substring match).

## Attribute Reference

- `networks` - List of networks. Each network contains:
  - `id` - Network ID (UUID).
  - `nasname` - Network name.
  - `region` - RADIUS region.
  - `auth_port` - RADIUS authentication port.
  - `acct_port` - RADIUS accounting port.
  - `primary_ip` - Primary RADIUS server IP.
  - `secret` - RADIUS shared secret.
