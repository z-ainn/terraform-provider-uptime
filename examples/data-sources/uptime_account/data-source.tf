# Retrieve information about the current account
data "uptime_account" "current" {}

output "account_email" {
  value = data.uptime_account.current.email
}

output "monitor_usage" {
  value = "${data.uptime_account.current.monitors_count}/${data.uptime_account.current.monitors_limit} monitors used"
}