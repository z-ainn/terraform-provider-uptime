# uptime_status_page (Resource)

The `uptime_status_page` resource allows you to create and manage status pages for displaying monitor statuses publicly or with authentication.

## Example Usage

### Basic Status Page

```hcl
resource "uptime_status_page" "public" {
  name = "Public Service Status"
  monitors = [
    uptime_monitor.api.id,
    uptime_monitor.website.id
  ]
  period = 7
  show_incident_reasons = true
}
```

### Status Page with Custom Domain

```hcl
resource "uptime_status_page" "custom" {
  name         = "Company Status Dashboard"
  monitors     = [uptime_monitor.main.id]
  period       = 30
  custom_domain = "status.example.com"
  show_incident_reasons = false
}
```

### Protected Status Page with Basic Auth

```hcl
resource "uptime_status_page" "internal" {
  name = "Internal Systems Status"
  monitors = [
    uptime_monitor.database.id,
    uptime_monitor.cache.id
  ]
  period = 90
  basic_auth = "admin:SecurePassword123"
  show_incident_reasons = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Display name for the status page. Must be between 1 and 100 characters.
* `monitors` - (Required) List of monitor IDs to display on the status page. Must contain between 1 and 20 monitor IDs.
* `period` - (Optional) Time period in days for uptime statistics. Valid values are `7`, `30`, or `90`. Defaults to `7`.
* `custom_domain` - (Optional) Custom domain for accessing the status page (e.g., `status.example.com`). Must be a valid domain name with at least one dot, cannot end with "uptime-monitor.io", and cannot contain forward slashes.
* `show_incident_reasons` - (Optional) Whether to show incident reasons publicly on the status page. Defaults to `false`.
* `basic_auth` - (Optional, Sensitive) Basic authentication credentials in `username:password` format. When set, visitors must authenticate to view the status page.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the status page.
* `created_at` - Unix timestamp when the status page was created.
* `url` - The URL where the status page can be accessed. This will be either the custom domain URL or the default status page URL.

## Import

Status pages can be imported using their ID:

```shell
terraform import uptime_status_page.example <status_page_id>
```