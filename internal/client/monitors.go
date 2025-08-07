package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateMonitor creates a new monitor
func (c *Client) CreateMonitor(req CreateMonitorRequest) (*Monitor, error) {
	resp, err := c.doRequest("POST", "/api/monitors", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var monitorResp MonitorResponse
	if err := json.NewDecoder(resp.Body).Decode(&monitorResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if monitorResp.Status != "ok" {
		if monitorResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Error)
		}
		if monitorResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
	}

	if monitorResp.Data == nil || monitorResp.Data.Monitor == nil {
		return nil, fmt.Errorf("invalid response: missing monitor data")
	}

	return monitorResp.Data.Monitor, nil
}

// GetMonitor retrieves a monitor by ID
func (c *Client) GetMonitor(id string) (*Monitor, error) {
	resp, err := c.doRequest("GET", "/api/monitors/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Monitor not found
	}

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var monitorResp MonitorResponse
	if err := json.NewDecoder(resp.Body).Decode(&monitorResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if monitorResp.Status != "ok" {
		if monitorResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Error)
		}
		if monitorResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
	}

	if monitorResp.Data == nil || monitorResp.Data.Monitor == nil {
		return nil, fmt.Errorf("invalid response: missing monitor data")
	}

	return monitorResp.Data.Monitor, nil
}

// UpdateMonitor updates an existing monitor
func (c *Client) UpdateMonitor(id string, req UpdateMonitorRequest) (*Monitor, error) {
	resp, err := c.doRequest("PUT", "/api/monitors/"+id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update monitor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var monitorResp MonitorResponse
	if err := json.NewDecoder(resp.Body).Decode(&monitorResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if monitorResp.Status != "ok" {
		if monitorResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Error)
		}
		if monitorResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *monitorResp.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
	}

	if monitorResp.Data == nil || monitorResp.Data.Monitor == nil {
		return nil, fmt.Errorf("invalid response: missing monitor data")
	}

	return monitorResp.Data.Monitor, nil
}

// DeleteMonitor deletes a monitor by ID
func (c *Client) DeleteMonitor(id string) error {
	resp, err := c.doRequest("DELETE", "/api/monitors/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Monitor already doesn't exist
	}

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return nil
}

// ListMonitors retrieves all monitors for the authenticated account
func (c *Client) ListMonitors() ([]Monitor, error) {
	resp, err := c.doRequest("GET", "/api/monitors", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var listResp ListMonitorsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if listResp.Status != "ok" {
		if listResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *listResp.Error)
		}
		if listResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *listResp.Message)
		}
		return nil, fmt.Errorf("API error: unknown error")
	}

	if listResp.Data == nil {
		return nil, fmt.Errorf("invalid response: missing data")
	}

	return listResp.Data.Monitors, nil
}
