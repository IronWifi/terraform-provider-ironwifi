# ironwifi_authentication_provider

Manages an IronWiFi authentication provider.

## Example Usage

```hcl
resource "ironwifi_authentication_provider" "ldap" {
  name              = "Corporate LDAP"
  type              = "ldap"
  captive_portal_id = ironwifi_captive_portal.guest.id
  group_id          = ironwifi_group.employees.id
  status            = "enabled"
  configuration     = jsonencode({"basedn" = "dc=example,dc=com"})
}
```

## Argument Reference

- `name` - (Required) Authentication provider name.
- `type` - (Required) Provider type: `ldap`, `saml`, `social`, `twilio`, etc.
- `captive_portal_id` - (Optional) Associated captive portal ID.
- `group_id` - (Optional) Associated group ID.
- `status` - (Optional) Provider status. Defaults to `enabled`.
- `configuration` - (Optional) JSON string with type-specific configuration.

## Attribute Reference

- `id` - Authentication provider ID (UUID).

## Import

```bash
terraform import ironwifi_authentication_provider.example <provider-id>
```
