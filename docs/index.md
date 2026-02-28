---
page_title: "IronWiFi Provider"
subcategory: ""
description: |-
  Manage IronWiFi resources as infrastructure-as-code.
---

# IronWiFi Provider

The IronWiFi provider allows you to manage WiFi networks, users, groups, policies, captive portals, and other IronWiFi resources using Terraform.

## Authentication

The provider supports two authentication methods:

### API Token (Recommended)

```hcl
provider "ironwifi" {
  api_token  = var.ironwifi_api_token
  company_id = var.ironwifi_company_id
}
```

### OAuth2 Credentials

```hcl
provider "ironwifi" {
  username      = var.ironwifi_username
  password      = var.ironwifi_password
  client_id     = "testclient"
  client_secret = "testpass"
  company_id    = var.ironwifi_company_id
}
```

## Environment Variables

All provider attributes can be set via environment variables:

| Attribute | Environment Variable |
|-----------|---------------------|
| `api_endpoint` | `IRONWIFI_API_ENDPOINT` |
| `api_token` | `IRONWIFI_API_TOKEN` |
| `company_id` | `IRONWIFI_COMPANY_ID` |
| `username` | `IRONWIFI_USERNAME` |
| `password` | `IRONWIFI_PASSWORD` |
| `client_id` | `IRONWIFI_CLIENT_ID` |
| `client_secret` | `IRONWIFI_CLIENT_SECRET` |

## Multi-Region

Target specific regions by setting `api_endpoint`:

```hcl
provider "ironwifi" {
  alias        = "europe"
  api_endpoint = "https://europe-west1.ironwifi.com"
  api_token    = var.ironwifi_api_token
  company_id   = var.ironwifi_company_id
}
```

## Example Usage

{{ tffile "examples/provider/main.tf" }}

{{ .SchemaMarkdown | trimspace }}
