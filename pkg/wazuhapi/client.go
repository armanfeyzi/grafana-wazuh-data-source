package wazuhapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

const tokenTTL = 14 * time.Minute

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

func NewClient(settings *models.PluginSettings, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(settings.ManagerURL, "/"),
		username:   settings.Username,
		password:   settings.Secrets.Password,
		httpClient: httpClient,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	token, err := c.getToken(ctx, false)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/agents?limit=1", nil)
	if err != nil {
		return fmt.Errorf("build manager API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("manager API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.getToken(ctx, true)
		if err != nil {
			return err
		}
		return c.pingWithToken(ctx, token)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("manager API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) pingWithToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/agents?limit=1", nil)
	if err != nil {
		return fmt.Errorf("build manager API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("manager API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("manager API authentication failed")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("manager API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) getToken(ctx context.Context, forceRefresh bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !forceRefresh && c.cachedToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.cachedToken, nil
	}

	token, err := c.authenticate(ctx)
	if err != nil {
		return "", err
	}

	c.cachedToken = token
	c.tokenExpiry = time.Now().Add(tokenTTL)
	return token, nil
}

func (c *Client) authenticate(ctx context.Context) (string, error) {
	endpoint := c.baseURL + "/security/user/authenticate?raw=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build authenticate request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("manager API unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("read authenticate response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("manager API authentication failed")
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("manager API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	token := strings.TrimSpace(string(body))
	if token == "" {
		return "", fmt.Errorf("manager API returned an empty token")
	}

	return token, nil
}

// ValidateURL checks that the manager URL is usable before making requests.
func ValidateURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("manager URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid manager URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("manager URL must use http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("manager URL must include a host")
	}
	return nil
}
