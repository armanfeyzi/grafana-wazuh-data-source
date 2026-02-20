package indexer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func NewClient(settings *models.PluginSettings, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(settings.IndexerURL, "/"),
		username:   settings.IndexerUser(),
		password:   settings.IndexerPass(),
		httpClient: httpClient,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_cluster/health", nil)
	if err != nil {
		return fmt.Errorf("build indexer request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("indexer unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("indexer authentication failed")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("indexer returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func ValidateURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("indexer URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid indexer URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("indexer URL must use http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("indexer URL must include a host")
	}
	return nil
}
