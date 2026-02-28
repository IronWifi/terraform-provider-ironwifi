resource "ironwifi_device" "printer" {
  name       = "AA:BB:CC:DD:EE:FF"
  email      = "admin@example.com"
  firstname  = "Office"
  lastname   = "Printer"
  notes      = "3rd floor network printer"
  authsource = "local"
}
