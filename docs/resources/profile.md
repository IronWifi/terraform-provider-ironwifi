# ironwifi_profile

Manages an IronWiFi authentication profile.

## Example Usage

```hcl
resource "ironwifi_profile" "eap_tls" {
  name        = "EAP-TLS Enterprise"
  description = "802.1X EAP-TLS profile"
  type        = "EAP-TLS"
}
```

## Argument Reference

- `name` - (Required) Profile name.
- `description` - (Optional) Profile description.
- `type` - (Optional) Profile type (e.g. `PEAP`, `EAP-TLS`).
- `configuration` - (Optional) Profile configuration as JSON string.

## Attribute Reference

- `id` - Profile ID (UUID).

## Import

```bash
terraform import ironwifi_profile.example <profile-id>
```
