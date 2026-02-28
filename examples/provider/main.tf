terraform {
  required_providers {
    ironwifi = {
      source  = "ironwifi/ironwifi"
      version = "~> 0.1"
    }
  }
}

variable "ironwifi_api_token" {
  type      = string
  sensitive = true
}

variable "ironwifi_company_id" {
  type = string
}

provider "ironwifi" {
  api_endpoint = "https://console.ironwifi.com"
  api_token    = var.ironwifi_api_token
  company_id   = var.ironwifi_company_id
}
