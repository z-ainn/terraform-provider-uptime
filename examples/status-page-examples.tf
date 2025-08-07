# Terraform Examples for Uptime Monitor Status Pages
# This file demonstrates various configurations for status pages

# -----------------------------------------------------------------------------
# BASIC STATUS PAGE
# -----------------------------------------------------------------------------
# A simple public status page showing multiple monitors
resource "uptime_status_page" "basic" {
  name = "Service Status"
  monitors = [
    uptime_monitor.api.id,
    uptime_monitor.website.id,
    uptime_monitor.database.id
  ]
  period = 7  # Show 7-day uptime statistics
}

# -----------------------------------------------------------------------------
# STATUS PAGE WITH CUSTOM DOMAIN
# -----------------------------------------------------------------------------
# Status page accessible via a custom domain
resource "uptime_status_page" "custom_domain" {
  name         = "Acme Corp Status"
  monitors     = [uptime_monitor.production.id]
  period       = 30  # Show 30-day uptime statistics
  custom_domain = "status.acme-corp.com"
  show_incident_reasons = false  # Hide detailed incident reasons from public
}

# -----------------------------------------------------------------------------
# PROTECTED STATUS PAGE WITH BASIC AUTH
# -----------------------------------------------------------------------------
# Internal status page requiring authentication
resource "uptime_status_page" "internal" {
  name = "Internal Infrastructure Status"
  monitors = [
    uptime_monitor.database.id,
    uptime_monitor.redis.id,
    uptime_monitor.elasticsearch.id,
    uptime_monitor.kafka.id
  ]
  period = 90  # Show 90-day uptime statistics
  basic_auth = "admin:${var.status_page_password}"  # Use variable for password
  show_incident_reasons = true  # Show detailed reasons to authenticated users
}

# -----------------------------------------------------------------------------
# PUBLIC STATUS PAGE WITH INCIDENT DETAILS
# -----------------------------------------------------------------------------
# Transparent status page showing all incident information
resource "uptime_status_page" "transparent" {
  name = "Public API Status"
  monitors = [
    uptime_monitor.api_v1.id,
    uptime_monitor.api_v2.id,
    uptime_monitor.webhooks.id
  ]
  period = 7
  show_incident_reasons = true  # Full transparency on incidents
}

# -----------------------------------------------------------------------------
# MINIMAL STATUS PAGE
# -----------------------------------------------------------------------------
# Single monitor status page with minimal configuration
resource "uptime_status_page" "minimal" {
  name     = "Main Site Status"
  monitors = [uptime_monitor.main_website.id]
  # Uses defaults: period = 7, show_incident_reasons = false, no auth
}

# -----------------------------------------------------------------------------
# MULTI-REGION STATUS PAGE
# -----------------------------------------------------------------------------
# Status page showing monitors from different regions
resource "uptime_status_page" "global" {
  name = "Global Infrastructure Status"
  monitors = [
    uptime_monitor.us_east.id,
    uptime_monitor.us_west.id,
    uptime_monitor.eu_central.id,
    uptime_monitor.asia_pacific.id
  ]
  period = 30
  custom_domain = "status.global-company.io"
}

# -----------------------------------------------------------------------------
# DYNAMIC STATUS PAGE WITH FOR_EACH
# -----------------------------------------------------------------------------
# Create multiple status pages dynamically
variable "status_pages" {
  type = map(object({
    name              = string
    monitors          = list(string)
    period            = number
    custom_domain     = optional(string)
    show_incidents    = optional(bool, false)
    basic_auth        = optional(string)
  }))
  default = {
    production = {
      name           = "Production Status"
      monitors       = ["monitor1", "monitor2"]
      period         = 7
      custom_domain  = "status.prod.example.com"
      show_incidents = true
    }
    staging = {
      name           = "Staging Status"
      monitors       = ["monitor3", "monitor4"]
      period         = 30
      basic_auth     = "stage:password123"
    }
  }
}

resource "uptime_status_page" "dynamic" {
  for_each = var.status_pages

  name                  = each.value.name
  monitors              = each.value.monitors
  period                = each.value.period
  custom_domain         = each.value.custom_domain
  show_incident_reasons = each.value.show_incidents
  basic_auth            = each.value.basic_auth
}

# -----------------------------------------------------------------------------
# DATA SOURCE EXAMPLE
# -----------------------------------------------------------------------------
# Read an existing status page
data "uptime_status_page" "existing" {
  id = "507f1f77bcf86cd799439011"
}

# Use the data source information
output "existing_status_page_info" {
  value = {
    name     = data.uptime_status_page.existing.name
    url      = data.uptime_status_page.existing.url
    monitors = data.uptime_status_page.existing.monitors
    period   = data.uptime_status_page.existing.period
  }
}

# -----------------------------------------------------------------------------
# OUTPUTS
# -----------------------------------------------------------------------------
output "status_page_urls" {
  description = "URLs of all created status pages"
  value = {
    basic         = uptime_status_page.basic.url
    custom_domain = uptime_status_page.custom_domain.url
    internal      = uptime_status_page.internal.url
    transparent   = uptime_status_page.transparent.url
    minimal       = uptime_status_page.minimal.url
    global        = uptime_status_page.global.url
  }
}

output "status_page_ids" {
  description = "IDs of all created status pages"
  value = {
    basic         = uptime_status_page.basic.id
    custom_domain = uptime_status_page.custom_domain.id
    internal      = uptime_status_page.internal.id
    transparent   = uptime_status_page.transparent.id
    minimal       = uptime_status_page.minimal.id
    global        = uptime_status_page.global.id
  }
  sensitive = false  # IDs are not sensitive
}

# -----------------------------------------------------------------------------
# VARIABLES FOR CONFIGURATION
# -----------------------------------------------------------------------------
variable "status_page_password" {
  description = "Password for protected status pages"
  type        = string
  sensitive   = true
  default     = "SecurePassword123"
}

variable "enable_custom_domains" {
  description = "Whether to use custom domains for status pages"
  type        = bool
  default     = true
}

# -----------------------------------------------------------------------------
# CONDITIONAL STATUS PAGE
# -----------------------------------------------------------------------------
# Create status page conditionally based on environment
resource "uptime_status_page" "conditional" {
  count = var.environment == "production" ? 1 : 0

  name = "Production Environment Status"
  monitors = [
    uptime_monitor.prod_api.id,
    uptime_monitor.prod_web.id
  ]
  period = 30
  custom_domain = var.enable_custom_domains ? "status.prod.example.com" : null
  show_incident_reasons = true
}

variable "environment" {
  description = "Current environment"
  type        = string
  default     = "production"
}