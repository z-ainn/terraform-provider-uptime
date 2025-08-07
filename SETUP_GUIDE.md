# Setting Up a Terraform Project with Uptime Monitor Provider

This guide walks you through setting up a new Terraform project to manage your Uptime Monitor infrastructure.

## Prerequisites

- Terraform installed on your machine
- Uptime Monitor API key
- Built terraform provider binary

## Step 1: Build the Provider

```bash
cd /path/to/uptime/terraform-provider
make build
```

This creates the `terraform-provider-uptime` binary.

## Step 2: Create a New Terraform Project

```bash
mkdir my-uptime-terraform-project
cd my-uptime-terraform-project
```

## Step 3: Create Provider Override Configuration

Create a `.terraformrc` file in your project directory:

```bash
cat > .terraformrc << 'EOF'
provider_installation {
  dev_overrides {
    "uptime-monitor/uptime" = "/path/to/uptime/terraform-provider"
  }
  direct {}
}
EOF
```

Replace `/path/to/uptime/terraform-provider` with the actual path to your provider directory (not the binary file itself).

## Step 4: Set Environment Variables

```bash
export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
export UPTIME_API_KEY="your-api-key-here"
export UPTIME_BASE_URL="https://uptime-monitor.io"
```

## Step 5: Create Terraform Configuration

Create `main.tf`:

```hcl
terraform {
  required_providers {
    uptime = {
      source = "uptime-monitor/uptime"
      version = "0.1.0"
    }
  }
}

provider "uptime" {
  # API key and base URL are set via environment variables
}

# Example monitor
resource "uptime_monitor" "example" {
  name           = "My Website"
  url            = "https://example.com/"
  type           = "https"
  check_interval = 60
  timeout        = 30
  regions        = ["us-east-1", "eu-west-1"]
  
  https_settings = {
    method                       = "get"
    check_certificate_expiration = true
    follow_redirects             = true
  }
}
```

## Step 6: Run Terraform Commands

**Important**: When using dev overrides, skip `terraform init`:

```bash
# View the plan
terraform plan

# Apply changes
terraform apply

# View current state
terraform show
```

## Step 7: Import Existing Monitors (Optional)

If you have existing monitors to import:

1. Copy the discovery script to your project:
```bash
cp /path/to/uptime/terraform-config/discover-monitors.sh .
chmod +x discover-monitors.sh
```

2. Run discovery:
```bash
./discover-monitors.sh
```

3. Copy the generated configuration to your `.tf` files

4. Import each monitor:
```bash
terraform import uptime_monitor.monitor_name monitor-id
```

## Complete Example Project Structure

```
my-uptime-terraform-project/
├── .terraformrc          # Provider override configuration
├── main.tf               # Main terraform configuration
├── monitors.tf           # Monitor definitions
├── terraform.tfstate     # Terraform state (created after first apply)
└── discover-monitors.sh  # Optional: for importing existing monitors
```

## Example `monitors.tf`

```hcl
# Production website monitor
resource "uptime_monitor" "production_website" {
  name           = "Production Website"
  url            = "https://mysite.com/"
  type           = "https"
  check_interval = 60
  timeout        = 30
  regions        = ["us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"]
  
  https_settings = {
    method                       = "get"
    expected_status_codes        = "200,301"
    check_certificate_expiration = true
    follow_redirects             = true
    request_headers = {
      "User-Agent" = "Uptime-Monitor"
    }
  }
}

# API health check
resource "uptime_monitor" "api_health" {
  name           = "API Health Check"
  url            = "https://api.mysite.com/health"
  type           = "https"
  check_interval = 300  # 5 minutes
  timeout        = 10
  regions        = ["us-east-1", "eu-west-1"]
  
  https_settings = {
    method                = "get"
    expected_status_codes = "200"
    expected_response_body = "ok"
  }
}

# TCP service monitor
resource "uptime_monitor" "database_port" {
  name           = "Database Port"
  url            = "db.mysite.com:5432"
  type           = "tcp"
  check_interval = 300
  timeout        = 5
  regions        = ["us-east-1"]
}
```

## Monitor Resource Reference

### Required Arguments

- `name` - (Required) Display name for the monitor
- `url` - (Required) The URL or endpoint to monitor
- `type` - (Required) Monitor type: `https`, `tcp`, or `ping`

### Optional Arguments

- `check_interval` - Check interval in seconds (default: 60)
- `timeout` - Request timeout in seconds (default: 30)
- `regions` - List of regions to perform checks from
- `https_settings` - HTTPS-specific configuration block (only for type="https")

### HTTPS Settings Block

For HTTPS monitors, you can configure:

- `method` - HTTP method: get, post, put, etc. (default: "get")
- `expected_status_codes` - Expected status codes (e.g., "200", "200-299", "200,201,301")
- `check_certificate_expiration` - Check SSL certificate expiration (default: true)
- `follow_redirects` - Follow HTTP redirects (default: true)
- `request_headers` - Map of HTTP headers to send
- `request_body` - Request body for POST/PUT requests
- `expected_response_body` - Expected response body content
- `expected_response_headers` - Map of expected response headers

## Data Source Reference

You can read existing monitors:

```hcl
data "uptime_monitor" "existing" {
  id = "monitor-id"
}

output "monitor_url" {
  value = data.uptime_monitor.existing.url
}
```

## Troubleshooting

### "Provider not found" error
Make sure `TF_CLI_CONFIG_FILE` is set correctly and points to your `.terraformrc` file.

### Authentication errors
Verify your API key is correct and the `UPTIME_API_KEY` environment variable is exported.

### "No changes" after import
This is expected - it means your configuration matches the imported state perfectly.

### "Inconsistent dependency lock file" error
When using dev overrides, skip `terraform init`. Remove any `.terraform` directory or `.terraform.lock.hcl` file if they exist.

## Best Practices

1. **Never commit API keys** - Always use environment variables
2. **Use consistent naming** - Follow a naming convention for your monitors
3. **Group related monitors** - Organize monitors in separate `.tf` files by service or environment
4. **Set appropriate intervals** - Don't over-monitor; adjust check intervals based on criticality
5. **Use meaningful regions** - Select regions close to your users and infrastructure
6. **Version control** - Keep your Terraform configurations in version control (without secrets)

## Notes

- This provider is private and requires local building
- Always use the dev override approach for private providers
- The provider binary must be built before use
- Environment variables are the recommended way to provide credentials