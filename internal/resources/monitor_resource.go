package resources

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-uptime/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MonitorResource{}
var _ resource.ResourceWithImportState = &MonitorResource{}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

// MonitorResource defines the resource implementation.
type MonitorResource struct {
	client *client.Client
}

// MonitorResourceModel describes the resource data model.
type MonitorResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`
	Type          types.String `tfsdk:"type"`
	Active        types.Bool   `tfsdk:"active"`
	CheckInterval types.Int64  `tfsdk:"check_interval"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	FailThreshold types.Int64  `tfsdk:"fail_threshold"`
	Regions       types.List   `tfsdk:"regions"`
	Contacts      types.List   `tfsdk:"contacts"`
	HTTPSSettings types.Object `tfsdk:"https_settings"`
	TCPSettings   types.Object `tfsdk:"tcp_settings"`
	PingSettings  types.Object `tfsdk:"ping_settings"`
	// Certificate monitoring fields
	Host types.String `tfsdk:"host"`
	Port types.Int64  `tfsdk:"port"`
}

// HTTPSSettingsModel represents HTTPS-specific configuration
type HTTPSSettingsModel struct {
	Method                     types.String `tfsdk:"method"`
	ExpectedStatusCodes        types.String `tfsdk:"expected_status_codes"`
	CheckCertificateExpiration types.Bool   `tfsdk:"check_certificate_expiration"`
	FollowRedirects            types.Bool   `tfsdk:"follow_redirects"`
	RequestHeaders             types.Map    `tfsdk:"request_headers"`
	RequestBody                types.String `tfsdk:"request_body"`
	ExpectedResponseBody       types.String `tfsdk:"expected_response_body"`
	ExpectedResponseHeaders    types.Map    `tfsdk:"expected_response_headers"`
}

// TCPSettingsModel represents TCP-specific configuration
type TCPSettingsModel struct {
	// TCP has minimal settings as it just checks connectivity
}

// PingSettingsModel represents Ping-specific configuration
type PingSettingsModel struct {
	// Ping has minimal settings as it just sends ICMP packets
}

func (r *MonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uptime monitor resource for monitoring HTTP/HTTPS, TCP, and Ping endpoints.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Monitor identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the monitor",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL or endpoint to monitor. For ping monitors, you can provide just an IP address or hostname - the provider will automatically add the ping:// scheme.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Monitor type: https, tcp, or ping",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the monitor is active and should perform checks",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"check_interval": schema.Int64Attribute{
				MarkdownDescription: "Check interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(30),
			},
			"fail_threshold": schema.Int64Attribute{
				MarkdownDescription: "Number of consecutive failed checks before marking monitor as down. Must not exceed the number of regions.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "List of regions to perform checks from",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"contacts": schema.ListAttribute{
				MarkdownDescription: "List of contact IDs to notify when monitor status changes",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"https_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "HTTPS-specific configuration (only applicable when type is 'https')",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						MarkdownDescription: "HTTP method to use (HEAD, GET, POST, PUT, etc.)",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("HEAD"),
					},
					"expected_status_codes": schema.StringAttribute{
						MarkdownDescription: "Expected HTTP status codes (e.g., '200', '200-299', '200,201,301')",
						Optional:            true,
					},
					"check_certificate_expiration": schema.BoolAttribute{
						MarkdownDescription: "Whether to check SSL certificate expiration",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"follow_redirects": schema.BoolAttribute{
						MarkdownDescription: "Whether to follow HTTP redirects",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"request_headers": schema.MapAttribute{
						MarkdownDescription: "HTTP headers to send with the request",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"request_body": schema.StringAttribute{
						MarkdownDescription: "HTTP request body (for POST/PUT requests)",
						Optional:            true,
					},
					"expected_response_body": schema.StringAttribute{
						MarkdownDescription: "Expected substring in the response body",
						Optional:            true,
					},
					"expected_response_headers": schema.MapAttribute{
						MarkdownDescription: "Expected HTTP response headers",
						ElementType:         types.StringType,
						Optional:            true,
					},
				},
			},
			"tcp_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "TCP-specific configuration (only applicable when type is 'tcp')",
				Optional:            true,
				Attributes:          map[string]schema.Attribute{
					// TCP monitors use the URL field for host:port, no additional settings needed
				},
			},
			"ping_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Ping-specific configuration (only applicable when type is 'ping')",
				Optional:            true,
				Attributes:          map[string]schema.Attribute{
					// Ping monitors use the URL field for hostname, no additional settings needed
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Host for certificate expiration monitoring (extracted from URL)",
				Optional:            true,
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port for certificate expiration monitoring (extracted from URL)",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *MonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API request
	createReq, err := r.modelToCreateRequest(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to convert configuration: %s", err))
		return
	}

	// Create monitor via API
	monitor, err := r.client.CreateMonitor(*createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create monitor: %s", err))
		return
	}

	// Convert API response back to Terraform model
	err = r.apiModelToTerraformModel(ctx, monitor, &data)
	if err != nil {
		resp.Diagnostics.AddError("Data Conversion Error", fmt.Sprintf("Unable to convert API response: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get monitor from API
	monitor, err := r.client.GetMonitor(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read monitor: %s", err))
		return
	}

	// If monitor is not found, remove from state
	if monitor == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert API response to Terraform model
	err = r.apiModelToTerraformModel(ctx, monitor, &data)
	if err != nil {
		resp.Diagnostics.AddError("Data Conversion Error", fmt.Sprintf("Unable to convert API response: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API request
	updateReq, err := r.modelToUpdateRequest(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to convert configuration: %s", err))
		return
	}

	// Update monitor via API
	monitor, err := r.client.UpdateMonitor(data.ID.ValueString(), *updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update monitor: %s", err))
		return
	}

	// Convert API response back to Terraform model
	err = r.apiModelToTerraformModel(ctx, monitor, &data)
	if err != nil {
		resp.Diagnostics.AddError("Data Conversion Error", fmt.Sprintf("Unable to convert API response: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API
	err := r.client.DeleteMonitor(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete monitor: %s", err))
		return
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions for data conversion

func (r *MonitorResource) modelToCreateRequest(ctx context.Context, data *MonitorResourceModel) (*client.CreateMonitorRequest, error) {
	req := &client.CreateMonitorRequest{
		Name:          data.Name.ValueString(),
		Active:        data.Active.ValueBool(),
		CheckInterval: int(data.CheckInterval.ValueInt64()),
		Timeout:       int(data.Timeout.ValueInt64()),
		FailThreshold: int(data.FailThreshold.ValueInt64()),
		Settings:      client.MonitorSettings{},
	}

	// Handle regions
	var regions []string
	if !data.Regions.IsNull() {
		data.Regions.ElementsAs(ctx, &regions, false)
		req.Regions = regions
	}

	// Validate fail_threshold
	if req.FailThreshold < 1 {
		return nil, fmt.Errorf("fail_threshold must be at least 1, got %d", req.FailThreshold)
	}
	if len(regions) > 0 && req.FailThreshold > len(regions) {
		return nil, fmt.Errorf("fail_threshold (%d) cannot exceed the number of regions (%d)", req.FailThreshold, len(regions))
	}

	// Handle contacts
	if !data.Contacts.IsNull() {
		var contacts []string
		data.Contacts.ElementsAs(ctx, &contacts, false)
		req.Contacts = contacts
	}

	// Set up type-specific settings based on monitor type
	monitorType := data.Type.ValueString()
	url := data.URL.ValueString()

	switch monitorType {
	case "https":
		httpsSettings := &client.HTTPSSettings{
			URL:                        url,
			CheckCertificateExpiration: true, // Default
			FollowRedirect:             true, // Default
		}

		// Handle additional HTTPS settings if provided
		if !data.HTTPSSettings.IsNull() {
			var tfHttpsSettings HTTPSSettingsModel
			data.HTTPSSettings.As(ctx, &tfHttpsSettings, basetypes.ObjectAsOptions{})

			if !tfHttpsSettings.Method.IsNull() {
				method := tfHttpsSettings.Method.ValueString()
				httpsSettings.HTTPMethod = &method
			}

			if !tfHttpsSettings.ExpectedStatusCodes.IsNull() {
				statuses := tfHttpsSettings.ExpectedStatusCodes.ValueString()
				httpsSettings.HTTPStatuses = &statuses
			}

			if !tfHttpsSettings.CheckCertificateExpiration.IsNull() {
				httpsSettings.CheckCertificateExpiration = tfHttpsSettings.CheckCertificateExpiration.ValueBool()
			}

			if !tfHttpsSettings.FollowRedirects.IsNull() {
				httpsSettings.FollowRedirect = tfHttpsSettings.FollowRedirects.ValueBool()
			}

			if !tfHttpsSettings.RequestHeaders.IsNull() {
				var headers map[string]string
				tfHttpsSettings.RequestHeaders.ElementsAs(ctx, &headers, false)
				var headerStrings []string
				for k, v := range headers {
					headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", k, v))
				}
				headerStr := strings.Join(headerStrings, "\n")
				httpsSettings.RequestHeaders = &headerStr
			}

			if !tfHttpsSettings.RequestBody.IsNull() {
				body := tfHttpsSettings.RequestBody.ValueString()
				httpsSettings.RequestBody = &body
			}

			if !tfHttpsSettings.ExpectedResponseBody.IsNull() {
				respBody := tfHttpsSettings.ExpectedResponseBody.ValueString()
				httpsSettings.ResponseBody = &respBody
			}

			if !tfHttpsSettings.ExpectedResponseHeaders.IsNull() {
				var headers map[string]string
				tfHttpsSettings.ExpectedResponseHeaders.ElementsAs(ctx, &headers, false)
				var headerStrings []string
				for k, v := range headers {
					headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", k, v))
				}
				headerStr := strings.Join(headerStrings, "\n")
				httpsSettings.ResponseHeaders = &headerStr
			}
		}

		req.Settings.HTTPS = httpsSettings

	case "tcp":
		req.Settings.TCP = &client.TCPSettings{
			URL: url,
		}

	case "ping":
		// For ping monitors, the API expects a URL with ping:// scheme
		pingURL := url
		if !strings.HasPrefix(url, "ping://") {
			pingURL = "ping://" + url
		}
		req.Settings.Ping = &client.PingSettings{
			URL: pingURL,
		}

	default:
		return nil, fmt.Errorf("unsupported monitor type: %s", monitorType)
	}

	// Handle host and port fields for certificate monitoring
	if !data.Host.IsNull() {
		req.Host = data.Host.ValueString()
	}
	if !data.Port.IsNull() {
		req.Port = int(data.Port.ValueInt64())
	}

	return req, nil
}

func (r *MonitorResource) modelToUpdateRequest(ctx context.Context, data *MonitorResourceModel) (*client.UpdateMonitorRequest, error) {
	req := &client.UpdateMonitorRequest{}

	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		req.Name = &name
	}

	if !data.Active.IsNull() {
		active := data.Active.ValueBool()
		req.Active = &active
	}

	if !data.CheckInterval.IsNull() {
		interval := int(data.CheckInterval.ValueInt64())
		req.CheckInterval = &interval
	}

	if !data.Timeout.IsNull() {
		timeout := int(data.Timeout.ValueInt64())
		req.Timeout = &timeout
	}

	if !data.FailThreshold.IsNull() {
		failThreshold := int(data.FailThreshold.ValueInt64())
		req.FailThreshold = &failThreshold
	}

	// Handle regions
	var regions []string
	if !data.Regions.IsNull() {
		data.Regions.ElementsAs(ctx, &regions, false)
		req.Regions = regions
	}

	// Validate fail_threshold
	if req.FailThreshold != nil {
		if *req.FailThreshold < 1 {
			return nil, fmt.Errorf("fail_threshold must be at least 1, got %d", *req.FailThreshold)
		}
		if len(regions) > 0 && *req.FailThreshold > len(regions) {
			return nil, fmt.Errorf("fail_threshold (%d) cannot exceed the number of regions (%d)", *req.FailThreshold, len(regions))
		}
	}

	// Handle contacts
	if !data.Contacts.IsNull() {
		var contacts []string
		data.Contacts.ElementsAs(ctx, &contacts, false)
		req.Contacts = contacts
	}

	// Handle type-specific settings
	monitorType := data.Type.ValueString()
	settings := &client.MonitorSettings{}

	switch monitorType {
	case "https":
		httpsSettings := &client.HTTPSSettings{
			URL:                        data.URL.ValueString(),
			CheckCertificateExpiration: true, // Default
			FollowRedirect:             true, // Default
		}

		// Handle additional HTTPS settings if provided
		if !data.HTTPSSettings.IsNull() {
			var tfHttpsSettings HTTPSSettingsModel
			data.HTTPSSettings.As(ctx, &tfHttpsSettings, basetypes.ObjectAsOptions{})

			if !tfHttpsSettings.Method.IsNull() {
				method := tfHttpsSettings.Method.ValueString()
				httpsSettings.HTTPMethod = &method
			}

			if !tfHttpsSettings.ExpectedStatusCodes.IsNull() {
				statuses := tfHttpsSettings.ExpectedStatusCodes.ValueString()
				httpsSettings.HTTPStatuses = &statuses
			}

			if !tfHttpsSettings.CheckCertificateExpiration.IsNull() {
				httpsSettings.CheckCertificateExpiration = tfHttpsSettings.CheckCertificateExpiration.ValueBool()
			}

			if !tfHttpsSettings.FollowRedirects.IsNull() {
				httpsSettings.FollowRedirect = tfHttpsSettings.FollowRedirects.ValueBool()
			}

			if !tfHttpsSettings.RequestHeaders.IsNull() {
				var headers map[string]string
				tfHttpsSettings.RequestHeaders.ElementsAs(ctx, &headers, false)
				var headerStrings []string
				for k, v := range headers {
					headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", k, v))
				}
				headerStr := strings.Join(headerStrings, "\n")
				httpsSettings.RequestHeaders = &headerStr
			}

			if !tfHttpsSettings.RequestBody.IsNull() {
				body := tfHttpsSettings.RequestBody.ValueString()
				httpsSettings.RequestBody = &body
			}

			if !tfHttpsSettings.ExpectedResponseBody.IsNull() {
				respBody := tfHttpsSettings.ExpectedResponseBody.ValueString()
				httpsSettings.ResponseBody = &respBody
			}

			if !tfHttpsSettings.ExpectedResponseHeaders.IsNull() {
				var headers map[string]string
				tfHttpsSettings.ExpectedResponseHeaders.ElementsAs(ctx, &headers, false)
				var headerStrings []string
				for k, v := range headers {
					headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", k, v))
				}
				headerStr := strings.Join(headerStrings, "\n")
				httpsSettings.ResponseHeaders = &headerStr
			}
		}

		settings.HTTPS = httpsSettings

	case "tcp":
		settings.TCP = &client.TCPSettings{
			URL: data.URL.ValueString(),
		}

	case "ping":
		// For ping monitors, the API expects a URL with ping:// scheme
		url := data.URL.ValueString()
		pingURL := url
		if !strings.HasPrefix(url, "ping://") {
			pingURL = "ping://" + url
		}
		settings.Ping = &client.PingSettings{
			URL: pingURL,
		}
	}

	req.Settings = settings

	// Handle host and port fields for certificate monitoring
	if !data.Host.IsNull() {
		host := data.Host.ValueString()
		req.Host = &host
	}
	if !data.Port.IsNull() {
		port := int(data.Port.ValueInt64())
		req.Port = &port
	}

	return req, nil
}

func (r *MonitorResource) apiModelToTerraformModel(ctx context.Context, monitor *client.Monitor, data *MonitorResourceModel) error {
	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.Name)
	data.Active = types.BoolValue(monitor.Active)
	data.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	data.Timeout = types.Int64Value(int64(monitor.Timeout))

	// Note: The API may clamp fail_threshold to the number of regions
	// This is expected behavior to prevent invalid configurations
	data.FailThreshold = types.Int64Value(int64(monitor.FailThreshold))

	// Determine type and URL from settings
	if monitor.Settings.HTTPS != nil {
		data.Type = types.StringValue("https")
		// Normalize URL to match configuration expectations
		normalizedURL := normalizeURL(monitor.Settings.HTTPS.URL)
		data.URL = types.StringValue(normalizedURL)
	} else if monitor.Settings.TCP != nil {
		data.Type = types.StringValue("tcp")
		data.URL = types.StringValue(monitor.Settings.TCP.URL)
	} else if monitor.Settings.Ping != nil {
		data.Type = types.StringValue("ping")
		// Strip ping:// prefix for Terraform state consistency
		pingURL := strings.TrimPrefix(monitor.Settings.Ping.URL, "ping://")
		data.URL = types.StringValue(pingURL)
	} else {
		return fmt.Errorf("monitor has no recognized settings type")
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

	// Handle contacts
	if len(monitor.Contacts) > 0 {
		contacts := make([]types.String, len(monitor.Contacts))
		for i, contact := range monitor.Contacts {
			contacts[i] = types.StringValue(contact)
		}
		contactsList, _ := types.ListValueFrom(ctx, types.StringType, contacts)
		data.Contacts = contactsList
	} else {
		data.Contacts = types.ListNull(types.StringType)
	}

	// Handle HTTPS settings
	if monitor.Settings.HTTPS != nil {
		httpsSettings := HTTPSSettingsModel{
			CheckCertificateExpiration: types.BoolValue(monitor.Settings.HTTPS.CheckCertificateExpiration),
			FollowRedirects:            types.BoolValue(monitor.Settings.HTTPS.FollowRedirect),
		}

		// Handle optional fields that are pointers
		if monitor.Settings.HTTPS.HTTPMethod != nil {
			httpsSettings.Method = types.StringValue(*monitor.Settings.HTTPS.HTTPMethod)
		} else {
			httpsSettings.Method = types.StringValue("HEAD") // Default
		}

		if monitor.Settings.HTTPS.HTTPStatuses != nil {
			httpsSettings.ExpectedStatusCodes = types.StringValue(*monitor.Settings.HTTPS.HTTPStatuses)
		} else {
			httpsSettings.ExpectedStatusCodes = types.StringNull()
		}

		// Convert headers string to Terraform map
		if monitor.Settings.HTTPS.RequestHeaders != nil && *monitor.Settings.HTTPS.RequestHeaders != "" {
			headersMap := make(map[string]types.String)
			// Parse "Header1: Value1\nHeader2: Value2" format
			lines := strings.Split(*monitor.Settings.HTTPS.RequestHeaders, "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						headersMap[key] = types.StringValue(value)
					}
				}
			}
			headersTerraformMap, _ := types.MapValueFrom(ctx, types.StringType, headersMap)
			httpsSettings.RequestHeaders = headersTerraformMap
		} else {
			httpsSettings.RequestHeaders = types.MapNull(types.StringType)
		}

		if monitor.Settings.HTTPS.RequestBody != nil && *monitor.Settings.HTTPS.RequestBody != "" {
			httpsSettings.RequestBody = types.StringValue(*monitor.Settings.HTTPS.RequestBody)
		} else {
			httpsSettings.RequestBody = types.StringNull()
		}

		if monitor.Settings.HTTPS.ResponseBody != nil && *monitor.Settings.HTTPS.ResponseBody != "" {
			httpsSettings.ExpectedResponseBody = types.StringValue(*monitor.Settings.HTTPS.ResponseBody)
		} else {
			httpsSettings.ExpectedResponseBody = types.StringNull()
		}

		if monitor.Settings.HTTPS.ResponseHeaders != nil && *monitor.Settings.HTTPS.ResponseHeaders != "" {
			headersMap := make(map[string]types.String)
			// Parse "Header1: Value1\nHeader2: Value2" format
			lines := strings.Split(*monitor.Settings.HTTPS.ResponseHeaders, "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						headersMap[key] = types.StringValue(value)
					}
				}
			}
			headersTerraformMap, _ := types.MapValueFrom(ctx, types.StringType, headersMap)
			httpsSettings.ExpectedResponseHeaders = headersTerraformMap
		} else {
			httpsSettings.ExpectedResponseHeaders = types.MapNull(types.StringType)
		}

		// Define the attribute types for HTTPS settings
		httpsSettingsAttrs := map[string]attr.Type{
			"method":                       types.StringType,
			"expected_status_codes":        types.StringType,
			"check_certificate_expiration": types.BoolType,
			"follow_redirects":             types.BoolType,
			"request_headers":              types.MapType{ElemType: types.StringType},
			"request_body":                 types.StringType,
			"expected_response_body":       types.StringType,
			"expected_response_headers":    types.MapType{ElemType: types.StringType},
		}

		httpsSettingsObj, _ := types.ObjectValueFrom(ctx, httpsSettingsAttrs, httpsSettings)
		data.HTTPSSettings = httpsSettingsObj
	} else {
		httpsSettingsAttrs := map[string]attr.Type{
			"method":                       types.StringType,
			"expected_status_codes":        types.StringType,
			"check_certificate_expiration": types.BoolType,
			"follow_redirects":             types.BoolType,
			"request_headers":              types.MapType{ElemType: types.StringType},
			"request_body":                 types.StringType,
			"expected_response_body":       types.StringType,
			"expected_response_headers":    types.MapType{ElemType: types.StringType},
		}
		data.HTTPSSettings = types.ObjectNull(httpsSettingsAttrs)
	}

	// Handle TCP settings
	if monitor.Settings.TCP != nil {
		// TCP settings are minimal, just create an empty object
		tcpSettingsAttrs := map[string]attr.Type{}
		tcpSettingsObj, _ := types.ObjectValueFrom(ctx, tcpSettingsAttrs, TCPSettingsModel{})
		data.TCPSettings = tcpSettingsObj
	} else {
		tcpSettingsAttrs := map[string]attr.Type{}
		data.TCPSettings = types.ObjectNull(tcpSettingsAttrs)
	}

	// Handle Ping settings
	if monitor.Settings.Ping != nil {
		// Ping settings are minimal, just create an empty object
		pingSettingsAttrs := map[string]attr.Type{}
		pingSettingsObj, _ := types.ObjectValueFrom(ctx, pingSettingsAttrs, PingSettingsModel{})
		data.PingSettings = pingSettingsObj
	} else {
		pingSettingsAttrs := map[string]attr.Type{}
		data.PingSettings = types.ObjectNull(pingSettingsAttrs)
	}

	// Handle host and port fields
	if monitor.Host != "" {
		data.Host = types.StringValue(monitor.Host)
	} else {
		data.Host = types.StringNull()
	}

	if monitor.Port > 0 {
		data.Port = types.Int64Value(int64(monitor.Port))
	} else {
		data.Port = types.Int64Null()
	}

	return nil
}

// normalizeURL removes trailing slashes from URLs to ensure consistent state
// The API may add trailing slashes, but Terraform configurations typically don't include them
func normalizeURL(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	// Parse the URL to handle it properly
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// If we can't parse it, just remove trailing slash manually
		if strings.HasSuffix(rawURL, "/") {
			return strings.TrimSuffix(rawURL, "/")
		}
		return rawURL
	}

	// Remove trailing slash from path if it's just "/"
	if parsedURL.Path == "/" {
		parsedURL.Path = ""
	}

	return parsedURL.String()
}
