package indexer

import (
	"bytes"
	"context"
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
		return nil, classifyHTTPError(resp.StatusCode)
	}

	return raw, nil
}
