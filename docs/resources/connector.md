# ironwifi_connector

Manages an IronWiFi authentication connector (LDAP, AD, SAML, OAuth).

## Example Usage

```hcl
resource "ironwifi_connector" "corp_ldap" {
  name       = "Corporate LDAP"
  type       = "ldap"
  domain     = "example.com"
  basedn     = "dc=example,dc=com"
  bind       = "cn=admin,dc=example,dc=com"
  password   = var.ldap_password
  authsource = "ldap"
  status     = "enabled"
}
```

## Argument Reference

- `name` - (Required) Connector name.
- `type` - (Required) Connector type: `ldap`, `ad`, `saml`, or `oauth`.
- `domain` - (Optional) Domain for the connector.
- `group` - (Optional) Group identifier.
- `groupname` - (Optional) Group name.
- `status` - (Optional) Connector status. Defaults to `enabled`.
- `authsource` - (Optional) Authentication source.
- `basedn` - (Optional) Base DN for LDAP.
- `bind` - (Optional) Bind DN for LDAP.
- `password` - (Optional, Sensitive) Password. Write-only; not returned by the API.
- `sync_interval` - (Optional) Sync interval in minutes.
- `user_takeover` - (Optional) Enable user takeover. Defaults to `false`.
- `client_id` - (Optional) Client ID for OAuth/SAML.
- `client_secret` - (Optional, Sensitive) Client secret for OAuth/SAML. Write-only.

## Attribute Reference

- `id` - Connector ID (UUID).
- `creationdate` - Creation date.

## Import

```bash
terraform import ironwifi_connector.example <connector-id>
```
