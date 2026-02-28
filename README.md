# Terraform Provider for IronWiFi

Manage IronWiFi cloud RADIUS, captive portals, and network access infrastructure as code.

## Quick Start

```hcl
terraform {
  required_providers {
    ironwifi = {
      source  = "ironwifi/ironwifi"
      version = "~> 0.1"
    }
  }
}

provider "ironwifi" {
  api_token  = var.ironwifi_api_token
  company_id = var.ironwifi_company_id
}

resource "ironwifi_network" "office" {
  name   = "Office-WiFi"
  region = "us-east1"
  coa    = true
}

resource "ironwifi_user" "admin" {
  username  = "admin@example.com"
  password  = "SecureP@ss"
  email     = "admin@example.com"
  user_type = "e"
}
```

## Authentication

The provider supports two authentication methods:

### API Token (recommended)

```hcl
provider "ironwifi" {
  api_token  = var.ironwifi_api_token
  company_id = var.ironwifi_company_id
}
```

Or via environment variables:

```bash
export IRONWIFI_API_TOKEN="your-api-token"
export IRONWIFI_COMPANY_ID="your-company-id"
```

### OAuth2 (username/password)

```hcl
provider "ironwifi" {
  username      = var.ironwifi_username
  password      = var.ironwifi_password
  client_id     = var.ironwifi_client_id
  client_secret = var.ironwifi_client_secret
  company_id    = var.ironwifi_company_id
}
```

Or via environment variables:

```bash
export IRONWIFI_USERNAME="your-username"
export IRONWIFI_PASSWORD="your-password"
export IRONWIFI_CLIENT_ID="your-client-id"
export IRONWIFI_CLIENT_SECRET="your-client-secret"
export IRONWIFI_COMPANY_ID="your-company-id"
```

### Provider Arguments

| Argument       | Description                                     | Required | Env Variable              |
|----------------|-------------------------------------------------|----------|---------------------------|
| `api_endpoint` | API base URL (default: `https://console.ironwifi.com`) | No | `IRONWIFI_API_ENDPOINT` |
| `api_token`    | API token for authentication                    | No*      | `IRONWIFI_API_TOKEN`      |
| `username`     | OAuth2 username                                 | No*      | `IRONWIFI_USERNAME`       |
| `password`     | OAuth2 password                                 | No*      | `IRONWIFI_PASSWORD`       |
| `client_id`    | OAuth2 client ID                                | No*      | `IRONWIFI_CLIENT_ID`      |
| `client_secret`| OAuth2 client secret                            | No*      | `IRONWIFI_CLIENT_SECRET`  |
| `company_id`   | Company/tenant ID                               | Yes      | `IRONWIFI_COMPANY_ID`     |

\* Either `api_token` or `username`/`password` must be provided.

## Resources

| Resource                              | Description                          |
|---------------------------------------|--------------------------------------|
| `ironwifi_network`                    | RADIUS network (SSID + auth config)  |
| `ironwifi_user`                       | End user / guest account             |
| `ironwifi_group`                      | User group                           |
| `ironwifi_policy`                     | Conditional access policy            |
| `ironwifi_authentication_provider`    | External IdP (LDAP, SAML, OAuth2)    |
| `ironwifi_captive_portal`             | Captive portal configuration         |
| `ironwifi_device`                     | Network device / access point        |
| `ironwifi_certificate`                | X.509 certificate                    |
| `ironwifi_profile`                    | EAP / 802.1X profile                 |
| `ironwifi_connector`                  | On-premise connector                 |
| `ironwifi_voucher`                    | Access voucher code                  |
| `ironwifi_org_unit`                   | Organizational unit                  |

## Data Sources

| Data Source                           | Description                          |
|---------------------------------------|--------------------------------------|
| `ironwifi_networks`                   | List/filter networks                 |
| `ironwifi_users`                      | List/filter users                    |
| `ironwifi_groups`                     | List/filter groups                   |
| `ironwifi_policies`                   | List/filter policies                 |
| `ironwifi_devices`                    | List/filter devices                  |
| `ironwifi_authentication_providers`   | List/filter auth providers           |

## Development

### Prerequisites

- Go 1.25+
- Terraform 1.0+

### Build

```bash
make build
```

### Install locally

```bash
make install
```

This installs to `~/.terraform.d/plugins/registry.terraform.io/ironwifi/ironwifi/0.1.0/darwin_arm64/`.

### Run tests

```bash
# Unit tests
make test

# Acceptance tests (requires live API credentials)
make testacc
```

### Lint

```bash
make lint
```

### Project structure

```
.
├── main.go                          # Provider entry point
├── internal/
│   ├── provider/                    # Provider configuration and schema
│   │   ├── provider.go
│   │   └── provider_test.go
│   ├── client/                      # HTTP client for IronWiFi API
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── auth.go
│   │   └── models.go
│   ├── resources/                   # Terraform resource implementations
│   │   ├── network_resource.go
│   │   ├── user_resource.go
│   │   ├── group_resource.go
│   │   └── ...
│   └── datasources/                 # Terraform data source implementations
│       ├── networks_data_source.go
│       └── ...
├── examples/                        # Example Terraform configurations
│   ├── provider/
│   ├── resources/
│   └── data-sources/
└── docs/                            # Generated documentation
    ├── resources/
    └── data-sources/
```
