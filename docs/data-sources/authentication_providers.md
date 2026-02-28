# ironwifi_auth_providers

Lists IronWiFi authentication providers with optional name filtering.

## Example Usage

```hcl
data "ironwifi_auth_providers" "all" {}

data "ironwifi_auth_providers" "ldap" {
  name_filter = "LDAP"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter authentication providers by name (substring match).

## Attribute Reference

- `auth_providers` - List of authentication providers. Each provider contains:
  - `id` - Provider ID (UUID).
  - `name` - Provider name.
  - `type` - Provider type (ldap, saml, social, etc.).
  - `status` - Provider status.
