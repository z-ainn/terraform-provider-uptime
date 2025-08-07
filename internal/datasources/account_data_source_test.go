package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountDataSource_Schema(t *testing.T) {
	// Create the data source
	dataSource := NewAccountDataSource()

	// Create schema request
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	// Call Schema method
	dataSource.Schema(context.Background(), req, resp)

	// Verify no diagnostics errors
	assert.False(t, resp.Diagnostics.HasError())

	// Verify schema attributes exist
	attrs := resp.Schema.Attributes
	assert.NotNil(t, attrs)

	// Check required attributes
	expectedAttributes := []string{
		"id", "email", "current_plan", "monitors_limit",
		"monitors_count", "up_monitors", "down_monitors", "paused_monitors",
	}

	for _, attrName := range expectedAttributes {
		attr, exists := attrs[attrName]
		assert.True(t, exists, "Attribute %s should exist", attrName)
		assert.True(t, attr.IsComputed(), "Attribute %s should be computed", attrName)
		assert.False(t, attr.IsRequired(), "Attribute %s should not be required", attrName)
		assert.False(t, attr.IsOptional(), "Attribute %s should not be optional", attrName)
	}

	// Verify specific attribute types
	idAttr, _ := attrs["id"].(schema.StringAttribute)
	assert.True(t, idAttr.Computed)

	emailAttr, _ := attrs["email"].(schema.StringAttribute)
	assert.True(t, emailAttr.Computed)

	monitorsLimitAttr, _ := attrs["monitors_limit"].(schema.Int64Attribute)
	assert.True(t, monitorsLimitAttr.Computed)

	monitorsCountAttr, _ := attrs["monitors_count"].(schema.Int64Attribute)
	assert.True(t, monitorsCountAttr.Computed)
}

func TestAccountDataSource_Metadata(t *testing.T) {
	// Create the data source
	dataSource := NewAccountDataSource()

	// Create metadata request
	req := datasource.MetadataRequest{
		ProviderTypeName: "uptime",
	}
	resp := &datasource.MetadataResponse{}

	// Call Metadata method
	dataSource.Metadata(context.Background(), req, resp)

	// Verify type name
	assert.Equal(t, "uptime_account", resp.TypeName)
}

func TestAccountDataSourceModel_TypeValidation(t *testing.T) {
	// Test that the model fields have correct types
	model := AccountDataSourceModel{
		ID:             types.StringValue("account123"),
		Email:          types.StringValue("test@example.com"),
		CurrentPlan:    types.StringValue("10-monthly"),
		MonitorsLimit:  types.Int64Value(100),
		MonitorsCount:  types.Int64Value(25),
		UpMonitors:     types.Int64Value(20),
		DownMonitors:   types.Int64Value(3),
		PausedMonitors: types.Int64Value(2),
	}

	// Verify types are correctly assigned
	assert.Equal(t, "account123", model.ID.ValueString())
	assert.Equal(t, "test@example.com", model.Email.ValueString())
	assert.Equal(t, "10-monthly", model.CurrentPlan.ValueString())
	assert.Equal(t, int64(100), model.MonitorsLimit.ValueInt64())
	assert.Equal(t, int64(25), model.MonitorsCount.ValueInt64())
	assert.Equal(t, int64(20), model.UpMonitors.ValueInt64())
	assert.Equal(t, int64(3), model.DownMonitors.ValueInt64())
	assert.Equal(t, int64(2), model.PausedMonitors.ValueInt64())
}

func TestAccountDataSource_NewDataSource(t *testing.T) {
	// Test that NewAccountDataSource creates a valid data source
	dataSource := NewAccountDataSource()
	assert.NotNil(t, dataSource)

	// Verify it implements the DataSource interface
	var _ datasource.DataSource = dataSource
}
