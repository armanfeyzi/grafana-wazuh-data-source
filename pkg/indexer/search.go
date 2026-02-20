package indexer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) Search(ctx context.Context, indices string, body []byte) ([]byte, error) {
	endpoint := c.baseURL + "/" + indices + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("indexer unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("read search response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("indexer authentication failed")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("indexer returned %s: %s", resp.Status, string(raw))
	}

	return raw, nil
}
