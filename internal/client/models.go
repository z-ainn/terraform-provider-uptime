package client

import "encoding/json"

// Monitor represents a monitor in the API
type Monitor struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Active        bool            `json:"active"`
	CheckInterval int             `json:"check_interval"`
	Timeout       int             `json:"timeout"`
	FailThreshold int             `json:"fail_threshold"`
	Regions       []string        `json:"regions,omitempty"`
	Settings      MonitorSettings `json:"settings"`
	Contacts      []string        `json:"contacts,omitempty"`
	LastStatus    string          `json:"last_status,omitempty"`
	CreatedAt     int64           `json:"created_at,omitempty"`
	UpdatedAt     int64           `json:"updated_at,omitempty"`
	Host          string          `json:"host,omitempty"`
	Port          int             `json:"port,omitempty"`
}

// MonitorSettings contains type-specific settings for monitors
// This is a discriminated union - only one field should be set based on monitor type
type MonitorSettings struct {
	HTTPS *HTTPSSettings `json:"https,omitempty"`
	TCP   *TCPSettings   `json:"tcp,omitempty"`
	Ping  *PingSettings  `json:"ping,omitempty"`
}

// HTTPSSettings represents HTTPS-specific monitor configuration
type HTTPSSettings struct {
	URL                        string  `json:"url"`
	HTTPMethod                 *string `json:"http_method,omitempty"`
	RequestHeaders             *string `json:"request_headers,omitempty"`
	RequestBody                *string `json:"request_body,omitempty"`
	HTTPStatuses               *string `json:"http_statuses,omitempty"`
	ResponseHeaders            *string `json:"response_headers,omitempty"`
	ResponseBody               *string `json:"response_body,omitempty"`
	CheckCertificateExpiration bool    `json:"check_certificate_expiration"`
	FollowRedirect             bool    `json:"follow_redirect"`
}

// TCPSettings represents TCP-specific monitor configuration
type TCPSettings struct {
	URL string `json:"url"`
}

// PingSettings represents Ping-specific monitor configuration
type PingSettings struct {
	URL string `json:"url"`
}

// CreateMonitorRequest represents the request payload for creating a monitor
type CreateMonitorRequest struct {
	Name          string          `json:"name"`
	Active        bool            `json:"active"`
	CheckInterval int             `json:"check_interval"`
	Timeout       int             `json:"timeout"`
	FailThreshold int             `json:"fail_threshold"`
	Regions       []string        `json:"regions,omitempty"`
	Settings      MonitorSettings `json:"settings"`
	Contacts      []string        `json:"contacts,omitempty"`
	Host          string          `json:"host,omitempty"`
	Port          int             `json:"port,omitempty"`
}

// UpdateMonitorRequest represents the request payload for updating a monitor
type UpdateMonitorRequest struct {
	Name          *string          `json:"name,omitempty"`
	Active        *bool            `json:"active,omitempty"`
	CheckInterval *int             `json:"check_interval,omitempty"`
	Timeout       *int             `json:"timeout,omitempty"`
	FailThreshold *int             `json:"fail_threshold,omitempty"`
	Regions       []string         `json:"regions,omitempty"`
	Settings      *MonitorSettings `json:"settings,omitempty"`
	Contacts      []string         `json:"contacts,omitempty"`
	Host          *string          `json:"host,omitempty"`
	Port          *int             `json:"port,omitempty"`
}

// MonitorResponse represents the API response for monitor operations
type MonitorResponse struct {
	Status  string       `json:"status"`
	Data    *MonitorData `json:"data,omitempty"`
	Error   *string      `json:"error,omitempty"`
	Message *string      `json:"message,omitempty"`
}

// MonitorData wraps the monitor in the API response
type MonitorData struct {
	Monitor *Monitor `json:"monitor,omitempty"`
}

// ListMonitorsResponse represents the API response for listing monitors
type ListMonitorsResponse struct {
	Status  string            `json:"status"`
	Data    *ListMonitorsData `json:"data,omitempty"`
	Error   *string           `json:"error,omitempty"`
	Message *string           `json:"message,omitempty"`
}

// ListMonitorsData contains the monitors array and pagination
type ListMonitorsData struct {
	Monitors   []Monitor   `json:"monitors"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Contact represents a notification contact with various channel types
type Contact struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Channel        string          `json:"channel"`
	Details        json.RawMessage `json:"details"`
	Active         bool            `json:"active"`
	DownAlertsOnly bool            `json:"down_alerts_only"`
	Error          *string         `json:"error,omitempty"`
	CreatedAt      int64           `json:"created_at,omitempty"`
}

// CreateContactRequest represents the request to create a contact
type CreateContactRequest struct {
	Name           string          `json:"name"`
	Channel        string          `json:"channel"`
	Details        json.RawMessage `json:"details"`
	Active         bool            `json:"active"`
	DownAlertsOnly bool            `json:"down_alerts_only"`
}

// UpdateContactRequest represents the request to update a contact
type UpdateContactRequest struct {
	Name           *string         `json:"name,omitempty"`
	Details        json.RawMessage `json:"details,omitempty"`
	Active         *bool           `json:"active,omitempty"`
	DownAlertsOnly *bool           `json:"down_alerts_only,omitempty"`
}

// ContactResponse represents the API response for contact operations
type ContactResponse struct {
	Status  string       `json:"status"`
	Data    *ContactData `json:"data,omitempty"`
	Error   *string      `json:"error,omitempty"`
	Message *string      `json:"message,omitempty"`
}

// ContactData wraps the contact in the API response
type ContactData struct {
	Contact *Contact `json:"contact,omitempty"`
}

// ListContactsResponse represents the API response for listing contacts
type ListContactsResponse struct {
	Status  string            `json:"status"`
	Data    *ListContactsData `json:"data,omitempty"`
	Error   *string           `json:"error,omitempty"`
	Message *string           `json:"message,omitempty"`
}

// ListContactsData contains the contacts array
type ListContactsData struct {
	Contacts []Contact `json:"contacts"`
}

// Account represents account information returned by the API
type Account struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	CurrentPlan    string `json:"current_plan"`
	MonitorsLimit  int    `json:"monitors_limit"`
	MonitorsCount  int    `json:"monitors_count"`
	UpMonitors     int    `json:"up_monitors"`
	DownMonitors   int    `json:"down_monitors"`
	PausedMonitors int    `json:"paused_monitors"`
}

// AccountResponse represents the API response for account operations
type AccountResponse struct {
	Status string   `json:"status"`
	Data   *Account `json:"data,omitempty"`
	Error  *string  `json:"error,omitempty"`
}

// StatusPage represents a status page in the API
type StatusPage struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Monitors            []string `json:"monitors"`
	Period              int      `json:"period"`
	CustomDomain        *string  `json:"custom_domain,omitempty"`
	ShowIncidentReasons bool     `json:"show_incident_reasons"`
	BasicAuth           *string  `json:"basic_auth,omitempty"`
	CreatedAt           int64    `json:"created_at,omitempty"`
	URL                 string   `json:"url,omitempty"`
}

// CreateStatusPageRequest represents the request to create a status page
type CreateStatusPageRequest struct {
	Name                string   `json:"name"`
	Monitors            []string `json:"monitors"`
	Period              *int     `json:"period,omitempty"`
	CustomDomain        *string  `json:"custom_domain,omitempty"`
	ShowIncidentReasons *bool    `json:"show_incident_reasons,omitempty"`
	BasicAuth           *string  `json:"basic_auth,omitempty"`
}

// UpdateStatusPageRequest represents the request to update a status page
type UpdateStatusPageRequest struct {
	Name                *string  `json:"name,omitempty"`
	Monitors            []string `json:"monitors,omitempty"`
	Period              *int     `json:"period,omitempty"`
	CustomDomain        *string  `json:"custom_domain,omitempty"`
	ShowIncidentReasons *bool    `json:"show_incident_reasons,omitempty"`
	BasicAuth           *string  `json:"basic_auth,omitempty"`
}

// StatusPageResponse represents the API response for status page operations
type StatusPageResponse struct {
	Status  string          `json:"status"`
	Data    *StatusPageData `json:"data,omitempty"`
	Error   *string         `json:"error,omitempty"`
	Message *string         `json:"message,omitempty"`
}

// StatusPageData wraps the status page in the API response
type StatusPageData struct {
	StatusPage *StatusPage `json:"status_page,omitempty"`
}

// ListStatusPagesResponse represents the API response for listing status pages
type ListStatusPagesResponse struct {
	Status  string               `json:"status"`
	Data    *ListStatusPagesData `json:"data,omitempty"`
	Error   *string              `json:"error,omitempty"`
	Message *string              `json:"message,omitempty"`
}

// ListStatusPagesData contains the status pages array and pagination
type ListStatusPagesData struct {
	StatusPages []StatusPage `json:"status_pages"`
	Pagination  *Pagination  `json:"pagination,omitempty"`
}
