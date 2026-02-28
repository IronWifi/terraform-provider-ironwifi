# ironwifi_users

Lists IronWiFi users with optional name filtering.

## Example Usage

```hcl
data "ironwifi_users" "all" {}

data "ironwifi_users" "admins" {
  name_filter = "admin"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter users by username (substring match).

## Attribute Reference

- `users` - List of users. Each user contains:
  - `id` - User ID (UUID).
  - `username` - Username.
  - `email` - Email address.
  - `firstname` - First name.
  - `lastname` - Last name.
