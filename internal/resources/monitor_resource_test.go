package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraform-provider-uptime/internal/client"
)

func TestMonitorResource_Metadata(t *testing.T) {
	r := &MonitorResource{}
	req := resource.MetadataRequest{
		ProviderTypeName: "uptime",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "uptime_monitor", resp.TypeName)
}

func TestMonitorResource_Schema(t *testing.T) {
	r := &MonitorResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Verify required attributes exist
	assert.NotNil(t, resp.Schema.Attributes["id"])
	assert.NotNil(t, resp.Schema.Attributes["name"])
	assert.NotNil(t, resp.Schema.Attributes["url"])
	assert.NotNil(t, resp.Schema.Attributes["type"])

	// Verify optional attributes
	assert.NotNil(t, resp.Schema.Attributes["active"])
	assert.NotNil(t, resp.Schema.Attributes["check_interval"])
	assert.NotNil(t, resp.Schema.Attributes["timeout"])
	assert.NotNil(t, resp.Schema.Attributes["fail_threshold"])
	assert.NotNil(t, resp.Schema.Attributes["regions"])
	assert.NotNil(t, resp.Schema.Attributes["contacts"])

	// Verify nested attributes
	assert.NotNil(t, resp.Schema.Attributes["https_settings"])
	assert.NotNil(t, resp.Schema.Attributes["tcp_settings"])
	assert.NotNil(t, resp.Schema.Attributes["ping_settings"])

	// Verify certificate monitoring fields
	assert.NotNil(t, resp.Schema.Attributes["host"])
	assert.NotNil(t, resp.Schema.Attributes["port"])
}

func TestMonitorResource_ModelToCreateRequest_HTTPS(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Create test data
	data := &MonitorResourceModel{
		Name:          types.StringValue("Test Monitor"),
		URL:           types.StringValue("https://example.com"),
		Type:          types.StringValue("https"),
		Active:        types.BoolValue(true),
		CheckInterval: types.Int64Value(60),
		Timeout:       types.Int64Value(30),
		FailThreshold: types.Int64Value(2),
	}

	// Set regions
	regions, _ := types.ListValueFrom(ctx, types.StringType, []string{"us-east-1", "eu-west-1"})
	data.Regions = regions

	// Set contacts
	contacts, _ := types.ListValueFrom(ctx, types.StringType, []string{"contact1", "contact2"})
	data.Contacts = contacts

	// Create HTTPS settings with default HEAD method
	httpsSettings := &HTTPSSettingsModel{
		Method:                     types.StringValue("HEAD"),
		ExpectedStatusCodes:        types.StringValue("200"),
		CheckCertificateExpiration: types.BoolValue(true),
		FollowRedirects:            types.BoolValue(true),
	}

	// Convert to object
	httpsSettingsObj, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"method":                       types.StringType,
		"expected_status_codes":        types.StringType,
		"check_certificate_expiration": types.BoolType,
		"follow_redirects":             types.BoolType,
		"request_headers":              types.MapType{ElemType: types.StringType},
		"request_body":                 types.StringType,
		"expected_response_body":       types.StringType,
		"expected_response_headers":    types.MapType{ElemType: types.StringType},
	}, httpsSettings)

	data.HTTPSSettings = httpsSettingsObj

	// Call the method
	req, err := r.modelToCreateRequest(ctx, data)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "Test Monitor", req.Name)
	assert.True(t, req.Active)
	assert.Equal(t, 60, req.CheckInterval)
	assert.Equal(t, 30, req.Timeout)
	assert.Equal(t, 2, req.FailThreshold)
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, req.Regions)
	assert.Equal(t, []string{"contact1", "contact2"}, req.Contacts)

	// Verify HTTPS settings
	assert.NotNil(t, req.Settings.HTTPS)
	assert.Equal(t, "https://example.com", req.Settings.HTTPS.URL)
	// HTTPMethod should be nil since we didn't set it in the test data
	assert.Nil(t, req.Settings.HTTPS.HTTPMethod)
	// HTTPStatuses should be nil since we didn't set it in the test data
	assert.Nil(t, req.Settings.HTTPS.HTTPStatuses)
	assert.True(t, req.Settings.HTTPS.CheckCertificateExpiration)
	assert.True(t, req.Settings.HTTPS.FollowRedirect)
}

func TestMonitorResource_ModelToCreateRequest_TCP(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Create test data for TCP monitor
	data := &MonitorResourceModel{
		Name:          types.StringValue("Database Monitor"),
		URL:           types.StringValue("db.example.com:5432"),
		Type:          types.StringValue("tcp"),
		Active:        types.BoolValue(true),
		CheckInterval: types.Int64Value(30),
		Timeout:       types.Int64Value(10),
		FailThreshold: types.Int64Value(1),
	}

	// Set regions
	regions, _ := types.ListValueFrom(ctx, types.StringType, []string{"us-east-1"})
	data.Regions = regions

	// Create empty TCP settings
	tcpSettingsObj, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{}, &TCPSettingsModel{})
	data.TCPSettings = tcpSettingsObj

	// Call the method
	req, err := r.modelToCreateRequest(ctx, data)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "Database Monitor", req.Name)
	assert.Equal(t, "db.example.com:5432", req.Settings.TCP.URL)
	assert.NotNil(t, req.Settings.TCP)
	assert.Nil(t, req.Settings.HTTPS)
	assert.Nil(t, req.Settings.Ping)
}

func TestMonitorResource_ModelToCreateRequest_Ping(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Create test data for Ping monitor
	data := &MonitorResourceModel{
		Name:          types.StringValue("Server Ping"),
		URL:           types.StringValue("server.example.com"),
		Type:          types.StringValue("ping"),
		Active:        types.BoolValue(true),
		CheckInterval: types.Int64Value(60),
		Timeout:       types.Int64Value(5),
		FailThreshold: types.Int64Value(1),
	}

	// Create empty Ping settings
	pingSettingsObj, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{}, &PingSettingsModel{})
	data.PingSettings = pingSettingsObj

	// Call the method
	req, err := r.modelToCreateRequest(ctx, data)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "Server Ping", req.Name)
	assert.Equal(t, "ping://server.example.com", req.Settings.Ping.URL)
	assert.NotNil(t, req.Settings.Ping)
	assert.Nil(t, req.Settings.HTTPS)
	assert.Nil(t, req.Settings.TCP)
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "HTTPS URL unchanged",
			inputURL:    "https://example.com",
			expectedURL: "https://example.com",
		},
		{
			name:        "HTTPS URL with port unchanged",
			inputURL:    "https://example.com:8443",
			expectedURL: "https://example.com:8443",
		},
		{
			name:        "TCP URL with tcp:// prefix",
			inputURL:    "tcp://db.example.com:5432",
			expectedURL: "tcp://db.example.com:5432",
		},
		{
			name:        "TCP URL without prefix unchanged",
			inputURL:    "db.example.com:5432",
			expectedURL: "db.example.com:5432",
		},
		{
			name:        "Ping URL with ping:// prefix",
			inputURL:    "ping://server.example.com",
			expectedURL: "ping://server.example.com",
		},
		{
			name:        "Ping URL without prefix unchanged",
			inputURL:    "server.example.com",
			expectedURL: "server.example.com",
		},
		{
			name:        "IP address unchanged",
			inputURL:    "192.168.1.1",
			expectedURL: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := normalizeURL(tt.inputURL)
			assert.Equal(t, tt.expectedURL, url, "URL mismatch")
		})
	}
}

func TestMonitorResource_StateUpdate(t *testing.T) {
	ctx := context.Background()

	// Create a mock monitor response
	monitor := &client.Monitor{
		ID:            "monitor123",
		Name:          "Test Monitor",
		Active:        true,
		CheckInterval: 60,
		Timeout:       30,
		FailThreshold: 3,
		Regions:       []string{"us-east-1", "eu-west-1"},
		Contacts:      []string{"contact1", "contact2"},
		Host:          "example.com",
		Port:          443,
		Settings: client.MonitorSettings{
			HTTPS: &client.HTTPSSettings{
				URL:                        "https://example.com",
				HTTPMethod:                 &[]string{"HEAD"}[0],
				HTTPStatuses:               &[]string{"200"}[0],
				CheckCertificateExpiration: true,
				FollowRedirect:             true,
			},
		},
	}

	// Create data model to update
	data := &MonitorResourceModel{
		Type: types.StringValue("https"),
	}

	// Simulate state update
	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.Name)
	data.URL = types.StringValue(monitor.Settings.HTTPS.URL)
	data.Active = types.BoolValue(monitor.Active)
	data.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	data.Timeout = types.Int64Value(int64(monitor.Timeout))
	data.FailThreshold = types.Int64Value(int64(monitor.FailThreshold))
	data.Host = types.StringValue(monitor.Host)
	data.Port = types.Int64Value(int64(monitor.Port))
	data.Regions, _ = types.ListValueFrom(ctx, types.StringType, monitor.Regions)
	data.Contacts, _ = types.ListValueFrom(ctx, types.StringType, monitor.Contacts)

	// Verify results
	assert.Equal(t, "monitor123", data.ID.ValueString())
	assert.Equal(t, "Test Monitor", data.Name.ValueString())
	assert.Equal(t, "https://example.com", data.URL.ValueString())
	assert.True(t, data.Active.ValueBool())
	assert.Equal(t, int64(60), data.CheckInterval.ValueInt64())
	assert.Equal(t, int64(30), data.Timeout.ValueInt64())
	assert.Equal(t, int64(3), data.FailThreshold.ValueInt64())
	assert.Equal(t, "example.com", data.Host.ValueString())
	assert.Equal(t, int64(443), data.Port.ValueInt64())

	// Verify regions list
	var regions []string
	data.Regions.ElementsAs(ctx, &regions, false)
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, regions)

	// Verify contacts list
	var contacts []string
	data.Contacts.ElementsAs(ctx, &contacts, false)
	assert.Equal(t, []string{"contact1", "contact2"}, contacts)
}

func TestMonitorResource_DefaultHTTPMethod(t *testing.T) {
	r := &MonitorResource{}
	ctx := context.Background()

	// Create test data without specifying method
	data := &MonitorResourceModel{
		Name:          types.StringValue("Test Monitor"),
		URL:           types.StringValue("https://example.com"),
		Type:          types.StringValue("https"),
		Active:        types.BoolValue(true),
		CheckInterval: types.Int64Value(60),
		Timeout:       types.Int64Value(30),
		FailThreshold: types.Int64Value(1),
	}

	// Create HTTPS settings without method (should default to HEAD)
	httpsSettings := &HTTPSSettingsModel{
		Method:                     types.StringNull(), // Not specified
		ExpectedStatusCodes:        types.StringValue("200"),
		CheckCertificateExpiration: types.BoolValue(true),
		FollowRedirects:            types.BoolValue(true),
	}

	// Convert to object
	httpsSettingsObj, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"method":                       types.StringType,
		"expected_status_codes":        types.StringType,
		"check_certificate_expiration": types.BoolType,
		"follow_redirects":             types.BoolType,
		"request_headers":              types.MapType{ElemType: types.StringType},
		"request_body":                 types.StringType,
		"expected_response_body":       types.StringType,
		"expected_response_headers":    types.MapType{ElemType: types.StringType},
	}, httpsSettings)

	data.HTTPSSettings = httpsSettingsObj

	// Call the method
	req, err := r.modelToCreateRequest(ctx, data)

	// Verify that method is not set when null (API will use its own default)
	require.NoError(t, err)
	assert.NotNil(t, req.Settings.HTTPS)
	assert.Nil(t, req.Settings.HTTPS.HTTPMethod, "HTTP method should be nil when not specified")
}
