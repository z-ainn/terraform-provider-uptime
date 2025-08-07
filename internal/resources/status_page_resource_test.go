package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCustomDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid domain",
			domain:  "status.example.com",
			wantErr: false,
		},
		{
			name:    "valid subdomain",
			domain:  "my.status.example.com",
			wantErr: false,
		},
		{
			name:    "domain ending with uptime-monitor.io",
			domain:  "status.uptime-monitor.io",
			wantErr: true,
			errMsg:  "custom domain cannot end with 'uptime-monitor.io'",
		},
		{
			name:    "subdomain of uptime-monitor.io",
			domain:  "my.uptime-monitor.io",
			wantErr: true,
			errMsg:  "custom domain cannot end with 'uptime-monitor.io'",
		},
		{
			name:    "TLD only",
			domain:  "com",
			wantErr: true,
			errMsg:  "custom domain must contain at least one dot",
		},
		{
			name:    "single word without dot",
			domain:  "localhost",
			wantErr: true,
			errMsg:  "custom domain must contain at least one dot",
		},
		{
			name:    "domain with forward slash",
			domain:  "example.com/path",
			wantErr: true,
			errMsg:  "custom domain cannot contain forward slashes",
		},
		{
			name:    "domain with multiple forward slashes",
			domain:  "example.com/path/to/page",
			wantErr: true,
			errMsg:  "custom domain cannot contain forward slashes",
		},
		{
			name:    "domain with protocol",
			domain:  "https://example.com",
			wantErr: true,
			errMsg:  "custom domain cannot contain forward slashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomDomain(tt.domain)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}