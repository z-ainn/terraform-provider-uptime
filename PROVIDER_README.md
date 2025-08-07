# Uptime Monitor Terraform Provider

This is a private Terraform provider for managing Uptime Monitor resources.

## Installation

1. Build the provider:
   ```bash
   make build
   ```

2. Create a `.terraformrc` file in your terraform configuration directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "uptime-monitor-io/uptime" = "/path/to/terraform-provider"
     }
     direct {}
   }
   ```

3. Set the environment variable to use the configuration:
   ```bash
   export TF_CLI_CONFIG_FILE=./.terraformrc
   ```

## Authentication

The provider uses API key authentication. You can configure it in two ways:

1. Environment variables (recommended):
   ```bash
   export UPTIME_API_KEY="your-api-key"
   export UPTIME_BASE_URL="https://uptime-monitor.io"  # Optional, defaults to production
   ```

2. Provider configuration:
   ```hcl
   provider "uptime" {
     api_key  = "your-api-key"
     base_url = "https://uptime-monitor.io"  # Optional
   }
   ```

## Usage

### Monitor Resource

The provider supports creating and managing uptime monitors.

```hcl
resource "uptime_monitor" "example" {
  name           = "Example Monitor"
  url            = "https://example.com/"
  type           = "https"
  check_interval = 60
  timeout        = 30
  regions        = ["us-east-1", "eu-west-1"]
  
  # HTTPS-specific settings (optional)
  https_settings = {
    method                       = "get"
    expected_status_codes        = "200,201"
    check_certificate_expiration = true
    follow_redirects             = true
    request_headers = {
      "User-Agent" = "Uptime-Monitor"
    }
  }
}
```

### Monitor Data Source

You can also read existing monitors:

```hcl
data "uptime_monitor" "existing" {
  id = "monitor-id"
}
```

## Importing Existing Monitors

To import existing monitors into Terraform:

1. Discover existing monitors:
   ```bash
   export UPTIME_API_KEY="your-api-key"
   ./discover-monitors.sh
   ```

2. Copy the generated terraform configuration to your `.tf` files

3. Run the import commands:
   ```bash
   terraform import uptime_monitor.monitor_name monitor-id
   ```

## Supported Monitor Types

- **HTTPS**: Web endpoint monitoring with SSL certificate checking
- **TCP**: TCP port connectivity monitoring
- **Ping**: ICMP ping monitoring (coming soon)

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Debugging

Run the provider with debug logging:
```bash
TF_LOG=DEBUG terraform plan
```