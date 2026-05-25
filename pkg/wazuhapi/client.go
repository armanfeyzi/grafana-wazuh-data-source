package wazuhapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

const (
	tokenTTL    = 14 * time.Minute
	scaCacheTTL = 45 * time.Second
)

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time

	// scaGroup deduplicates concurrent identical ListSCAForAgents calls so that
	// multiple dashboard panels loading at the same time share a single set of
	// Wazuh API requests instead of each making their own N+1 round-trips.
	scaGroup singleflight.Group
	// scaCache stores the marshalled SCA result for scaCacheTTL so that the
	// 1-minute dashboard auto-refresh does not burst the Wazuh rate limit.
	scaCache *gocache.Cache
}

func NewClient(settings *models.PluginSettings, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(settings.ManagerURL, "/"),
		username:   settings.Username,
		password:   settings.Secrets.Password,
		httpClient: httpClient,
		scaCache:   gocache.New(scaCacheTTL, 2*scaCacheTTL),
	}
}

func (c *Client) Ping(ctx context.Context) error {
	token, err := c.getToken(ctx, false)
	if err != nil {
		return err
	}

	status, body, err := c.pingWithToken(ctx, token)
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized {
		token, err = c.getToken(ctx, true)
		if err != nil {
			return err
		}
		status, body, err = c.pingWithToken(ctx, token)
		if err != nil {
			return err
		}
	}

	if status >= 400 {
		return classifyHTTPError(status, body, "Wazuh manager API", c.password)
	}
	return nil
}

// pingWithToken sends a lightweight GET /agents?limit=1 and returns the HTTP
// status code and response body.
func (c *Client) pingWithToken(ctx context.Context, token string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/agents?limit=1", nil)
	if err != nil {
		return 0, nil, models.NewWazuhError(models.ErrBadResponse,
			"failed to build manager API request", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, classifyNetworkError(err, "Wazuh manager API")
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return resp.StatusCode, body, nil
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
		return "", models.NewWazuhError(models.ErrBadResponse,
			"failed to build authentication request", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", classifyNetworkError(err, "Wazuh manager API")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", models.NewWazuhError(models.ErrBadResponse,
			"failed to read authentication response", err)
	}

	if resp.StatusCode >= 400 {
		return "", classifyHTTPError(resp.StatusCode, body, "Wazuh manager API", c.password)
	}

	token := strings.TrimSpace(string(body))
	if token == "" {
		return "", models.NewWazuhError(models.ErrBadResponse,
			"Wazuh manager API returned an empty token — check credentials", nil)
	}

	return token, nil
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	token, err := c.getToken(ctx, false)
	if err != nil {
		return nil, err
	}

	body, status, err := c.getWithToken(ctx, path, token)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized {
		token, err = c.getToken(ctx, true)
		if err != nil {
			return nil, err
		}
		body, status, err = c.getWithToken(ctx, path, token)
		if err != nil {
			return nil, err
		}
	}

	if status >= 400 {
		return nil, classifyHTTPError(status, body, "Wazuh manager API", c.password)
	}

	return body, nil
}

func (c *Client) getWithToken(ctx context.Context, path, token string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, 0, models.NewWazuhError(models.ErrBadResponse,
			"failed to build manager API request", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, classifyNetworkError(err, "Wazuh manager API")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, models.NewWazuhError(models.ErrBadResponse,
			"failed to read manager API response", err)
	}

	return body, resp.StatusCode, nil
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

// classifyHTTPError maps an HTTP status code to a typed WazuhError with a
// user-readable message. The password is used only to sanitize the body
// excerpt so credentials never appear in error output.
func classifyHTTPError(status int, body []byte, component, password string) *models.WazuhError {
	excerpt := sanitizeExcerpt(body, password)
	switch status {
	case http.StatusUnauthorized:
		return models.NewWazuhError(models.ErrAuth,
			component+": authentication failed — check username and password", nil)
	case http.StatusForbidden:
		return models.NewWazuhError(models.ErrForbidden,
			component+": permission denied — check RBAC roles (agent:read, sca:read)", nil)
	case http.StatusNotFound:
		return models.NewWazuhError(models.ErrIndexMissing,
			component+": endpoint not found — check the manager URL", nil)
	default:
		return models.NewWazuhError(models.ErrBadResponse,
			fmt.Sprintf("%s returned HTTP %d: %s", component, status, excerpt), nil)
	}
}

// classifyNetworkError maps a transport-level error to ErrTimeout or
// ErrUnreachable.
func classifyNetworkError(err error, component string) *models.WazuhError {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.NewWazuhError(models.ErrTimeout,
			component+": request timed out — check connectivity or try again", err)
	}
	return models.NewWazuhError(models.ErrUnreachable,
		component+": cannot connect — check the URL and that the service is running", err)
}

// sanitizeExcerpt limits body to 200 characters and redacts any password.
func sanitizeExcerpt(body []byte, password string) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	if password != "" {
		s = strings.ReplaceAll(s, password, "[REDACTED]")
	}
	return s
}
