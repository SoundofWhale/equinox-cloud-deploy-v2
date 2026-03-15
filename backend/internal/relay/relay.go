package relay

import (
	"bytes"
	"fmt"
	"net/http"
)

// Client is an HTTP client for the E2EE Blind Relay server.
// The relay stores only encrypted blobs — it never sees plaintext.
type Client struct {
	BaseURL    string       // e.g. "http://localhost:9090"
	HTTPClient *http.Client
}

// NewClient creates a relay client pointed at the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// UploadBackup uploads an encrypted backup blob to the relay.
// POST /api/v1/backup
func (c *Client) UploadBackup(encryptedBlob []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/backup", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(encryptedBlob))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	return c.HTTPClient.Do(req)
}

// GetBackup retrieves an encrypted backup blob by hash ID.
// GET /api/v1/backup/:id
func (c *Client) GetBackup(hashID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/backup/%s", c.BaseURL, hashID)
	return c.HTTPClient.Get(url)
}

// PushDelta uploads an encrypted delta (sync change) to the relay.
// POST /api/v1/sync/delta
func (c *Client) PushDelta(encryptedDelta []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/sync/delta", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(encryptedDelta))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	return c.HTTPClient.Do(req)
}

// PullDeltas fetches encrypted deltas since the given timestamp.
// GET /api/v1/sync/delta?since=<timestamp>
func (c *Client) PullDeltas(sinceTimestamp string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/sync/delta?since=%s", c.BaseURL, sinceTimestamp)
	return c.HTTPClient.Get(url)
}
