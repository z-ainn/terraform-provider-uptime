terraform {
  required_providers {
    uptime = {
      source = "registry.terraform.io/codematters-llc/uptime"
    }
  }
}

provider "uptime" {
  api_key  = var.uptime_api_key
  base_url = "https://api.uptime-monitor.io" # Optional, defaults to production
}

variable "uptime_api_key" {
  description = "Uptime Monitor API key"
  type        = string
  sensitive   = true
}

# Example 1: Full HTTPS monitor with all options
resource "uptime_monitor" "full_https_monitor" {
  name           = "🌐 Production API"
  url            = "https://api.example.com/health"
  type           = "https"
  active         = true
  check_interval = 60      # Check every 60 seconds
  timeout        = 30      # 30 second timeout
  fail_threshold = 3       # Mark as down after 3 consecutive failures
  
  regions = [
    "us-east-1",
    "us-west-1",
    "eu-west-1",
    "eu-central-1",
    "ap-southeast-1"
  ]
  
  contacts = [
    "contact_id_1",  # Replace with actual contact IDs
    "contact_id_2"
  ]
  
  https_settings = {
    method                       = "POST"
    expected_status_codes        = "200,201,202"
    check_certificate_expiration = true
    follow_redirects            = false
    
    request_headers = {
      "Authorization" = "Bearer ${var.api_token}"
      "Content-Type"  = "application/json"
      "X-API-Version" = "v2"
    }
    
    request_body = jsonencode({
      test = "health_check"
      timestamp = "{{timestamp}}"
    })
    
    expected_response_body = "\"status\":\"healthy\""
    
    expected_response_headers = {
      "X-API-Version" = "v2"
      "Content-Type"  = "application/json"
    }
  }
  
  # Certificate monitoring fields (auto-extracted from URL)
  # host and port are computed from the URL
}

# Example 2: Simple HTTPS monitor
resource "uptime_monitor" "simple_https" {
  name           = "🏠 Company Website"
  url            = "https://www.example.com"
  type           = "https"
  active         = true
  check_interval = 300     # Check every 5 minutes
  timeout        = 10      # 10 second timeout
  fail_threshold = 2       # Mark as down after 2 failures
  
  regions = ["us-east-1", "eu-west-1"]
  
  https_settings = {
    method = "HEAD"
    expected_status_codes        = "200-299"  # Accept any 2xx status
    check_certificate_expiration = true
    follow_redirects            = true
  }
}

# Example 3: TCP monitor for database
resource "uptime_monitor" "database_tcp" {
  name           = "🗄️ PostgreSQL Database"
  url            = "db.example.com:5432"
  type           = "tcp"
  active         = true
  check_interval = 60
  timeout        = 5
  fail_threshold = 2
  
  regions = ["us-east-1", "us-west-1"]
  contacts = ["contact_id_1"]
  
  tcp_settings = {
    # TCP monitors just check port connectivity
  }
}

# Example 4: TCP monitor for Redis
resource "uptime_monitor" "redis_tcp" {
  name           = "⚡ Redis Cache"
  url            = "redis.example.com:6379"
  type           = "tcp"
  active         = true
  check_interval = 30
  timeout        = 3
  fail_threshold = 1  # Very sensitive - alert immediately
  
  regions = ["us-east-1"]
  
  tcp_settings = {}
}

# Example 5: Ping monitor for network device
resource "uptime_monitor" "network_ping" {
  name           = "🌐 Gateway Router"
  url            = "gateway.example.com"
  type           = "ping"
  active         = true
  check_interval = 60
  timeout        = 5
  fail_threshold = 3
  
  regions = ["us-east-1", "eu-west-1"]
  
  ping_settings = {
    # Ping monitors use ICMP to check host availability
  }
}

# Example 6: Ping monitor for internal server
resource "uptime_monitor" "internal_server_ping" {
  name           = "🖥️ Internal Application Server"
  url            = "10.0.1.50"
  type           = "ping"
  active         = true
  check_interval = 120
  timeout        = 10
  fail_threshold = 2
  
  regions = ["us-east-1"]
  contacts = ["contact_id_3"]
  
  ping_settings = {}
}

# Example 7: HTTPS monitor with custom certificate monitoring
resource "uptime_monitor" "custom_cert_monitor" {
  name           = "🔒 API with Custom Certificate"
  url            = "https://secure-api.example.com:8443/status"
  type           = "https"
  active         = true
  check_interval = 300
  timeout        = 15
  fail_threshold = 2
  
  regions = ["us-east-1", "eu-west-1"]
  
  https_settings = {
    method = "HEAD"
    expected_status_codes        = "200"
    check_certificate_expiration = true
    follow_redirects            = false
  }
  
  # These will be auto-extracted from the URL
  # host = "secure-api.example.com"
  # port = 8443
}

# Example 8: Development environment monitor (initially paused)
resource "uptime_monitor" "dev_environment" {
  name           = "🔧 Development Environment"
  url            = "https://dev.example.com"
  type           = "https"
  active         = false  # Start in paused state
  check_interval = 600    # Check every 10 minutes when active
  timeout        = 30
  fail_threshold = 5      # More tolerant for dev environment
  
  regions = ["us-east-1"]
  
  https_settings = {
    method = "HEAD"
    expected_status_codes        = "200-299,301,302"
    check_certificate_expiration = false  # Don't check cert for dev
    follow_redirects            = true
  }
}

# Example 9: Multi-region critical service
resource "uptime_monitor" "critical_service" {
  name           = "🚨 Critical Payment Gateway"
  url            = "https://payments.example.com/health"
  type           = "https"
  active         = true
  check_interval = 30     # Check every 30 seconds
  timeout        = 10     # Quick timeout for fast detection
  fail_threshold = 1      # Alert immediately on first failure
  
  # Monitor from ALL available regions for maximum coverage
  regions = [
    "us-east-1",
    "us-east-2", 
    "us-west-1",
    "us-west-2",
    "eu-west-1",
    "eu-central-1",
    "eu-north-1",
    "ap-southeast-1",
    "ap-southeast-2",
    "ap-northeast-1"
  ]
  
  # Alert all contacts
  contacts = [
    "contact_id_1",
    "contact_id_2",
    "contact_id_3",
    "contact_id_4"
  ]
  
  https_settings = {
    method = "HEAD"
    expected_status_codes        = "200"
    check_certificate_expiration = true
    follow_redirects            = false
    expected_response_body      = "\"status\":\"operational\""
  }
}

# Example 10: Load balancer health check
resource "uptime_monitor" "load_balancer" {
  name           = "⚖️ Load Balancer"
  url            = "lb.example.com:443"
  type           = "tcp"
  active         = true
  check_interval = 60
  timeout        = 5
  fail_threshold = 2
  
  regions = ["us-east-1", "us-west-1"]
  
  tcp_settings = {}
}

# Outputs for reference
output "https_monitor_ids" {
  value = {
    full_api     = uptime_monitor.full_https_monitor.id
    simple_site  = uptime_monitor.simple_https.id
    custom_cert  = uptime_monitor.custom_cert_monitor.id
    dev_env      = uptime_monitor.dev_environment.id
    critical     = uptime_monitor.critical_service.id
  }
}

output "tcp_monitor_ids" {
  value = {
    database     = uptime_monitor.database_tcp.id
    redis        = uptime_monitor.redis_tcp.id
    load_balancer = uptime_monitor.load_balancer.id
  }
}

output "ping_monitor_ids" {
  value = {
    gateway      = uptime_monitor.network_ping.id
    internal     = uptime_monitor.internal_server_ping.id
  }
}