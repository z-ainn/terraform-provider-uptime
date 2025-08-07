terraform {
  required_providers {
    uptime = {
      source  = "localhost/uptime/uptime"
      version = "dev"
    }
  }
}

provider "uptime" {
  api_key  = "your-api-key-here"
  base_url = "http://localhost:8000"
}

# Get current account information
data "uptime_account" "current" {}

# Example: Create monitors conditionally based on account limits
resource "uptime_monitor" "conditional_monitor" {
  count = data.uptime_account.current.monitors_count < data.uptime_account.current.monitors_limit ? 1 : 0

  name           = "Conditional Monitor ${count.index + 1}"
  url            = "https://example.com"
  type           = "https"
  check_interval = 60
  timeout        = 30
  regions        = ["us-east"]

  https_settings {
    method                = "HEAD"
    expected_status_codes = "200"
  }
}

# Output account information
output "account_info" {
  value = {
    email          = data.uptime_account.current.email
    current_plan   = data.uptime_account.current.current_plan
    monitors_limit = data.uptime_account.current.monitors_limit
    monitors_count = data.uptime_account.current.monitors_count
  }
  description = "Current account information and monitor statistics"
}

output "monitor_statistics" {
  value = {
    up_monitors     = data.uptime_account.current.up_monitors
    down_monitors   = data.uptime_account.current.down_monitors
    paused_monitors = data.uptime_account.current.paused_monitors
  }
  description = "Current monitor status breakdown"
}

output "capacity_info" {
  value = {
    monitors_used     = data.uptime_account.current.monitors_count
    monitors_limit    = data.uptime_account.current.monitors_limit
    monitors_available = data.uptime_account.current.monitors_limit - data.uptime_account.current.monitors_count
    at_capacity       = data.uptime_account.current.monitors_count >= data.uptime_account.current.monitors_limit
  }
  description = "Account capacity and usage information"
}

# Example: Local value demonstrating usage in conditions
locals {
  can_create_monitors = data.uptime_account.current.monitors_count < data.uptime_account.current.monitors_limit
  capacity_warning    = data.uptime_account.current.monitors_count > (data.uptime_account.current.monitors_limit * 0.8)
  has_down_monitors   = data.uptime_account.current.down_monitors > 0
}

# Example: Conditional warning message
output "capacity_warning" {
  value = local.capacity_warning ? "Warning: Account is using more than 80% of monitor capacity" : "Account capacity OK"
}

output "health_status" {
  value = local.has_down_monitors ? "Alert: ${data.uptime_account.current.down_monitors} monitors are DOWN" : "All monitors are healthy"
}