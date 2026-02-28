resource "ironwifi_captive_portal" "guest" {
  name               = "Guest Portal"
  description        = "Guest WiFi captive portal"
  splash_page        = "https://splash.example.com"
  mac_authentication = true
  cloud_cdn          = true
}
