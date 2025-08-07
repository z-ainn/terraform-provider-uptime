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

# Basic HTTPS monitor
resource "uptime_monitor" "website" {
  name           = "Website Health Check"
  url            = "https://example.com"
  type           = "https"
  check_interval = 60
  timeout        = 30
  regions        = ["us-east", "eu-west"]

  https_settings {
    method                        = "HEAD"
    expected_status_codes        = "200"
    check_certificate_expiration = true
    follow_redirects            = true
  }
}

# API endpoint with authentication
resource "uptime_monitor" "api" {
  name           = "API Health Check"
  url            = "https://api.example.com/health"
  type           = "https"
  check_interval = 120
  timeout        = 30
  regions        = ["us-east", "eu-west", "ap-southeast"]

  https_settings {
    method                = "HEAD"
    expected_status_codes = "200,201"
    
    request_headers = {
      "Authorization" = "Bearer api-token"
      "Content-Type"  = "application/json"
    }
    
    expected_response_body = "healthy"
  }
}

# TCP service monitor
resource "uptime_monitor" "database" {
  name           = "Database Connection"
  url            = "tcp://db.example.com:5432"
  type           = "tcp"
  check_interval = 300
  timeout        = 10
  regions        = ["us-east"]
}

# Ping monitor
resource "uptime_monitor" "server" {
  name           = "Server Ping"
  url            = "ping://server.example.com"
  type           = "ping"
  check_interval = 60
  timeout        = 5
  regions        = ["us-east", "eu-west"]
}

# Output monitor IDs
output "monitor_ids" {
  value = {
    website  = uptime_monitor.website.id
    api      = uptime_monitor.api.id
    database = uptime_monitor.database.id
    server   = uptime_monitor.server.id
  }
}