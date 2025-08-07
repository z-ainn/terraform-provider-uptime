package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-uptime/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StatusPageDataSource{}

func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

// StatusPageDataSource defines the data source implementation.
type StatusPageDataSource struct {
	client *client.Client
}

// StatusPageDataSourceModel describes the data source data model.
type StatusPageDataSourceModel struct {
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

func (d *StatusPageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (d *StatusPageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Status page data source for reading existing status pages",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the status page to read",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Display name for the status page",
			},
			"monitors": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of monitor IDs displayed on the status page",
			},
			"period": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Time period in days for uptime statistics",
			},
			"custom_domain": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Custom domain for accessing the status page",
			},
			"show_incident_reasons": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether incident reasons are shown publicly",
			},
			"basic_auth": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Basic authentication credentials",
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

func (d *StatusPageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider is not configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StatusPageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get status page from API
	statusPage, err := d.client.GetStatusPage(data.ID.ValueString())
	if err != nil {
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
