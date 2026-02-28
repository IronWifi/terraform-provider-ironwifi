# ironwifi_captive_portal

Manages an IronWiFi captive portal.

## Example Usage

```hcl
resource "ironwifi_captive_portal" "guest" {
  name               = "Guest Portal"
  description        = "Guest WiFi portal"
  vendor             = "meraki"
  network_id         = ironwifi_network.office.id
  mac_authentication = true
}
```

## Argument Reference

- `name` - (Required) Captive portal name.
- `description` - (Optional) Captive portal description.
- `vendor` - (Optional) Hardware vendor for the captive portal.
- `network_id` - (Optional) Associated network ID.
- `splash_page` - (Optional) Splash page URL or template.
- `success_page` - (Optional) Success/redirect page URL.
- `portal_theme` - (Optional) Portal theme identifier.
- `mac_authentication` - (Optional) Enable MAC-based authentication. Defaults to `false`.
- `cloud_cdn` - (Optional) Enable Cloud CDN for portal assets. Defaults to `false`.
- `webhook_url` - (Optional) Webhook URL for portal events.

## Attribute Reference

- `id` - Captive portal ID (UUID).

## Import

```bash
terraform import ironwifi_captive_portal.example <portal-id>
```
