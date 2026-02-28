# ironwifi_user

Manages an IronWiFi user.

## Example Usage

```hcl
resource "ironwifi_user" "employee" {
  username   = "john.doe@example.com"
  password   = "SecureP@ssw0rd"
  email      = "john.doe@example.com"
  firstname  = "John"
  lastname   = "Doe"
  user_type  = "e"
  authsource = "local"
}
```

## Argument Reference

- `username` - (Required) Username for authentication.
- `password` - (Optional, Sensitive) User password. Write-only; not returned by the API.
- `email` - (Optional) User email address.
- `firstname` - (Optional) User first name.
- `lastname` - (Optional) User last name.
- `notes` - (Optional) Notes about the user.
- `user_type` - (Optional) User type: `e` for employee, `u` for user. Defaults to `e`.
- `mobilephone` - (Optional) User mobile phone number.
- `authsource` - (Optional) Authentication source. Defaults to `local`.
- `orgunit` - (Optional) Organizational unit.

## Attribute Reference

- `id` - User ID (UUID).
- `status` - User status.
- `deletiondate` - Deletion date.
- `creationdate` - Creation date.

## Import

```bash
terraform import ironwifi_user.example <user-id>
```
