package wazuhapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

type scaResponse struct {
	Data struct {
		AffectedItems []scaPolicyItem `json:"affected_items"`
	} `json:"data"`
}

type scaPolicyItem struct {
	PolicyID    string  `json:"policy_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Pass        int64   `json:"pass"`
	Fail        int64   `json:"fail"`
	Score       float64 `json:"score"`
	TotalChecks int64   `json:"total_checks"`
	EndScan     string  `json:"end_scan"`
}

func (c *Client) GetSCA(ctx context.Context, agentID string) ([]byte, error) {
	path := fmt.Sprintf("/sca/%s", url.PathEscape(agentID))
	return c.get(ctx, path)
}

func (c *Client) ListSCAForAgents(ctx context.Context, agentNames []string, agentLimit int) ([]byte, error) {
	agentsRaw, err := c.ListAgents(ctx, agentLimit)
	if err != nil {
		return nil, err
	}

	var agentsResp agentsResponse
	if err := json.Unmarshal(agentsRaw, &agentsResp); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}

	nameFilter := map[string]struct{}{}
	for _, name := range models.SanitizeStringList(agentNames) {
		nameFilter[name] = struct{}{}
	}

	rows := make([]map[string]any, 0)
	for _, agent := range agentsResp.Data.AffectedItems {
		if len(nameFilter) > 0 {
			if _, ok := nameFilter[agent.Name]; !ok {
				continue
			}
		}

		raw, err := c.GetSCA(ctx, agent.ID)
		if err != nil {
			continue
		}

		var sca scaResponse
		if err := json.Unmarshal(raw, &sca); err != nil {
			continue
		}

		for _, policy := range sca.Data.AffectedItems {
			rows = append(rows, map[string]any{
				"agent_id":      agent.ID,
				"agent":         agent.Name,
				"policy_id":     policy.PolicyID,
				"policy":        policy.Name,
				"description":   policy.Description,
				"score":         policy.Score,
				"pass":          policy.Pass,
				"fail":          policy.Fail,
				"total_checks":  policy.TotalChecks,
				"end_scan":      policy.EndScan,
			})
		}
	}

	return json.Marshal(rows)
}

func ParseSCALiveTableFrame(raw []byte, refID string) (*data.Frame, error) {
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse sca rows: %w", err)
	}

	agents := make([]string, 0, len(rows))
	agentIDs := make([]string, 0, len(rows))
	policyIDs := make([]string, 0, len(rows))
	policies := make([]string, 0, len(rows))
	descriptions := make([]string, 0, len(rows))
	scores := make([]float64, 0, len(rows))
	passed := make([]int64, 0, len(rows))
	failed := make([]int64, 0, len(rows))
	totalChecks := make([]int64, 0, len(rows))
	endScans := make([]string, 0, len(rows))

	for _, row := range rows {
		agents = append(agents, fmt.Sprintf("%v", row["agent"]))
		agentIDs = append(agentIDs, fmt.Sprintf("%v", row["agent_id"]))
		policyIDs = append(policyIDs, fmt.Sprintf("%v", row["policy_id"]))
		policies = append(policies, fmt.Sprintf("%v", row["policy"]))
		descriptions = append(descriptions, fmt.Sprintf("%v", row["description"]))
		scores = append(scores, toFloat64(row["score"]))
		passed = append(passed, toInt64(row["pass"]))
		failed = append(failed, toInt64(row["fail"]))
		totalChecks = append(totalChecks, toInt64(row["total_checks"]))
		endScans = append(endScans, fmt.Sprintf("%v", row["end_scan"]))
	}

	frame := data.NewFrame("sca")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("agent", nil, agents),
		data.NewField("agent_id", nil, agentIDs),
		data.NewField("policy_id", nil, policyIDs),
		data.NewField("policy", nil, policies),
		data.NewField("description", nil, descriptions),
		data.NewField("score", nil, scores),
		data.NewField("pass", nil, passed),
		data.NewField("fail", nil, failed),
		data.NewField("total_checks", nil, totalChecks),
		data.NewField("end_scan", nil, endScans),
	)

	return frame, nil
}

func toFloat64(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
