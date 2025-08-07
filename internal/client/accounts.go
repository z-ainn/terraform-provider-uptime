package client

import (
	"encoding/json"
	"fmt"
)

// GetAccount retrieves current account information
func (c *Client) GetAccount() (*Account, error) {
	resp, err := c.doRequest("GET", "/api/account", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var accountResp AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&accountResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if accountResp.Status != "ok" {
		if accountResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", *accountResp.Error)
		}
		return nil, fmt.Errorf("unknown API error")
	}

	if accountResp.Data == nil {
		return nil, fmt.Errorf("no account data in response")
	}

	return accountResp.Data, nil
}
