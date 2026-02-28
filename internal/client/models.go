package client

import "encoding/json"

// HALResponse represents a HAL+JSON response from the IronWiFi API.
type HALResponse struct {
	Embedded   json.RawMessage        `json:"_embedded"`
	Links      HALLinks               `json:"_links"`
	Page       int                    `json:"page"`
	PageCount  int                    `json:"page_count"`
	PageSize   int                    `json:"page_size"`
	TotalItems int                    `json:"total_items"`
	Extra      map[string]interface{} `json:"-"`
}

// HALLinks represents pagination links in HAL responses.
type HALLinks struct {
	Self  *HALLink `json:"self,omitempty"`
	First *HALLink `json:"first,omitempty"`
	Last  *HALLink `json:"last,omitempty"`
	Next  *HALLink `json:"next,omitempty"`
	Prev  *HALLink `json:"prev,omitempty"`
}

// HALLink is a single HAL link.
type HALLink struct {
	Href string `json:"href"`
}

// APIError represents an RFC 7807 problem detail from the API.
type APIError struct {
	Type               string                 `json:"type"`
	Title              string                 `json:"title"`
	Status             int                    `json:"status"`
	Detail             string                 `json:"detail"`
	ValidationMessages map[string]interface{} `json:"validation_messages,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Title
}

// OAuthTokenResponse is the response from the OAuth2 token endpoint.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}
