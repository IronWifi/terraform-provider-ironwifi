# ironwifi_device

Manages an IronWiFi device (MAC-authenticated).

## Example Usage

```hcl
resource "ironwifi_device" "printer" {
  name       = "AA:BB:CC:DD:EE:FF"
  email      = "admin@example.com"
  firstname  = "Office"
  lastname   = "Printer"
  notes      = "3rd floor network printer"
  authsource = "local"
}
```

## Argument Reference

- `name` - (Required) Device identifier (MAC address).
- `email` - (Optional) Email address associated with the device.
- `firstname` - (Optional) First name of the device owner.
- `lastname` - (Optional) Last name of the device owner.
- `notes` - (Optional) Notes about the device.
- `mobilephone` - (Optional) Mobile phone number.
- `authsource` - (Optional) Authentication source. Defaults to `local`.
- `orgunit` - (Optional) Organizational unit.

## Attribute Reference

- `id` - Device ID (UUID).
- `status` - Device status.
- `creationdate` - Creation date.

## Import

```bash
terraform import ironwifi_device.example <device-id>
```
