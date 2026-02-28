# ironwifi_org_unit

Manages an IronWiFi organizational unit.

## Example Usage

```hcl
resource "ironwifi_org_unit" "engineering" {
  name        = "Engineering"
  description = "Engineering department"
}

resource "ironwifi_org_unit" "frontend" {
  name        = "Frontend Team"
  description = "Frontend engineering team"
  parent_id   = ironwifi_org_unit.engineering.id
}
```

## Argument Reference

- `name` - (Required) Org unit name.
- `description` - (Optional) Org unit description.
- `parent_id` - (Optional) Parent org unit ID for nesting.

## Attribute Reference

- `id` - Org unit ID (UUID).

## Import

```bash
terraform import ironwifi_org_unit.example <org-unit-id>
```
