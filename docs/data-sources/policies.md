# ironwifi_policies

Lists IronWiFi policies with optional name filtering.

## Example Usage

```hcl
data "ironwifi_policies" "all" {}

data "ironwifi_policies" "bandwidth" {
  name_filter = "Bandwidth"
}
```

## Argument Reference

- `name_filter` - (Optional) Filter policies by name (substring match).

## Attribute Reference

- `policies` - List of policies. Each policy contains:
  - `id` - Policy ID (UUID).
  - `name` - Policy name.
  - `enabled` - Whether the policy is enabled.
  - `priority` - Evaluation priority.
