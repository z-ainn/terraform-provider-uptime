package resources

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-uptime/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StatusPageResource{}
var _ resource.ResourceWithImportState = &StatusPageResource{}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

// StatusPageResource defines the resource implementation.
type StatusPageResource struct {
	client *client.Client
}

// StatusPageResourceModel describes the resource data model.
type StatusPageResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Monitors            types.List   `tfsdk:"monitors"`
	Period              types.Int64  `tfsdk:"period"`
	CustomDomain        types.String `tfsdk:"custom_domain"`
	ShowIncidentReasons types.Bool   `tfsdk:"show_incident_reasons"`
	BasicAuth           types.String `tfsdk:"basic_auth"`
	CreatedAt           types.Int64  `tfsdk:"created_at"`
	URL                 types.String `tfsdk:"url"`
}

func (r *StatusPageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *StatusPageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Status page resource for displaying monitor statuses publicly",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the status page",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the status page",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"monitors": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of monitor IDs to display on the status page (1-20 monitors)",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 20),
				},
			},
			"period": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(7),
				MarkdownDescription: "Time period in days for uptime statistics (7, 30, or 90)",
				Validators: []validator.Int64{
					int64validator.OneOf(7, 30, 90),
				},
			},
			"custom_domain": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom domain for accessing the status page (e.g., status.example.com)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?)*$`),
						"must be a valid domain name",
					),
				},
			},
			"show_incident_reasons": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to show incident reasons publicly on the status page",
			},
			"basic_auth": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Basic authentication credentials in 'username:password' format",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^:]+:[^:]+$`),
						"must be in 'username:password' format",
					),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp when the status page was created",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL where the status page can be accessed",
			},
		},
	}
}

func (r *StatusPageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider is not configured.
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

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StatusPageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate custom domain
	if !data.CustomDomain.IsNull() && !data.CustomDomain.IsUnknown() {
		domain := data.CustomDomain.ValueString()
		if err := validateCustomDomain(domain); err != nil {
			resp.Diagnostics.AddError("Invalid custom domain", err.Error())
			return
		}
	}

	// Convert monitors list to string slice
	monitors := make([]string, 0)
	resp.Diagnostics.Append(data.Monitors.ElementsAs(ctx, &monitors, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request
	createReq := client.CreateStatusPageRequest{
		Name:     data.Name.ValueString(),
		Monitors: monitors,
	}

	if !data.Period.IsNull() && !data.Period.IsUnknown() {
		period := int(data.Period.ValueInt64())
		createReq.Period = &period
	}

	if !data.CustomDomain.IsNull() && !data.CustomDomain.IsUnknown() {
		domain := data.CustomDomain.ValueString()
		createReq.CustomDomain = &domain
	}

	if !data.ShowIncidentReasons.IsNull() && !data.ShowIncidentReasons.IsUnknown() {
		show := data.ShowIncidentReasons.ValueBool()
		createReq.ShowIncidentReasons = &show
	}

	if !data.BasicAuth.IsNull() && !data.BasicAuth.IsUnknown() {
		auth := data.BasicAuth.ValueString()
		createReq.BasicAuth = &auth
	}

	// Create status page via API
	statusPage, err := r.client.CreateStatusPage(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create status page, got error: %s", err))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(statusPage.ID)
	data.CreatedAt = types.Int64Value(statusPage.CreatedAt)
	data.URL = types.StringValue(statusPage.URL)
	data.Period = types.Int64Value(int64(statusPage.Period))
	data.ShowIncidentReasons = types.BoolValue(statusPage.ShowIncidentReasons)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusPageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get status page from API
	statusPage, err := r.client.GetStatusPage(data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read status page, got error: %s", err))
		return
	}

	// Update model with API data
	data.Name = types.StringValue(statusPage.Name)
	data.Period = types.Int64Value(int64(statusPage.Period))
	data.ShowIncidentReasons = types.BoolValue(statusPage.ShowIncidentReasons)
	data.CreatedAt = types.Int64Value(statusPage.CreatedAt)
	data.URL = types.StringValue(statusPage.URL)

	// Convert monitors to list
	monitorList, diags := types.ListValueFrom(ctx, types.StringType, statusPage.Monitors)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Monitors = monitorList

	// Handle optional fields
	if statusPage.CustomDomain != nil {
		data.CustomDomain = types.StringValue(*statusPage.CustomDomain)
	} else {
		data.CustomDomain = types.StringNull()
	}

	if statusPage.BasicAuth != nil {
		data.BasicAuth = types.StringValue(*statusPage.BasicAuth)
	} else {
		data.BasicAuth = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StatusPageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate custom domain
	if !data.CustomDomain.IsNull() && !data.CustomDomain.IsUnknown() {
		domain := data.CustomDomain.ValueString()
		if err := validateCustomDomain(domain); err != nil {
			resp.Diagnostics.AddError("Invalid custom domain", err.Error())
			return
		}
	}

	// Convert monitors list to string slice
	monitors := make([]string, 0)
	resp.Diagnostics.Append(data.Monitors.ElementsAs(ctx, &monitors, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create update request
	name := data.Name.ValueString()
	updateReq := client.UpdateStatusPageRequest{
		Name:     &name,
		Monitors: monitors,
	}

	if !data.Period.IsNull() && !data.Period.IsUnknown() {
		period := int(data.Period.ValueInt64())
		updateReq.Period = &period
	}

	if !data.CustomDomain.IsNull() && !data.CustomDomain.IsUnknown() {
		domain := data.CustomDomain.ValueString()
		updateReq.CustomDomain = &domain
	} else {
		// Explicitly set to empty string to remove custom domain
		empty := ""
		updateReq.CustomDomain = &empty
	}

	if !data.ShowIncidentReasons.IsNull() && !data.ShowIncidentReasons.IsUnknown() {
		show := data.ShowIncidentReasons.ValueBool()
		updateReq.ShowIncidentReasons = &show
	}

	if !data.BasicAuth.IsNull() && !data.BasicAuth.IsUnknown() {
		auth := data.BasicAuth.ValueString()
		updateReq.BasicAuth = &auth
	} else {
		// Explicitly set to empty string to remove basic auth
		empty := ""
		updateReq.BasicAuth = &empty
	}

	// Update status page via API
	statusPage, err := r.client.UpdateStatusPage(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update status page, got error: %s", err))
		return
	}

	// Update model with response data
	data.URL = types.StringValue(statusPage.URL)
	data.Period = types.Int64Value(int64(statusPage.Period))
	data.ShowIncidentReasons = types.BoolValue(statusPage.ShowIncidentReasons)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StatusPageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete status page via API
	err := r.client.DeleteStatusPage(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete status page, got error: %s", err))
		return
	}
}

func (r *StatusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateCustomDomain validates the custom domain according to backend rules
func validateCustomDomain(domain string) error {
	// Cannot end with uptime-monitor.io
	if strings.HasSuffix(domain, "uptime-monitor.io") {
		return fmt.Errorf("custom domain cannot end with 'uptime-monitor.io'")
	}

	// Must contain at least one dot (no TLDs)
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("custom domain must contain at least one dot")
	}

	// Cannot contain forward slashes
	if strings.Contains(domain, "/") {
		return fmt.Errorf("custom domain cannot contain forward slashes")
	}

	return nil
}
