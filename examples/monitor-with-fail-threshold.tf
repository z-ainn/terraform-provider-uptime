terraform {
  required_providers {
    uptime = {
      source = "registry.terraform.io/uptime-monitor-io/uptime"
    }
  }
}

provider "uptime" {
  api_key = var.uptime_api_key
  base_url = "https://api.uptime-monitor.io" # Optional
}

variable "uptime_api_key" {
  description = "Uptime Monitor API key"
  type        = string
  sensitive   = true
}

# Example: HTTPS monitor with custom fail_threshold and contacts
resource "uptime_monitor" "example_https" {
  name           = "Example Website"
  url            = "https://example.com"
  type           = "https"
  active         = true
  check_interval = 60
  timeout        = 30
  fail_threshold = 3  # Will mark as down after 3 consecutive failures
  
  regions = [
    "us-east-1",
    "eu-west-1",
    "ap-southeast-1"
  ]
  
  contacts = [
    "contact_id_1",  # Replace with actual contact IDs
    "contact_id_2"
  ]
  
  https_settings = {
    method                       = "HEAD"
    expected_status_codes        = "200-299"
    check_certificate_expiration = true
    follow_redirects            = true
    
    expected_response_body = "Welcome to Example"
    
    request_headers = {
      "User-Agent" = "Uptime Monitor"
      "Accept"     = "text/html"
    }
    
    expected_response_headers = {
      "Content-Type" = "text/html"
    }
  }
}

# Example: TCP monitor with fail_threshold
resource "uptime_monitor" "example_tcp" {
  name           = "Database Server"
  url            = "db.example.com:5432"
  type           = "tcp"
  active         = true
  check_interval = 30
  timeout        = 10
  fail_threshold = 2  # Will mark as down after 2 consecutive failures
  
  regions = ["us-east-1"]
  contacts = ["contact_id_1"]
}

# Example: Ping monitor (paused initially)
resource "uptime_monitor" "example_ping" {
  name           = "Server Ping Check"
  url            = "server.example.com"
  type           = "ping"
  active         = false  # Start in paused state
  check_interval = 120
  timeout        = 5
  fail_threshold = 1  # Very sensitive - mark down after 1 failure
  
  regions = ["us-east-1", "eu-west-1"]
}

# Output the monitor IDs for reference
output "https_monitor_id" {
  value = uptime_monitor.example_https.id
}

output "tcp_monitor_id" {
  value = uptime_monitor.example_tcp.id
}

output "ping_monitor_id" {
  value = uptime_monitor.example_ping.id
}