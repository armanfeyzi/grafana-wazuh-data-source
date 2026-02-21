package wazuhapi

import (
	"context"
	"encoding/json"
	"fmt"
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

type AgentOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
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

func ParseAgentOptions(raw []byte) ([]AgentOption, error) {
	var resp agentsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}

	options := make([]AgentOption, 0, len(resp.Data.AffectedItems))
	for _, item := range resp.Data.AffectedItems {
		if item.Name == "" {
			continue
		}
		label := item.Name
		if item.ID != "" {
			label = fmt.Sprintf("%s (%s)", item.Name, item.ID)
		}
		options = append(options, AgentOption{
			Label: label,
			Value: item.Name,
		})
	}

	return options, nil
}
