# ironwifi_certificate

Manages an IronWiFi certificate.

## Example Usage

```hcl
resource "ironwifi_certificate" "employee_cert" {
  user_id      = ironwifi_user.employee.id
  cn           = "john.doe"
  subject      = "CN=john.doe,O=Example Corp"
  validity     = 365
  distribution = "email"
  hash         = "sha2"
}
```

## Argument Reference

- `user_id` - (Required) UUID of the user this certificate belongs to.
- `cn` - (Optional) Common name.
- `subject` - (Optional) Certificate subject.
- `validity` - (Optional) Validity period in days. Defaults to `365`.
- `distribution` - (Optional) Distribution method. Defaults to `email`.
- `hash` - (Optional) Hash algorithm. Defaults to `sha2`.
- `status` - (Optional) Certificate status: `valid`, `revoked`, or `pending`. Defaults to `valid`.

## Attribute Reference

- `id` - Certificate ID (UUID).
- `serial` - Certificate serial number.
- `expirationdate` - Expiration date.
- `revocationdate` - Revocation date.
- `creationdate` - Creation date.

## Import

```bash
terraform import ironwifi_certificate.example <certificate-id>
```
