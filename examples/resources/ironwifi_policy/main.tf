resource "ironwifi_policy" "restrict_guest" {
  name        = "Guest Bandwidth Limit"
  description = "Limit guest users to 10 Mbps"
  priority    = 50
  enabled     = true
  match_mode  = "all"
  target_type = "group"
  conditions = jsonencode([{
    type  = "user_group"
    value = "guests"
  }])
  actions = jsonencode([{
    type  = "bandwidth_limit"
    value = "10"
  }])
}
