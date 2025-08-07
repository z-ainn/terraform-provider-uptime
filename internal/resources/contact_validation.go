package resources

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// validateContactRequest performs comprehensive validation for all contact types
func (r *ContactResource) validateContactRequest(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	// Validate that exactly one channel type is configured
	channelCount := 0
	var channelType string

	if !data.EmailSettings.IsNull() {
		channelCount++
		channelType = "email"
	}
	if !data.SmsSettings.IsNull() {
		channelCount++
		channelType = "sms"
	}
	if !data.WebhookSettings.IsNull() {
		channelCount++
		channelType = "webhook"
	}
	if !data.SlackSettings.IsNull() {
		channelCount++
		channelType = "slack"
	}
	if !data.DiscordSettings.IsNull() {
		channelCount++
		channelType = "discord"
	}
	if !data.PagerdutySettings.IsNull() {
		channelCount++
		channelType = "pagerduty"
	}
	if !data.IncidentioSettings.IsNull() {
		channelCount++
		channelType = "incidentio"
	}
	if !data.OpsgenieSettings.IsNull() {
		channelCount++
		channelType = "opsgenie"
	}
	if !data.ZendeskSettings.IsNull() {
		channelCount++
		channelType = "zendesk"
	}

	if channelCount == 0 {
		resp.Diagnostics.AddError(
			"Missing Channel Configuration",
			"At least one channel type must be configured (email, sms, webhook, slack, discord, pagerduty, incidentio, opsgenie, or zendesk)",
		)
		return
	}

	if channelCount > 1 {
		resp.Diagnostics.AddError(
			"Multiple Channel Configurations",
			"Only one channel type can be configured per contact",
		)
		return
	}

	// Perform channel-specific validation
	switch channelType {
	case "email":
		r.validateEmailSettings(ctx, data, resp)
	case "sms":
		r.validateSmsSettings(ctx, data, resp)
	case "webhook":
		r.validateWebhookSettings(ctx, data, resp)
	case "slack":
		r.validateSlackSettings(ctx, data, resp)
	case "discord":
		r.validateDiscordSettings(ctx, data, resp)
	case "pagerduty":
		r.validatePagerdutySettings(ctx, data, resp)
	case "incidentio":
		r.validateIncidentioSettings(ctx, data, resp)
	case "opsgenie":
		r.validateOpsgenieSettings(ctx, data, resp)
	case "zendesk":
		r.validateZendeskSettings(ctx, data, resp)
	}
}

func (r *ContactResource) validateEmailSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings EmailSettingsModel
	diags := data.EmailSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.Email.IsNull() || settings.Email.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Email Settings",
			"Email address is required for email contacts",
		)
		return
	}

	// Validate email format
	_, err := mail.ParseAddress(settings.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Email Format",
			fmt.Sprintf("Invalid email address format: %s", settings.Email.ValueString()),
		)
	}
}

func (r *ContactResource) validateSmsSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings SmsSettingsModel
	diags := data.SmsSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.Phone.IsNull() || settings.Phone.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid SMS Settings",
			"Phone number is required for SMS contacts",
		)
		return
	}

	// Validate phone number format (basic E.164 validation)
	phone := settings.Phone.ValueString()
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		resp.Diagnostics.AddError(
			"Invalid Phone Number Format",
			fmt.Sprintf("Phone number must be in E.164 format (e.g., +1234567890): %s", phone),
		)
	}
}

func (r *ContactResource) validateWebhookSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings WebhookSettingsModel
	diags := data.WebhookSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.URL.IsNull() || settings.URL.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Webhook Settings",
			"URL is required for webhook contacts",
		)
		return
	}

	// Validate URL format
	webhookURL := settings.URL.ValueString()
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Webhook URL",
			fmt.Sprintf("Invalid URL format: %s", webhookURL),
		)
		return
	}

	// Ensure URL uses HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		resp.Diagnostics.AddError(
			"Invalid Webhook URL Scheme",
			fmt.Sprintf("Webhook URL must use HTTP or HTTPS scheme, got: %s", parsedURL.Scheme),
		)
	}
}

func (r *ContactResource) validateSlackSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings SlackSettingsModel
	diags := data.SlackSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.WebhookURL.IsNull() || settings.WebhookURL.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Slack Settings",
			"Webhook URL is required for Slack contacts",
		)
		return
	}

	// Validate Slack webhook URL format
	webhookURL := settings.WebhookURL.ValueString()
	if !strings.HasPrefix(webhookURL, "https://hooks.slack.com/") {
		resp.Diagnostics.AddError(
			"Invalid Slack Webhook URL",
			fmt.Sprintf("Slack webhook URL must start with 'https://hooks.slack.com/', got: %s", webhookURL),
		)
	}
}

func (r *ContactResource) validateDiscordSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings DiscordSettingsModel
	diags := data.DiscordSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.WebhookURL.IsNull() || settings.WebhookURL.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Discord Settings",
			"Webhook URL is required for Discord contacts",
		)
		return
	}

	// Validate Discord webhook URL format
	webhookURL := settings.WebhookURL.ValueString()
	if !strings.HasPrefix(webhookURL, "https://discord.com/api/webhooks/") &&
		!strings.HasPrefix(webhookURL, "https://discordapp.com/api/webhooks/") {
		resp.Diagnostics.AddError(
			"Invalid Discord Webhook URL",
			fmt.Sprintf("Discord webhook URL must start with 'https://discord.com/api/webhooks/' or 'https://discordapp.com/api/webhooks/', got: %s", webhookURL),
		)
	}
}

func (r *ContactResource) validatePagerdutySettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings PagerdutySettingsModel
	diags := data.PagerdutySettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.IntegrationKey.IsNull() || settings.IntegrationKey.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid PagerDuty Settings",
			"Integration key is required for PagerDuty contacts",
		)
		return
	}

	// Validate integration key format (should be 32 characters)
	integrationKey := settings.IntegrationKey.ValueString()
	if len(integrationKey) != 32 {
		resp.Diagnostics.AddError(
			"Invalid PagerDuty Integration Key",
			fmt.Sprintf("PagerDuty integration key must be exactly 32 characters, got %d characters", len(integrationKey)),
		)
	}

	// Validate severity mapping if provided
	if !settings.SeverityMapping.IsNull() {
		var mapping SeverityMappingModel
		settings.SeverityMapping.As(ctx, &mapping, basetypes.ObjectAsOptions{})

		validSeverities := map[string]bool{
			"critical": true,
			"error":    true,
			"warning":  true,
			"info":     true,
		}

		// Check each severity level if provided
		r.validateSeverity(mapping.Critical, "critical", validSeverities, resp)
		r.validateSeverity(mapping.High, "high", validSeverities, resp)
		r.validateSeverity(mapping.Medium, "medium", validSeverities, resp)
		r.validateSeverity(mapping.Low, "low", validSeverities, resp)
	}
}

func (r *ContactResource) validateSeverity(severity types.String, level string, validSeverities map[string]bool, resp *resource.ModifyPlanResponse) {
	if !severity.IsNull() && severity.ValueString() != "" {
		if !validSeverities[severity.ValueString()] {
			resp.Diagnostics.AddError(
				"Invalid PagerDuty Severity",
				fmt.Sprintf("Invalid severity '%s' for %s priority. Must be one of: critical, error, warning, info", severity.ValueString(), level),
			)
		}
	}
}

func (r *ContactResource) validateIncidentioSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings IncidentioSettingsModel
	diags := data.IncidentioSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.WebhookURL.IsNull() || settings.WebhookURL.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Incident.io Settings",
			"Webhook URL is required for Incident.io contacts",
		)
		return
	}

	if settings.BearerToken.IsNull() || settings.BearerToken.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Incident.io Settings",
			"Bearer token is required for Incident.io contacts",
		)
		return
	}

	// Validate webhook URL format
	webhookURL := settings.WebhookURL.ValueString()
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Incident.io Webhook URL",
			fmt.Sprintf("Invalid URL format: %s", webhookURL),
		)
		return
	}

	if parsedURL.Scheme != "https" {
		resp.Diagnostics.AddError(
			"Invalid Incident.io Webhook URL",
			"Incident.io webhook URL must use HTTPS",
		)
	}
}

func (r *ContactResource) validateOpsgenieSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings OpsgenieSettingsModel
	diags := data.OpsgenieSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.APIKey.IsNull() || settings.APIKey.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Opsgenie Settings",
			"API key is required for Opsgenie contacts",
		)
		return
	}

	// Validate priority if provided
	if !settings.Priority.IsNull() && settings.Priority.ValueString() != "" {
		validPriorities := map[string]bool{
			"P1": true, "P2": true, "P3": true, "P4": true, "P5": true,
		}
		if !validPriorities[settings.Priority.ValueString()] {
			resp.Diagnostics.AddError(
				"Invalid Opsgenie Priority",
				fmt.Sprintf("Priority must be one of: P1, P2, P3, P4, P5. Got: %s", settings.Priority.ValueString()),
			)
		}
	}

	// Validate responders if provided
	if !settings.Responders.IsNull() {
		var responders []OpsgenieResponderModel
		diags := settings.Responders.ElementsAs(ctx, &responders, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		for i, responder := range responders {
			r.validateOpsgenieResponder(ctx, responder, i, resp)
		}
	}
}

func (r *ContactResource) validateOpsgenieResponder(ctx context.Context, responder OpsgenieResponderModel, index int, resp *resource.ModifyPlanResponse) {
	if responder.Type.IsNull() || responder.Type.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Opsgenie Responder",
			fmt.Sprintf("Responder at index %d must have a type", index),
		)
		return
	}

	validTypes := map[string]bool{
		"team": true, "user": true, "escalation": true, "schedule": true,
	}
	if !validTypes[responder.Type.ValueString()] {
		resp.Diagnostics.AddError(
			"Invalid Opsgenie Responder Type",
			fmt.Sprintf("Responder type at index %d must be one of: team, user, escalation, schedule. Got: %s", index, responder.Type.ValueString()),
		)
		return
	}

	// Ensure at least one identifier is provided
	hasIdentifier := (!responder.ID.IsNull() && responder.ID.ValueString() != "") ||
		(!responder.Name.IsNull() && responder.Name.ValueString() != "") ||
		(!responder.Username.IsNull() && responder.Username.ValueString() != "")

	if !hasIdentifier {
		resp.Diagnostics.AddError(
			"Invalid Opsgenie Responder",
			fmt.Sprintf("Responder at index %d must have at least one of: id, name, or username", index),
		)
	}
}

func (r *ContactResource) validateZendeskSettings(ctx context.Context, data *ContactResourceModel, resp *resource.ModifyPlanResponse) {
	var settings ZendeskSettingsModel
	diags := data.ZendeskSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if settings.Subdomain.IsNull() || settings.Subdomain.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Zendesk Settings",
			"Subdomain is required for Zendesk contacts",
		)
		return
	}

	if settings.Email.IsNull() || settings.Email.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Zendesk Settings",
			"Email is required for Zendesk contacts",
		)
		return
	}

	if settings.APIToken.IsNull() || settings.APIToken.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Zendesk Settings",
			"API token is required for Zendesk contacts",
		)
		return
	}

	// Validate email format
	_, err := mail.ParseAddress(settings.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Zendesk Email",
			fmt.Sprintf("Invalid email address format: %s", settings.Email.ValueString()),
		)
	}

	// Validate subdomain format (alphanumeric and hyphens only)
	subdomain := settings.Subdomain.ValueString()
	subdomainRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !subdomainRegex.MatchString(subdomain) {
		resp.Diagnostics.AddError(
			"Invalid Zendesk Subdomain",
			fmt.Sprintf("Subdomain must contain only lowercase letters, numbers, and hyphens: %s", subdomain),
		)
	}

	// Validate priority if provided
	if !settings.Priority.IsNull() && settings.Priority.ValueString() != "" {
		validPriorities := map[string]bool{
			"low": true, "normal": true, "high": true, "urgent": true,
		}
		if !validPriorities[settings.Priority.ValueString()] {
			resp.Diagnostics.AddError(
				"Invalid Zendesk Priority",
				fmt.Sprintf("Priority must be one of: low, normal, high, urgent. Got: %s", settings.Priority.ValueString()),
			)
		}
	}
}
