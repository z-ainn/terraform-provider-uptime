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
var _ datasource.DataSource = &AccountDataSource{}

func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

// AccountDataSource defines the data source implementation.
type AccountDataSource struct {
	client *client.Client
}

// AccountDataSourceModel describes the data source data model.
type AccountDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	CurrentPlan    types.String `tfsdk:"current_plan"`
	MonitorsLimit  types.Int64  `tfsdk:"monitors_limit"`
	MonitorsCount  types.Int64  `tfsdk:"monitors_count"`
	UpMonitors     types.Int64  `tfsdk:"up_monitors"`
	DownMonitors   types.Int64  `tfsdk:"down_monitors"`
	PausedMonitors types.Int64  `tfsdk:"paused_monitors"`
}

func (d *AccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *AccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Account data source provides information about the current account, including plan limits and monitor statistics.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Account identifier",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Account email address",
				Computed:            true,
			},
			"current_plan": schema.StringAttribute{
				MarkdownDescription: "Current subscription plan ID",
				Computed:            true,
			},
			"monitors_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of monitors allowed by the current plan",
				Computed:            true,
			},
			"monitors_count": schema.Int64Attribute{
				MarkdownDescription: "Total number of monitors currently configured",
				Computed:            true,
			},
			"up_monitors": schema.Int64Attribute{
				MarkdownDescription: "Number of active monitors with UP status",
				Computed:            true,
			},
			"down_monitors": schema.Int64Attribute{
				MarkdownDescription: "Number of active monitors with DOWN status",
				Computed:            true,
			},
			"paused_monitors": schema.Int64Attribute{
				MarkdownDescription: "Number of paused monitors",
				Computed:            true,
			},
		},
	}
}

func (d *AccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get account information from API
	account, err := d.client.GetAccount()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read account, got error: %s", err))
		return
	}

	// Map response to model
	data.ID = types.StringValue(account.ID)
	data.Email = types.StringValue(account.Email)
	data.CurrentPlan = types.StringValue(account.CurrentPlan)
	data.MonitorsLimit = types.Int64Value(int64(account.MonitorsLimit))
	data.MonitorsCount = types.Int64Value(int64(account.MonitorsCount))
	data.UpMonitors = types.Int64Value(int64(account.UpMonitors))
	data.DownMonitors = types.Int64Value(int64(account.DownMonitors))
	data.PausedMonitors = types.Int64Value(int64(account.PausedMonitors))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
