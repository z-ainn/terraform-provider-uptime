package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-uptime/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ContactResource{}
var _ resource.ResourceWithImportState = &ContactResource{}

func NewContactResource() resource.Resource {
	return &ContactResource{}
}

// ContactResource defines the resource implementation.
type ContactResource struct {
	client *client.Client
}

// ContactResourceModel describes the resource data model.
type ContactResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Channel            types.String `tfsdk:"channel"`
	Active             types.Bool   `tfsdk:"active"`
	DownAlertsOnly     types.Bool   `tfsdk:"down_alerts_only"`
	Error              types.String `tfsdk:"error"`
	EmailSettings      types.Object `tfsdk:"email_settings"`
	SmsSettings        types.Object `tfsdk:"sms_settings"`
	WebhookSettings    types.Object `tfsdk:"webhook_settings"`
	SlackSettings      types.Object `tfsdk:"slack_settings"`
	DiscordSettings    types.Object `tfsdk:"discord_settings"`
	PagerdutySettings  types.Object `tfsdk:"pagerduty_settings"`
	IncidentioSettings types.Object `tfsdk:"incidentio_settings"`
	OpsgenieSettings   types.Object `tfsdk:"opsgenie_settings"`
	ZendeskSettings    types.Object `tfsdk:"zendesk_settings"`
}

// Settings models for each channel type
type EmailSettingsModel struct {
	Email types.String `tfsdk:"email"`
}

type SmsSettingsModel struct {
	Phone types.String `tfsdk:"phone"`
}

type WebhookSettingsModel struct {
	URL types.String `tfsdk:"url"`
}

type SlackSettingsModel struct {
	WebhookURL types.String `tfsdk:"webhook_url"`
}

type DiscordSettingsModel struct {
	WebhookURL types.String `tfsdk:"webhook_url"`
}

type PagerdutySettingsModel struct {
	IntegrationKey       types.String `tfsdk:"integration_key"`
	SeverityMapping      types.Object `tfsdk:"severity_mapping"`
	AutoResolveIncidents types.Bool   `tfsdk:"auto_resolve_incidents"`
}

type SeverityMappingModel struct {
	Critical types.String `tfsdk:"critical"`
	High     types.String `tfsdk:"high"`
	Medium   types.String `tfsdk:"medium"`
	Low      types.String `tfsdk:"low"`
}

type IncidentioSettingsModel struct {
	WebhookURL           types.String `tfsdk:"webhook_url"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	AutoResolveIncidents types.Bool   `tfsdk:"auto_resolve_incidents"`
}

type OpsgenieSettingsModel struct {
	APIKey          types.String `tfsdk:"api_key"`
	Priority        types.String `tfsdk:"priority"`
	Responders      types.List   `tfsdk:"responders"`
	Tags            types.List   `tfsdk:"tags"`
	AutoCloseAlerts types.Bool   `tfsdk:"auto_close_alerts"`
	EUInstance      types.Bool   `tfsdk:"eu_instance"`
}

type OpsgenieResponderModel struct {
	Type     types.String `tfsdk:"type"`
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Username types.String `tfsdk:"username"`
}

type ZendeskSettingsModel struct {
	Subdomain        types.String `tfsdk:"subdomain"`
	Email            types.String `tfsdk:"email"`
	APIToken         types.String `tfsdk:"api_token"`
	Priority         types.String `tfsdk:"priority"`
	CustomFields     types.List   `tfsdk:"custom_fields"`
	Tags             types.List   `tfsdk:"tags"`
	AutoSolveTickets types.Bool   `tfsdk:"auto_solve_tickets"`
}

type ZendeskCustomFieldModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
}

func (r *ContactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (r *ContactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Contact for monitor notifications. Supports multiple channel types including email, SMS, webhooks, and various third-party integrations.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Contact identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the contact",
				Required:            true,
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "Contact channel type (email, sms, webhook, slack, discord, pagerduty, incidentio, opsgenie, zendesk)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("email", "sms", "webhook", "slack", "discord", "pagerduty", "incidentio", "opsgenie", "zendesk"),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether this contact is active (system-managed based on delivery status)",
				Computed:            true,
			},
			"down_alerts_only": schema.BoolAttribute{
				MarkdownDescription: "Only receive alerts when monitors go down (not up)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"error": schema.StringAttribute{
				MarkdownDescription: "Error message if contact failed (e.g., email bounce)",
				Computed:            true,
			},
			// Email settings
			"email_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Email channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"email": schema.StringAttribute{
						MarkdownDescription: "Email address",
						Required:            true,
					},
				},
			},
			// SMS settings
			"sms_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "SMS channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"phone": schema.StringAttribute{
						MarkdownDescription: "Phone number",
						Required:            true,
					},
				},
			},
			// Webhook settings
			"webhook_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Webhook channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Webhook URL (must use HTTP or HTTPS)",
						Required:            true,
					},
				},
			},
			// Slack settings
			"slack_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Slack channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"webhook_url": schema.StringAttribute{
						MarkdownDescription: "Slack webhook URL (must start with https://hooks.slack.com/)",
						Required:            true,
					},
				},
			},
			// Discord settings
			"discord_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Discord channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"webhook_url": schema.StringAttribute{
						MarkdownDescription: "Discord webhook URL (must start with https://discord.com/api/webhooks/)",
						Required:            true,
					},
				},
			},
			// PagerDuty settings
			"pagerduty_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "PagerDuty channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"integration_key": schema.StringAttribute{
						MarkdownDescription: "PagerDuty integration key (32 characters)",
						Required:            true,
						Sensitive:           true,
					},
					"auto_resolve_incidents": schema.BoolAttribute{
						MarkdownDescription: "Automatically resolve incidents when monitor recovers",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"severity_mapping": schema.SingleNestedAttribute{
						MarkdownDescription: "Map monitor priority levels to PagerDuty severities",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"critical": schema.StringAttribute{
								MarkdownDescription: "Severity for critical priority (critical, error, warning, info)",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("critical", "error", "warning", "info"),
								},
							},
							"high": schema.StringAttribute{
								MarkdownDescription: "Severity for high priority (critical, error, warning, info)",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("critical", "error", "warning", "info"),
								},
							},
							"medium": schema.StringAttribute{
								MarkdownDescription: "Severity for medium priority (critical, error, warning, info)",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("critical", "error", "warning", "info"),
								},
							},
							"low": schema.StringAttribute{
								MarkdownDescription: "Severity for low priority (critical, error, warning, info)",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.OneOf("critical", "error", "warning", "info"),
								},
							},
						},
					},
				},
			},
			// IncidentIo settings
			"incidentio_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Incident.io channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"webhook_url": schema.StringAttribute{
						MarkdownDescription: "Incident.io webhook URL",
						Required:            true,
					},
					"bearer_token": schema.StringAttribute{
						MarkdownDescription: "Bearer token for authentication",
						Required:            true,
						Sensitive:           true,
					},
					"auto_resolve_incidents": schema.BoolAttribute{
						MarkdownDescription: "Automatically resolve incidents when monitor recovers",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			},
			// Opsgenie settings
			"opsgenie_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Opsgenie channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						MarkdownDescription: "Opsgenie API key",
						Required:            true,
						Sensitive:           true,
					},
					"priority": schema.StringAttribute{
						MarkdownDescription: "Alert priority (P1, P2, P3, P4, P5)",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("P1", "P2", "P3", "P4", "P5"),
						},
					},
					"responders": schema.ListNestedAttribute{
						MarkdownDescription: "List of responders to notify",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Responder type (team, user, escalation, schedule)",
									Required:            true,
									Validators: []validator.String{
										stringvalidator.OneOf("team", "user", "escalation", "schedule"),
									},
								},
								"id": schema.StringAttribute{
									MarkdownDescription: "Responder ID",
									Optional:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Responder name",
									Optional:            true,
								},
								"username": schema.StringAttribute{
									MarkdownDescription: "Responder username (for user type)",
									Optional:            true,
								},
							},
						},
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: "Tags to add to alerts",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"auto_close_alerts": schema.BoolAttribute{
						MarkdownDescription: "Automatically close alerts when monitor recovers",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"eu_instance": schema.BoolAttribute{
						MarkdownDescription: "Use EU instance of Opsgenie",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			// Zendesk settings
			"zendesk_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Zendesk channel configuration",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"subdomain": schema.StringAttribute{
						MarkdownDescription: "Zendesk subdomain",
						Required:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Zendesk account email",
						Required:            true,
					},
					"api_token": schema.StringAttribute{
						MarkdownDescription: "Zendesk API token",
						Required:            true,
						Sensitive:           true,
					},
					"priority": schema.StringAttribute{
						MarkdownDescription: "Ticket priority (low, normal, high, urgent)",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("low", "normal", "high", "urgent"),
						},
					},
					"custom_fields": schema.ListNestedAttribute{
						MarkdownDescription: "Custom fields to set on tickets",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									MarkdownDescription: "Custom field ID",
									Required:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Custom field value",
									Required:            true,
								},
							},
						},
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: "Tags to add to tickets",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"auto_solve_tickets": schema.BoolAttribute{
						MarkdownDescription: "Automatically solve tickets when monitor recovers",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			},
		},
	}
}

func (r *ContactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ContactResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip validation during resource destruction
	if req.Plan.Raw.IsNull() {
		return
	}

	var data ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform comprehensive validation
	r.validateContactRequest(ctx, &data, resp)
}

func (r *ContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContactResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the details JSON based on channel type
	details, err := r.buildDetailsJSON(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to build contact details: %s", err))
		return
	}

	// Create contact via API (always active initially)
	contact, err := r.client.CreateContact(&client.CreateContactRequest{
		Name:           data.Name.ValueString(),
		Channel:        data.Channel.ValueString(),
		Details:        details,
		Active:         true, // New contacts always start as active
		DownAlertsOnly: data.DownAlertsOnly.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create contact: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(contact.ID)
	data.Active = types.BoolValue(contact.Active)
	if contact.Error != nil {
		data.Error = types.StringValue(*contact.Error)
	} else {
		data.Error = types.StringNull()
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get contact from API
	contact, err := r.client.GetContact(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read contact: %s", err))
		return
	}

	// If contact is not found, remove from state
	if contact == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with API data
	data.Name = types.StringValue(contact.Name)
	data.Channel = types.StringValue(contact.Channel)
	data.Active = types.BoolValue(contact.Active)
	data.DownAlertsOnly = types.BoolValue(contact.DownAlertsOnly)

	if contact.Error != nil {
		data.Error = types.StringValue(*contact.Error)
	} else {
		data.Error = types.StringNull()
	}

	// Parse details JSON back into appropriate settings
	err = r.parseDetailsJSON(ctx, contact.Channel, contact.Details, &data)
	if err != nil {
		resp.Diagnostics.AddError("Data Conversion Error", fmt.Sprintf("Unable to parse contact details: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContactResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := &client.UpdateContactRequest{}

	// Only update fields that have changed
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if name changed
	if !data.Name.Equal(state.Name) {
		name := data.Name.ValueString()
		updateReq.Name = &name
	}

	// Note: active field is system-managed and should not be updated

	// Check if down_alerts_only changed
	if !data.DownAlertsOnly.Equal(state.DownAlertsOnly) {
		downAlertsOnly := data.DownAlertsOnly.ValueBool()
		updateReq.DownAlertsOnly = &downAlertsOnly
	}

	// Check if channel-specific settings changed
	if !r.settingsEqual(ctx, &data, &state) {
		details, err := r.buildDetailsJSON(ctx, &data)
		if err != nil {
			resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to build contact details: %s", err))
			return
		}
		updateReq.Details = details
	}

	// Update contact via API
	contact, err := r.client.UpdateContact(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update contact: %s", err))
		return
	}

	// Update the model with API response
	data.Name = types.StringValue(contact.Name)
	data.Channel = types.StringValue(contact.Channel)
	data.Active = types.BoolValue(contact.Active)
	data.DownAlertsOnly = types.BoolValue(contact.DownAlertsOnly)

	if contact.Error != nil {
		data.Error = types.StringValue(*contact.Error)
	} else {
		data.Error = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete contact via API
	err := r.client.DeleteContact(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete contact: %s", err))
		return
	}
}

func (r *ContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

func (r *ContactResource) buildDetailsJSON(ctx context.Context, data *ContactResourceModel) (json.RawMessage, error) {
	channel := data.Channel.ValueString()
	var details map[string]interface{}

	switch channel {
	case "email":
		if data.EmailSettings.IsNull() {
			return nil, fmt.Errorf("email_settings is required for email channel")
		}
		var settings EmailSettingsModel
		data.EmailSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"email": settings.Email.ValueString(),
		}

	case "sms":
		if data.SmsSettings.IsNull() {
			return nil, fmt.Errorf("sms_settings is required for sms channel")
		}
		var settings SmsSettingsModel
		data.SmsSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"phone": settings.Phone.ValueString(),
		}

	case "webhook":
		if data.WebhookSettings.IsNull() {
			return nil, fmt.Errorf("webhook_settings is required for webhook channel")
		}
		var settings WebhookSettingsModel
		data.WebhookSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"url": settings.URL.ValueString(),
		}

	case "slack":
		if data.SlackSettings.IsNull() {
			return nil, fmt.Errorf("slack_settings is required for slack channel")
		}
		var settings SlackSettingsModel
		data.SlackSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"webhook_url": settings.WebhookURL.ValueString(),
		}

	case "discord":
		if data.DiscordSettings.IsNull() {
			return nil, fmt.Errorf("discord_settings is required for discord channel")
		}
		var settings DiscordSettingsModel
		data.DiscordSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"webhook_url": settings.WebhookURL.ValueString(),
		}

	case "pagerduty":
		if data.PagerdutySettings.IsNull() {
			return nil, fmt.Errorf("pagerduty_settings is required for pagerduty channel")
		}
		var settings PagerdutySettingsModel
		data.PagerdutySettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"integration_key":        settings.IntegrationKey.ValueString(),
			"auto_resolve_incidents": settings.AutoResolveIncidents.ValueBool(),
		}

		// Add severity mapping if provided
		if !settings.SeverityMapping.IsNull() {
			var mapping SeverityMappingModel
			settings.SeverityMapping.As(ctx, &mapping, basetypes.ObjectAsOptions{})
			severityMap := map[string]string{}
			if !mapping.Critical.IsNull() {
				severityMap["critical"] = mapping.Critical.ValueString()
			}
			if !mapping.High.IsNull() {
				severityMap["high"] = mapping.High.ValueString()
			}
			if !mapping.Medium.IsNull() {
				severityMap["medium"] = mapping.Medium.ValueString()
			}
			if !mapping.Low.IsNull() {
				severityMap["low"] = mapping.Low.ValueString()
			}
			if len(severityMap) > 0 {
				details["severity_mapping"] = severityMap
			}
		}

	case "incidentio":
		if data.IncidentioSettings.IsNull() {
			return nil, fmt.Errorf("incidentio_settings is required for incidentio channel")
		}
		var settings IncidentioSettingsModel
		data.IncidentioSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"webhook_url":            settings.WebhookURL.ValueString(),
			"bearer_token":           settings.BearerToken.ValueString(),
			"auto_resolve_incidents": settings.AutoResolveIncidents.ValueBool(),
		}

	case "opsgenie":
		if data.OpsgenieSettings.IsNull() {
			return nil, fmt.Errorf("opsgenie_settings is required for opsgenie channel")
		}
		var settings OpsgenieSettingsModel
		data.OpsgenieSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"api_key":           settings.APIKey.ValueString(),
			"auto_close_alerts": settings.AutoCloseAlerts.ValueBool(),
			"eu_instance":       settings.EUInstance.ValueBool(),
		}

		if !settings.Priority.IsNull() {
			details["priority"] = settings.Priority.ValueString()
		}

		if !settings.Responders.IsNull() {
			var responders []OpsgenieResponderModel
			settings.Responders.ElementsAs(ctx, &responders, false)
			var responderList []map[string]interface{}
			for _, r := range responders {
				responder := map[string]interface{}{
					"type": r.Type.ValueString(),
				}
				if !r.ID.IsNull() {
					responder["id"] = r.ID.ValueString()
				}
				if !r.Name.IsNull() {
					responder["name"] = r.Name.ValueString()
				}
				if !r.Username.IsNull() {
					responder["username"] = r.Username.ValueString()
				}
				responderList = append(responderList, responder)
			}
			if len(responderList) > 0 {
				details["responders"] = responderList
			}
		}

		if !settings.Tags.IsNull() {
			var tags []string
			settings.Tags.ElementsAs(ctx, &tags, false)
			if len(tags) > 0 {
				details["tags"] = tags
			}
		}

	case "zendesk":
		if data.ZendeskSettings.IsNull() {
			return nil, fmt.Errorf("zendesk_settings is required for zendesk channel")
		}
		var settings ZendeskSettingsModel
		data.ZendeskSettings.As(ctx, &settings, basetypes.ObjectAsOptions{})
		details = map[string]interface{}{
			"subdomain":          settings.Subdomain.ValueString(),
			"email":              settings.Email.ValueString(),
			"api_token":          settings.APIToken.ValueString(),
			"auto_solve_tickets": settings.AutoSolveTickets.ValueBool(),
		}

		if !settings.Priority.IsNull() {
			details["priority"] = settings.Priority.ValueString()
		}

		if !settings.CustomFields.IsNull() {
			var customFields []ZendeskCustomFieldModel
			settings.CustomFields.ElementsAs(ctx, &customFields, false)
			var fieldList []map[string]interface{}
			for _, f := range customFields {
				fieldList = append(fieldList, map[string]interface{}{
					"id":    f.ID.ValueInt64(),
					"value": f.Value.ValueString(),
				})
			}
			if len(fieldList) > 0 {
				details["custom_fields"] = fieldList
			}
		}

		if !settings.Tags.IsNull() {
			var tags []string
			settings.Tags.ElementsAs(ctx, &tags, false)
			if len(tags) > 0 {
				details["tags"] = tags
			}
		}

	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channel)
	}

	return json.Marshal(details)
}

func (r *ContactResource) parseDetailsJSON(ctx context.Context, channel string, details json.RawMessage, data *ContactResourceModel) error {
	// Clear all settings first
	data.EmailSettings = types.ObjectNull(r.getEmailSettingsAttrs())
	data.SmsSettings = types.ObjectNull(r.getSmsSettingsAttrs())
	data.WebhookSettings = types.ObjectNull(r.getWebhookSettingsAttrs())
	data.SlackSettings = types.ObjectNull(r.getSlackSettingsAttrs())
	data.DiscordSettings = types.ObjectNull(r.getDiscordSettingsAttrs())
	data.PagerdutySettings = types.ObjectNull(r.getPagerdutySettingsAttrs())
	data.IncidentioSettings = types.ObjectNull(r.getIncidentioSettingsAttrs())
	data.OpsgenieSettings = types.ObjectNull(r.getOpsgenieSettingsAttrs())
	data.ZendeskSettings = types.ObjectNull(r.getZendeskSettingsAttrs())

	var detailsMap map[string]interface{}
	if err := json.Unmarshal(details, &detailsMap); err != nil {
		return err
	}

	switch channel {
	case "email":
		settings := EmailSettingsModel{
			Email: types.StringValue(detailsMap["email"].(string)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getEmailSettingsAttrs(), settings)
		data.EmailSettings = obj

	case "sms":
		settings := SmsSettingsModel{
			Phone: types.StringValue(detailsMap["phone"].(string)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getSmsSettingsAttrs(), settings)
		data.SmsSettings = obj

	case "webhook":
		settings := WebhookSettingsModel{
			URL: types.StringValue(detailsMap["url"].(string)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getWebhookSettingsAttrs(), settings)
		data.WebhookSettings = obj

	case "slack":
		settings := SlackSettingsModel{
			WebhookURL: types.StringValue(detailsMap["webhook_url"].(string)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getSlackSettingsAttrs(), settings)
		data.SlackSettings = obj

	case "discord":
		settings := DiscordSettingsModel{
			WebhookURL: types.StringValue(detailsMap["webhook_url"].(string)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getDiscordSettingsAttrs(), settings)
		data.DiscordSettings = obj

	case "pagerduty":
		settings := PagerdutySettingsModel{
			IntegrationKey:       types.StringValue(detailsMap["integration_key"].(string)),
			AutoResolveIncidents: types.BoolValue(detailsMap["auto_resolve_incidents"].(bool)),
		}

		if severityMapping, ok := detailsMap["severity_mapping"].(map[string]interface{}); ok {
			mapping := SeverityMappingModel{
				Critical: types.StringNull(),
				High:     types.StringNull(),
				Medium:   types.StringNull(),
				Low:      types.StringNull(),
			}
			if v, ok := severityMapping["critical"].(string); ok {
				mapping.Critical = types.StringValue(v)
			}
			if v, ok := severityMapping["high"].(string); ok {
				mapping.High = types.StringValue(v)
			}
			if v, ok := severityMapping["medium"].(string); ok {
				mapping.Medium = types.StringValue(v)
			}
			if v, ok := severityMapping["low"].(string); ok {
				mapping.Low = types.StringValue(v)
			}
			mappingObj, _ := types.ObjectValueFrom(ctx, r.getSeverityMappingAttrs(), mapping)
			settings.SeverityMapping = mappingObj
		} else {
			settings.SeverityMapping = types.ObjectNull(r.getSeverityMappingAttrs())
		}

		obj, _ := types.ObjectValueFrom(ctx, r.getPagerdutySettingsAttrs(), settings)
		data.PagerdutySettings = obj

	case "incidentio":
		settings := IncidentioSettingsModel{
			WebhookURL:           types.StringValue(detailsMap["webhook_url"].(string)),
			BearerToken:          types.StringValue(detailsMap["bearer_token"].(string)),
			AutoResolveIncidents: types.BoolValue(detailsMap["auto_resolve_incidents"].(bool)),
		}
		obj, _ := types.ObjectValueFrom(ctx, r.getIncidentioSettingsAttrs(), settings)
		data.IncidentioSettings = obj

	case "opsgenie":
		settings := OpsgenieSettingsModel{
			APIKey:          types.StringValue(detailsMap["api_key"].(string)),
			AutoCloseAlerts: types.BoolValue(detailsMap["auto_close_alerts"].(bool)),
			EUInstance:      types.BoolValue(detailsMap["eu_instance"].(bool)),
		}

		if priority, ok := detailsMap["priority"].(string); ok {
			settings.Priority = types.StringValue(priority)
		} else {
			settings.Priority = types.StringNull()
		}

		if responders, ok := detailsMap["responders"].([]interface{}); ok {
			var responderList []attr.Value
			for _, resp := range responders {
				respMap := resp.(map[string]interface{})
				responder := OpsgenieResponderModel{
					Type:     types.StringValue(respMap["type"].(string)),
					ID:       types.StringNull(),
					Name:     types.StringNull(),
					Username: types.StringNull(),
				}
				if id, ok := respMap["id"].(string); ok {
					responder.ID = types.StringValue(id)
				}
				if name, ok := respMap["name"].(string); ok {
					responder.Name = types.StringValue(name)
				}
				if username, ok := respMap["username"].(string); ok {
					responder.Username = types.StringValue(username)
				}
				respObj, _ := types.ObjectValueFrom(ctx, r.getOpsgenieResponderAttrs(), responder)
				responderList = append(responderList, respObj)
			}
			settings.Responders, _ = types.ListValue(types.ObjectType{AttrTypes: r.getOpsgenieResponderAttrs()}, responderList)
		} else {
			settings.Responders = types.ListNull(types.ObjectType{AttrTypes: r.getOpsgenieResponderAttrs()})
		}

		if tags, ok := detailsMap["tags"].([]interface{}); ok {
			var tagList []attr.Value
			for _, t := range tags {
				tagList = append(tagList, types.StringValue(t.(string)))
			}
			settings.Tags, _ = types.ListValue(types.StringType, tagList)
		} else {
			settings.Tags = types.ListNull(types.StringType)
		}

		obj, _ := types.ObjectValueFrom(ctx, r.getOpsgenieSettingsAttrs(), settings)
		data.OpsgenieSettings = obj

	case "zendesk":
		settings := ZendeskSettingsModel{
			Subdomain:        types.StringValue(detailsMap["subdomain"].(string)),
			Email:            types.StringValue(detailsMap["email"].(string)),
			APIToken:         types.StringValue(detailsMap["api_token"].(string)),
			AutoSolveTickets: types.BoolValue(detailsMap["auto_solve_tickets"].(bool)),
		}

		if priority, ok := detailsMap["priority"].(string); ok {
			settings.Priority = types.StringValue(priority)
		} else {
			settings.Priority = types.StringNull()
		}

		if customFields, ok := detailsMap["custom_fields"].([]interface{}); ok {
			var fieldList []attr.Value
			for _, f := range customFields {
				fieldMap := f.(map[string]interface{})
				field := ZendeskCustomFieldModel{
					ID:    types.Int64Value(int64(fieldMap["id"].(float64))),
					Value: types.StringValue(fieldMap["value"].(string)),
				}
				fieldObj, _ := types.ObjectValueFrom(ctx, r.getZendeskCustomFieldAttrs(), field)
				fieldList = append(fieldList, fieldObj)
			}
			settings.CustomFields, _ = types.ListValue(types.ObjectType{AttrTypes: r.getZendeskCustomFieldAttrs()}, fieldList)
		} else {
			settings.CustomFields = types.ListNull(types.ObjectType{AttrTypes: r.getZendeskCustomFieldAttrs()})
		}

		if tags, ok := detailsMap["tags"].([]interface{}); ok {
			var tagList []attr.Value
			for _, t := range tags {
				tagList = append(tagList, types.StringValue(t.(string)))
			}
			settings.Tags, _ = types.ListValue(types.StringType, tagList)
		} else {
			settings.Tags = types.ListNull(types.StringType)
		}

		obj, _ := types.ObjectValueFrom(ctx, r.getZendeskSettingsAttrs(), settings)
		data.ZendeskSettings = obj
	}

	return nil
}

func (r *ContactResource) settingsEqual(ctx context.Context, data1, data2 *ContactResourceModel) bool {
	// Compare channel-specific settings based on channel type
	channel := data1.Channel.ValueString()

	switch channel {
	case "email":
		return data1.EmailSettings.Equal(data2.EmailSettings)
	case "sms":
		return data1.SmsSettings.Equal(data2.SmsSettings)
	case "webhook":
		return data1.WebhookSettings.Equal(data2.WebhookSettings)
	case "slack":
		return data1.SlackSettings.Equal(data2.SlackSettings)
	case "discord":
		return data1.DiscordSettings.Equal(data2.DiscordSettings)
	case "pagerduty":
		return data1.PagerdutySettings.Equal(data2.PagerdutySettings)
	case "incidentio":
		return data1.IncidentioSettings.Equal(data2.IncidentioSettings)
	case "opsgenie":
		return data1.OpsgenieSettings.Equal(data2.OpsgenieSettings)
	case "zendesk":
		return data1.ZendeskSettings.Equal(data2.ZendeskSettings)
	}

	return true
}

// Attribute type definitions for nested objects
func (r *ContactResource) getEmailSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"email": types.StringType,
	}
}

func (r *ContactResource) getSmsSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"phone": types.StringType,
	}
}

func (r *ContactResource) getWebhookSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

func (r *ContactResource) getSlackSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"webhook_url": types.StringType,
	}
}

func (r *ContactResource) getDiscordSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"webhook_url": types.StringType,
	}
}

func (r *ContactResource) getPagerdutySettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"integration_key":        types.StringType,
		"auto_resolve_incidents": types.BoolType,
		"severity_mapping":       types.ObjectType{AttrTypes: r.getSeverityMappingAttrs()},
	}
}

func (r *ContactResource) getSeverityMappingAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"critical": types.StringType,
		"high":     types.StringType,
		"medium":   types.StringType,
		"low":      types.StringType,
	}
}

func (r *ContactResource) getIncidentioSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"webhook_url":            types.StringType,
		"bearer_token":           types.StringType,
		"auto_resolve_incidents": types.BoolType,
	}
}

func (r *ContactResource) getOpsgenieSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key":           types.StringType,
		"priority":          types.StringType,
		"responders":        types.ListType{ElemType: types.ObjectType{AttrTypes: r.getOpsgenieResponderAttrs()}},
		"tags":              types.ListType{ElemType: types.StringType},
		"auto_close_alerts": types.BoolType,
		"eu_instance":       types.BoolType,
	}
}

func (r *ContactResource) getOpsgenieResponderAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"id":       types.StringType,
		"name":     types.StringType,
		"username": types.StringType,
	}
}

func (r *ContactResource) getZendeskSettingsAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"subdomain":          types.StringType,
		"email":              types.StringType,
		"api_token":          types.StringType,
		"priority":           types.StringType,
		"custom_fields":      types.ListType{ElemType: types.ObjectType{AttrTypes: r.getZendeskCustomFieldAttrs()}},
		"tags":               types.ListType{ElemType: types.StringType},
		"auto_solve_tickets": types.BoolType,
	}
}

func (r *ContactResource) getZendeskCustomFieldAttrs() map[string]attr.Type {
	return map[string]attr.Type{
		"id":    types.Int64Type,
		"value": types.StringType,
	}
}
