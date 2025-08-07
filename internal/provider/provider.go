package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-uptime/internal/client"
	"terraform-provider-uptime/internal/datasources"
	"terraform-provider-uptime/internal/resources"
)

// Ensure UptimeProvider satisfies various provider interfaces.
var _ provider.Provider = &UptimeProvider{}
var _ provider.ProviderWithFunctions = &UptimeProvider{}

// UptimeProvider defines the provider implementation.
type UptimeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// UptimeProviderModel describes the provider data model.
type UptimeProviderModel struct {
	ApiKey  types.String `tfsdk:"api_key"`
	BaseUrl types.String `tfsdk:"base_url"`
}

func (p *UptimeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "uptime"
	resp.Version = p.version
}

func (p *UptimeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for authenticating with the Uptime Monitor service. Can also be set via the UPTIME_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for the Uptime Monitor API. Defaults to production API. Can also be set via the UPTIME_BASE_URL environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *UptimeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UptimeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but let explicit configuration override
	apiKey := data.ApiKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("UPTIME_API_KEY")
	}

	baseUrl := data.BaseUrl.ValueString()
	if baseUrl == "" {
		baseUrl = os.Getenv("UPTIME_BASE_URL")
		if baseUrl == "" {
			baseUrl = "https://api.uptime-monitor.io" // Default to production API
		}
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"While configuring the provider, the API key was not found in "+
				"the configuration or UPTIME_API_KEY environment variable. "+
				"This is required for the provider to authenticate with the Uptime Monitor API.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API client and make it available during DataSource and Resource
	// type Configure methods.
	client := client.NewClient(baseUrl, apiKey)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *UptimeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewMonitorResource,
		resources.NewContactResource,
		resources.NewStatusPageResource,
	}
}

func (p *UptimeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewMonitorDataSource,
		datasources.NewAccountDataSource,
		datasources.NewStatusPageDataSource,
	}
}

func (p *UptimeProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// No custom functions for now
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UptimeProvider{
			version: version,
		}
	}
}
