package indexer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

// Search executes an OpenSearch _search request against the given indices.
// It injects a server-side timeout and returns a typed WazuhError on failure.
func (c *Client) Search(ctx context.Context, indices string, body []byte) ([]byte, error) {
	endpoint := c.baseURL + "/" + indices + "/_search?timeout=25s"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, models.NewWazuhError(models.ErrBadResponse,
			"failed to build indexer search request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, classifyNetworkError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes))
	if err != nil {
		return nil, models.NewWazuhError(models.ErrBadResponse,
			"failed to read indexer response", err)
	}

	if resp.StatusCode >= 400 {
		return nil, classifyHTTPError(resp.StatusCode, raw, c.password)
	}

	return raw, nil
}

// classifyHTTPError maps indexer HTTP status codes to typed WazuhErrors.
func classifyHTTPError(status int, body []byte, password string) *models.WazuhError {
	excerpt := sanitizeExcerpt(body, password)
	switch status {
	case http.StatusUnauthorized:
		return models.NewWazuhError(models.ErrAuth,
			"Wazuh indexer: authentication failed — check indexer username and password", nil)
	case http.StatusForbidden:
		return models.NewWazuhError(models.ErrForbidden,
			"Wazuh indexer: permission denied — check indexer RBAC (read access to wazuh-* indices)", nil)
	case http.StatusNotFound:
		return models.NewWazuhError(models.ErrIndexMissing,
			"Wazuh indexer: index not found — Wazuh may not have generated this data yet", nil)
	default:
		return models.NewWazuhError(models.ErrBadResponse,
			fmt.Sprintf("Wazuh indexer returned HTTP %d: %s", status, excerpt), nil)
	}
}

// classifyNetworkError maps transport errors to ErrTimeout or ErrUnreachable.
func classifyNetworkError(err error) *models.WazuhError {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.NewWazuhError(models.ErrTimeout,
			"Wazuh indexer: query timed out — try a shorter time range", err)
	}
	return models.NewWazuhError(models.ErrUnreachable,
		"Wazuh indexer: cannot connect — check the indexer URL and that the service is running", err)
}

// sanitizeExcerpt limits body to 200 bytes and redacts the password.
func sanitizeExcerpt(body []byte, password string) string {
	s := string(body)
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	if password != "" {
		// Use a simple replace; strings import is available via models.
		result := []byte(s)
		pass := []byte(password)
		if len(pass) > 0 {
			var out []byte
			redacted := []byte("[REDACTED]")
			for i := 0; i <= len(result)-len(pass); {
				match := true
				for j, b := range pass {
					if result[i+j] != b {
						match = false
						break
					}
				}
				if match {
					out = append(out, redacted...)
					i += len(pass)
				} else {
					out = append(out, result[i])
					i++
				}
			}
			if len(out) > 0 {
				return string(out)
			}
		}
	}
	return s
}
