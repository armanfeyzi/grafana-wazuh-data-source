package wazuhapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type agentsResponse struct {
	Data struct {
		AffectedItems []agentItem `json:"affected_items"`
	} `json:"data"`
}

type agentItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	IP            string `json:"ip"`
	Version       string `json:"version"`
	LastKeepAlive string `json:"lastKeepAlive"`
	OS            struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Version  string `json:"version"`
	} `json:"os"`
}

func (c *Client) ListAgents(ctx context.Context, limit int) ([]byte, error) {
	if limit <= 0 {
		limit = 500
	}
	if limit > 1000 {
		limit = 1000
	}

	path := fmt.Sprintf("/agents?limit=%s", url.QueryEscape(strconv.Itoa(limit)))
	return c.get(ctx, path)
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

	if status == http.StatusUnauthorized {
		return nil, fmt.Errorf("manager API authentication failed")
	}
	if status >= 400 {
		return nil, fmt.Errorf("manager API returned %s: %s", http.StatusText(status), string(body))
	}

	return body, nil
}

func (c *Client) getWithToken(ctx context.Context, path, token string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build manager API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("manager API unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read manager API response: %w", err)
	}

	return body, resp.StatusCode, nil
}

func ParseAgentsFrame(raw []byte, refID string) (*data.Frame, error) {
	var resp agentsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}

	items := resp.Data.AffectedItems
	ids := make([]string, len(items))
	names := make([]string, len(items))
	statuses := make([]string, len(items))
	ips := make([]string, len(items))
	osNames := make([]string, len(items))
	osVersions := make([]string, len(items))
	versions := make([]string, len(items))
	lastSeen := make([]string, len(items))

	for i, item := range items {
		ids[i] = item.ID
		names[i] = item.Name
		statuses[i] = item.Status
		ips[i] = item.IP
		osNames[i] = item.OS.Name
		if osNames[i] == "" {
			osNames[i] = item.OS.Platform
		}
		osVersions[i] = item.OS.Version
		versions[i] = item.Version
		lastSeen[i] = item.LastKeepAlive
	}

	frame := data.NewFrame("agents")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("agent_id", nil, ids),
		data.NewField("agent", nil, names),
		data.NewField("status", nil, statuses),
		data.NewField("host_ip", nil, ips),
		data.NewField("os", nil, osNames),
		data.NewField("os_version", nil, osVersions),
		data.NewField("wazuh_version", nil, versions),
		data.NewField("last_keepalive", nil, lastSeen),
	)

	return frame, nil
}
