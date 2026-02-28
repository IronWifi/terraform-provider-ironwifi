# ironwifi_groups

Lists IronWiFi groups with optional name filtering.

## Example Usage

```hcl
data "ironwifi_groups" "all" {}

data "ironwifi_groups" "eng" {
  name_filter = "Engineering"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter groups by name (substring match).

## Attribute Reference

- `groups` - List of groups. Each group contains:
  - `id` - Group ID (UUID).
  - `groupname` - Group name.
  - `description` - Group description.
  - `priority` - Group priority.
