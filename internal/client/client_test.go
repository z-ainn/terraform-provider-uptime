package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		apiKey      string
		wantBaseURL string
	}{
		{
			name:        "Valid configuration",
			baseURL:     "https://api.example.com",
			apiKey:      "test-api-key",
			wantBaseURL: "https://api.example.com",
		},
		{
			name:        "Empty base URL",
			baseURL:     "",
			apiKey:      "test-api-key",
			wantBaseURL: "",
		},
		{
			name:        "Empty API key",
			baseURL:     "https://api.example.com",
			apiKey:      "",
			wantBaseURL: "https://api.example.com",
		},
		{
			name:        "Base URL with trailing slash",
			baseURL:     "https://api.example.com/",
			apiKey:      "test-api-key",
			wantBaseURL: "https://api.example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.apiKey)

			assert.NotNil(t, client)
			assert.Equal(t, tt.baseURL, client.BaseURL)
			assert.Equal(t, tt.apiKey, client.APIKey)
			assert.NotNil(t, client.HTTPClient)
		})
	}
}

func TestClient_CreateMonitor(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/monitors", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Decode request body
		var req CreateMonitorRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request data
		assert.Equal(t, "Test Monitor", req.Name)
		assert.True(t, req.Active)
		assert.Equal(t, 60, req.CheckInterval)

		// Send response
		resp := MonitorResponse{
			Status: "ok",
			Data: &MonitorData{
				Monitor: &Monitor{
					ID:            "monitor123",
					Name:          req.Name,
					Active:        req.Active,
					CheckInterval: req.CheckInterval,
					Timeout:       req.Timeout,
					FailThreshold: req.FailThreshold,
					Settings:      req.Settings,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Create monitor request
	req := &CreateMonitorRequest{
		Name:          "Test Monitor",
		Active:        true,
		CheckInterval: 60,
		Timeout:       30,
		FailThreshold: 1,
		Settings: MonitorSettings{
			HTTPS: &HTTPSSettings{
				URL:        "https://example.com",
				HTTPMethod: stringPtr("head"),
			},
		},
	}

	// Call CreateMonitor
	monitor, err := client.CreateMonitor(*req)

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, monitor)
	assert.Equal(t, "monitor123", monitor.ID)
	assert.Equal(t, "Test Monitor", monitor.Name)
	assert.True(t, monitor.Active)
}

func TestClient_GetMonitor(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/monitors/monitor123", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Send response
		resp := MonitorResponse{
			Status: "ok",
			Data: &MonitorData{
				Monitor: &Monitor{
					ID:            "monitor123",
					Name:          "Test Monitor",
					Active:        true,
					CheckInterval: 60,
					Timeout:       30,
					FailThreshold: 1,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Call GetMonitor
	monitor, err := client.GetMonitor("monitor123")

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, monitor)
	assert.Equal(t, "monitor123", monitor.ID)
	assert.Equal(t, "Test Monitor", monitor.Name)
}

func TestClient_UpdateMonitor(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/monitors/monitor123", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Decode request body
		var req UpdateMonitorRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request data
		assert.Equal(t, "Updated Monitor", *req.Name)
		assert.False(t, *req.Active)

		// Send response
		resp := MonitorResponse{
			Status: "ok",
			Data: &MonitorData{
				Monitor: &Monitor{
					ID:            "monitor123",
					Name:          *req.Name,
					Active:        *req.Active,
					CheckInterval: 60,
					Timeout:       30,
					FailThreshold: 1,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Create update request
	name := "Updated Monitor"
	active := false
	req := UpdateMonitorRequest{
		Name:   &name,
		Active: &active,
	}

	// Call UpdateMonitor
	monitor, err := client.UpdateMonitor("monitor123", req)

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, monitor)
	assert.Equal(t, "Updated Monitor", monitor.Name)
	assert.False(t, monitor.Active)
}

func TestClient_DeleteMonitor(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/monitors/monitor123", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Send response
		resp := MonitorResponse{
			Status:  "ok",
			Message: stringPtr("Monitor deleted successfully"),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Call DeleteMonitor
	err := client.DeleteMonitor("monitor123")

	// Verify no error
	assert.NoError(t, err)
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		response      interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:       "404 Not Found",
			statusCode: 404,
			response: MonitorResponse{
				Status: "error",
				Error:  stringPtr("Monitor not found"),
			},
			expectError:   false, // 404 returns nil monitor, no error
			expectedError: "",
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			response: MonitorResponse{
				Status: "error",
				Error:  stringPtr("Invalid API key"),
			},
			expectError:   true,
			expectedError: "HTTP 401",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			response: MonitorResponse{
				Status: "error",
				Error:  stringPtr("Internal server error"),
			},
			expectError:   true,
			expectedError: "HTTP 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if err := json.NewEncoder(w).Encode(tt.response); err != nil {
					t.Errorf("Failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key")

			// Call GetMonitor (any method will do for error testing)
			monitor, err := client.GetMonitor("monitor123")

			// Verify error
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Nil(t, monitor)
		})
	}
}

func TestClient_HTTPSSettingsDefaults(t *testing.T) {
	settings := &HTTPSSettings{
		URL:        "https://example.com",
		HTTPMethod: stringPtr(""), // Empty, should default to HEAD
	}

	// In actual implementation, the default is set at the resource level
	// This test verifies the structure is correct
	assert.Equal(t, "https://example.com", settings.URL)
	assert.Equal(t, "", *settings.HTTPMethod) // Will be defaulted to "head" in resource
}

func TestClient_GetAccount(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/account", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Send response
		resp := AccountResponse{
			Status: "ok",
			Data: &Account{
				ID:             "account123",
				Email:          "test@example.com",
				CurrentPlan:    "10-monthly",
				MonitorsLimit:  100,
				MonitorsCount:  25,
				UpMonitors:     20,
				DownMonitors:   3,
				PausedMonitors: 2,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Call GetAccount
	account, err := client.GetAccount()

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "account123", account.ID)
	assert.Equal(t, "test@example.com", account.Email)
	assert.Equal(t, "10-monthly", account.CurrentPlan)
	assert.Equal(t, 100, account.MonitorsLimit)
	assert.Equal(t, 25, account.MonitorsCount)
	assert.Equal(t, 20, account.UpMonitors)
	assert.Equal(t, 3, account.DownMonitors)
	assert.Equal(t, 2, account.PausedMonitors)
}

func TestClient_GetAccount_Error(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AccountResponse{
			Status: "error",
			Error:  stringPtr("Invalid API key"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key")

	// Call GetAccount
	account, err := client.GetAccount()

	// Verify error response
	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "HTTP 401")
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
