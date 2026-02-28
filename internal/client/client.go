package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// Config holds the configuration for the IronWiFi API client.
type Config struct {
	APIEndpoint string
	APIToken    string
	Username    string
	Password    string
	ClientID    string
	ClientSecret string
	CompanyID   string
	UserAgent   string
}

// Client is the IronWiFi API client.
type Client struct {
	endpoint  string
	companyID string
	userAgent string
	http      *http.Client
	auth      *authProvider
}

// New creates a new IronWiFi API client.
func New(cfg *Config) (*Client, error) {
	if cfg.APIEndpoint == "" {
		return nil, fmt.Errorf("api_endpoint is required")
	}
	if cfg.CompanyID == "" {
		return nil, fmt.Errorf("company_id is required")
	}
	if cfg.APIToken == "" && cfg.Username == "" {
		return nil, fmt.Errorf("either api_token or username/password is required")
	}

	ua := cfg.UserAgent
	if ua == "" {
		ua = "terraform-provider-ironwifi"
	}

	return &Client{
		endpoint:  strings.TrimRight(cfg.APIEndpoint, "/"),
		companyID: cfg.CompanyID,
		userAgent: ua,
		http:      &http.Client{Timeout: 60 * time.Second},
		auth:      newAuthProvider(cfg),
	}, nil
}

// resourceURL builds the full API URL for a resource.
func (c *Client) resourceURL(resource string, id ...string) string {
	base := fmt.Sprintf("%s/api/%s/%s", c.endpoint, c.companyID, resource)
	if len(id) > 0 && id[0] != "" {
		base += "/" + id[0]
	}
	return base
}

// Create sends a POST request to create a resource.
func (c *Client) Create(resource string, body map[string]interface{}) (map[string]interface{}, error) {
	return c.doJSON(http.MethodPost, c.resourceURL(resource), body)
}

// Read sends a GET request to fetch a single resource.
func (c *Client) Read(resource, id string) (map[string]interface{}, error) {
	return c.doJSON(http.MethodGet, c.resourceURL(resource, id), nil)
}

// Update sends a PATCH request to update a resource.
func (c *Client) Update(resource, id string, body map[string]interface{}) (map[string]interface{}, error) {
	return c.doJSON(http.MethodPatch, c.resourceURL(resource, id), body)
}

// Delete sends a DELETE request to remove a resource.
func (c *Client) Delete(resource, id string) error {
	_, err := c.doJSON(http.MethodDelete, c.resourceURL(resource, id), nil)
	return err
}

// List fetches all items from a collection endpoint, handling pagination.
func (c *Client) List(resource string, embeddedKey string) ([]map[string]interface{}, error) {
	var all []map[string]interface{}
	url := c.resourceURL(resource)

	for url != "" {
		resp, err := c.doRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var hal HALResponse
		if err := json.Unmarshal(resp, &hal); err != nil {
			return nil, fmt.Errorf("parsing HAL response: %w", err)
		}

		if hal.Embedded != nil {
			var embedded map[string]json.RawMessage
			if err := json.Unmarshal(hal.Embedded, &embedded); err != nil {
				return nil, fmt.Errorf("parsing embedded: %w", err)
			}

			if items, ok := embedded[embeddedKey]; ok {
				var page []map[string]interface{}
				if err := json.Unmarshal(items, &page); err != nil {
					return nil, fmt.Errorf("parsing items: %w", err)
				}
				all = append(all, page...)
			}
		}

		url = ""
		if hal.Links.Next != nil && hal.Links.Next.Href != "" {
			nextHref := hal.Links.Next.Href
			if strings.HasPrefix(nextHref, "/") {
				url = c.endpoint + nextHref
			} else {
				url = nextHref
			}
		}
	}

	return all, nil
}

// doJSON performs a JSON request and returns the parsed response.
func (c *Client) doJSON(method, url string, body map[string]interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	respBody, err := c.doRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if len(respBody) == 0 {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result, nil
}

// doRequest performs an HTTP request with auth, retries, and error handling.
func (c *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(backoff)
		}

		var bodyReader io.Reader
		if body != nil {
			// Re-read body for retries
			if seeker, ok := body.(io.ReadSeeker); ok {
				seeker.Seek(0, io.SeekStart)
				bodyReader = seeker
			} else {
				bodyReader = body
			}
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		authHeader, err := c.auth.getAuthHeader()
		if err != nil {
			return nil, fmt.Errorf("getting auth: %w", err)
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/hal+json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return respBody, nil

		case resp.StatusCode == 404:
			return nil, &NotFoundError{Resource: url}

		case resp.StatusCode == 429:
			// Rate limited — retry
			lastErr = fmt.Errorf("rate limited (429)")
			continue

		case resp.StatusCode >= 500:
			// Server error — retry
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
			continue

		default:
			// Client error — parse API error and don't retry
			var apiErr APIError
			if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Status != 0 {
				return nil, &apiErr
			}
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// NotFoundError indicates a 404 response.
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.Resource)
}

// IsNotFound checks if an error is a 404.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
