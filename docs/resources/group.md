# ironwifi_group

Manages an IronWiFi group.

## Example Usage

```hcl
resource "ironwifi_group" "engineering" {
  name        = "Engineering"
  description = "Engineering team"
  priority    = 10
}
```

## Argument Reference

- `name` - (Required) Group name.
- `description` - (Optional) Group description.
- `priority` - (Optional) Group priority. Defaults to `0`.

## Attribute Reference

- `id` - Group ID (UUID).

## Import

```bash
terraform import ironwifi_group.example <group-id>
```
