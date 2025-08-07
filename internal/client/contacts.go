package client

import (
	"encoding/json"
	"fmt"
)

// CreateContact creates a new contact
func (c *Client) CreateContact(req *CreateContactRequest) (*Contact, error) {
	resp, err := c.doRequest("POST", "/api/contacts", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var contactResp ContactResponse
	if err := json.NewDecoder(resp.Body).Decode(&contactResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if contactResp.Data == nil || contactResp.Data.Contact == nil {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contactResp.Data.Contact, nil
}

// GetContact retrieves a contact by ID
func (c *Client) GetContact(id string) (*Contact, error) {
	resp, err := c.doRequest("GET", "/api/contacts/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	defer resp.Body.Close()

	// Handle 404 as nil (not found)
	if resp.StatusCode == 404 {
		return nil, nil
	}

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var contactResp ContactResponse
	if err := json.NewDecoder(resp.Body).Decode(&contactResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if contactResp.Data == nil || contactResp.Data.Contact == nil {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contactResp.Data.Contact, nil
}

// UpdateContact updates an existing contact
func (c *Client) UpdateContact(id string, req *UpdateContactRequest) (*Contact, error) {
	resp, err := c.doRequest("PUT", "/api/contacts/"+id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var contactResp ContactResponse
	if err := json.NewDecoder(resp.Body).Decode(&contactResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if contactResp.Data == nil || contactResp.Data.Contact == nil {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contactResp.Data.Contact, nil
}

// DeleteContact deletes a contact
func (c *Client) DeleteContact(id string) error {
	resp, err := c.doRequest("DELETE", "/api/contacts/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return nil
}

// ListContacts retrieves all contacts
func (c *Client) ListContacts() ([]Contact, error) {
	resp, err := c.doRequest("GET", "/api/contacts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var contactsResp ListContactsResponse
	if err := json.NewDecoder(resp.Body).Decode(&contactsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if contactsResp.Data == nil {
		return nil, fmt.Errorf("unexpected response format")
	}

	return contactsResp.Data.Contacts, nil
}
