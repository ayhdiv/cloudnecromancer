package splunk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps the Splunk REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithInsecureSkipVerify disables TLS certificate verification.
// Common for internal Splunk instances with self-signed certs.
func WithInsecureSkipVerify() ClientOption {
	return func(c *Client) {
		transport := c.httpClient.Transport.(*http.Transport)
		transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // user-opted-in
	}
}

// NewClient creates a Splunk REST API client.
// baseURL is the management API endpoint (e.g. https://splunk.corp:8089).
// token is a Splunk authentication token (bearer token or session key).
func NewClient(baseURL, token string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // searches can be long-running
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			},
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// maxResponseBytes limits the total response body to prevent OOM from a
// malicious or misconfigured Splunk server.
const maxResponseBytes = 10 * 1024 * 1024 * 1024 // 10 GB

// SearchResult represents a single result row from Splunk.
type SearchResult struct {
	Result map[string]any `json:"result"`
	// Preview indicates whether this is a preview (partial) result.
	Preview bool `json:"preview"`
}

// ExportSearch runs a oneshot search using the /services/search/jobs/export
// endpoint. This streams results without creating a persistent search job,
// which is the recommended approach for bulk data extraction.
//
// The callback fn is called for each result row as it arrives.
func (c *Client) ExportSearch(search string, earliest, latest time.Time, fn func(SearchResult) error) error {
	params := url.Values{
		"search":        {search},
		"earliest_time": {earliest.UTC().Format(time.RFC3339)},
		"latest_time":   {latest.UTC().Format(time.RFC3339)},
		"output_mode":   {"json"},
	}

	req, err := http.NewRequest("POST", c.baseURL+"/services/search/jobs/export", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("splunk request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read and discard body to drain the connection, but don't expose
		// Splunk error details (may contain internal hostnames/stack traces).
		_, _ = io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("splunk returned HTTP %d", resp.StatusCode)
	}

	// Limit total response size to prevent OOM
	limited := io.LimitReader(resp.Body, maxResponseBytes)

	// Splunk streams newline-delimited JSON objects
	dec := json.NewDecoder(limited)
	for dec.More() {
		var row SearchResult
		if err := dec.Decode(&row); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
		if row.Preview {
			continue
		}
		if err := fn(row); err != nil {
			return err
		}
	}

	return nil
}
