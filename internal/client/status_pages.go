package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateStatusPage creates a new status page
func (c *Client) CreateStatusPage(req CreateStatusPageRequest) (*StatusPage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/status_pages", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp StatusPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Error)
		}
		if apiResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if apiResp.Data == nil || apiResp.Data.StatusPage == nil {
		return nil, fmt.Errorf("no status page data in response")
	}

	return apiResp.Data.StatusPage, nil
}

// GetStatusPage retrieves a status page by ID
func (c *Client) GetStatusPage(id string) (*StatusPage, error) {
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/status_pages/%s", c.BaseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("status page not found")
	}

	var apiResp StatusPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Error)
		}
		if apiResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if apiResp.Data == nil || apiResp.Data.StatusPage == nil {
		return nil, fmt.Errorf("no status page data in response")
	}

	return apiResp.Data.StatusPage, nil
}

// UpdateStatusPage updates an existing status page
func (c *Client) UpdateStatusPage(id string, req UpdateStatusPageRequest) (*StatusPage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/status_pages/%s", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp StatusPageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Error)
		}
		if apiResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if apiResp.Data == nil || apiResp.Data.StatusPage == nil {
		return nil, fmt.Errorf("no status page data in response")
	}

	return apiResp.Data.StatusPage, nil
}

// DeleteStatusPage deletes a status page
func (c *Client) DeleteStatusPage(id string) error {
	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/status_pages/%s", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var apiResp StatusPageResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil {
			if apiResp.Error != nil {
				return fmt.Errorf("API error: %s", *apiResp.Error)
			}
			if apiResp.Message != nil {
				return fmt.Errorf("API error: %s", *apiResp.Message)
			}
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ListStatusPages retrieves all status pages
func (c *Client) ListStatusPages() ([]StatusPage, error) {
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/status_pages", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp ListStatusPagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Error)
		}
		if apiResp.Message != nil {
			return nil, fmt.Errorf("API error: %s", *apiResp.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if apiResp.Data == nil {
		return nil, fmt.Errorf("no data in response")
	}

	return apiResp.Data.StatusPages, nil
}
