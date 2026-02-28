# ironwifi_policy

Manages an IronWiFi policy (conditional access).

## Example Usage

```hcl
resource "ironwifi_policy" "bandwidth_limit" {
  name        = "Bandwidth Limit"
  description = "Limit bandwidth for guest users"
  priority    = 50
  enabled     = true
  match_mode  = "all"
  target_type = "group"
  target_id   = ironwifi_group.guests.id
  conditions  = jsonencode([{"type" = "user_group"}])
  actions     = jsonencode([{"type" = "bandwidth_limit", "value" = "10M"}])
}
```

## Argument Reference

- `name` - (Required) Policy name.
- `description` - (Optional) Policy description.
- `priority` - (Optional) Evaluation priority (lower values evaluated first). Defaults to `100`.
- `enabled` - (Optional) Whether the policy is enabled. Defaults to `false`.
- `match_mode` - (Optional) Condition matching mode: `all` or `any`. Defaults to `all`.
- `target_type` - (Optional) Target type: `global`, `network`, `group`, etc. Defaults to `global`.
- `target_id` - (Optional) Target entity ID (when target_type is not `global`).
- `conditions` - (Optional) JSON string defining policy conditions.
- `actions` - (Optional) JSON string defining policy actions.

## Attribute Reference

- `id` - Policy ID (UUID).
- `created_at` - Date the policy was created.
- `updated_at` - Date the policy was last updated.

## Import

```bash
terraform import ironwifi_policy.example <policy-id>
```
