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
var _ datasource.DataSource = &MonitorDataSource{}

func NewMonitorDataSource() datasource.DataSource {
	return &MonitorDataSource{}
}

// MonitorDataSource defines the data source implementation.
type MonitorDataSource struct {
	client *client.Client
}

// MonitorDataSourceModel describes the data source data model.
type MonitorDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`
	Type          types.String `tfsdk:"type"`
	CheckInterval types.Int64  `tfsdk:"check_interval"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	Regions       types.List   `tfsdk:"regions"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (d *MonitorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (d *MonitorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Monitor data source for reading existing monitor configurations.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Monitor identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name of the monitor",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL or endpoint being monitored",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Monitor type: https, tcp, or ping",
				Computed:            true,
			},
			"check_interval": schema.Int64Attribute{
				MarkdownDescription: "Check interval in seconds",
				Computed:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds",
				Computed:            true,
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "List of regions performing checks",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the monitor was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "When the monitor was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *MonitorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *MonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get monitor from API
	monitor, err := d.client.GetMonitor(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read monitor: %s", err))
		return
	}

	// If monitor is not found, return error
	if monitor == nil {
		resp.Diagnostics.AddError("Monitor Not Found", fmt.Sprintf("Monitor with ID %s was not found", data.ID.ValueString()))
		return
	}

	// Map response body to model
	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.Name)
	data.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	data.Timeout = types.Int64Value(int64(monitor.Timeout))

	// Determine type and URL from settings
	if monitor.Settings.HTTPS != nil {
		data.Type = types.StringValue("https")
		data.URL = types.StringValue(monitor.Settings.HTTPS.URL)
	} else if monitor.Settings.TCP != nil {
		data.Type = types.StringValue("tcp")
		data.URL = types.StringValue(monitor.Settings.TCP.URL)
	} else if monitor.Settings.Ping != nil {
		data.Type = types.StringValue("ping")
		data.URL = types.StringValue(monitor.Settings.Ping.URL)
	}

	// Handle regions
	if len(monitor.Regions) > 0 {
		regions := make([]types.String, len(monitor.Regions))
		for i, region := range monitor.Regions {
			regions[i] = types.StringValue(region)
		}
		regionsList, _ := types.ListValueFrom(ctx, types.StringType, regions)
		data.Regions = regionsList
	} else {
		data.Regions = types.ListNull(types.StringType)
	}

	// Handle timestamps
	if monitor.CreatedAt > 0 {
		data.CreatedAt = types.StringValue(fmt.Sprintf("%d", monitor.CreatedAt))
	} else {
		data.CreatedAt = types.StringNull()
	}

	if monitor.UpdatedAt > 0 {
		data.UpdatedAt = types.StringValue(fmt.Sprintf("%d", monitor.UpdatedAt))
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
