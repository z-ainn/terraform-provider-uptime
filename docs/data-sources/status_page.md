# uptime_status_page (Data Source)

The `uptime_status_page` data source allows you to read information about an existing status page.

## Example Usage

```hcl
data "uptime_status_page" "example" {
  id = "507f1f77bcf86cd799439011"
}

output "status_page_url" {
  value = data.uptime_status_page.example.url
}

output "status_page_monitors" {
  value = data.uptime_status_page.example.monitors
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The unique identifier of the status page to read.

## Attributes Reference

The following attributes are exported:

* `name` - Display name for the status page.
* `monitors` - List of monitor IDs displayed on the status page.
* `period` - Time period in days for uptime statistics (7, 30, or 90).
* `custom_domain` - Custom domain for accessing the status page, if configured.
* `show_incident_reasons` - Whether incident reasons are shown publicly.
* `basic_auth` - Basic authentication credentials (sensitive field).
* `created_at` - Unix timestamp when the status page was created.
* `url` - The URL where the status page can be accessed.