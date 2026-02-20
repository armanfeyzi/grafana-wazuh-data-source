package indexer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

type alertQueryParams struct {
	From         time.Time
	To           time.Time
	MaxDataPoints int64
	Limit        int
	Filters      models.QueryFilters
}

func BuildAlertsTimeSeriesQuery(p alertQueryParams) ([]byte, error) {
	filters := buildAlertFilters(p.From, p.To, p.Filters)
	interval := fixedInterval(p.From, p.To, p.MaxDataPoints)

	body := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"aggs": map[string]any{
			"histogram": map[string]any{
				"date_histogram": map[string]any{
					"field":          "@timestamp",
					"fixed_interval": interval,
					"min_doc_count":  0,
					"extended_bounds": map[string]any{
						"min": p.From.UTC().Format(time.RFC3339),
						"max": p.To.UTC().Format(time.RFC3339),
					},
				},
			},
		},
	}

	return json.Marshal(body)
}

func BuildAlertsTableQuery(p alertQueryParams) ([]byte, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	filters := buildAlertFilters(p.From, p.To, p.Filters)
	body := map[string]any{
		"size": limit,
		"sort": []map[string]any{
			{"@timestamp": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"_source": []string{
			"@timestamp",
			"agent.name",
			"agent.id",
			"rule.id",
			"rule.level",
			"rule.description",
			"rule.groups",
		},
	}

	return json.Marshal(body)
}

func BuildAlertsStatQuery(p alertQueryParams) ([]byte, error) {
	filters := buildAlertFilters(p.From, p.To, p.Filters)
	body := map[string]any{
		"size": 0,
		"track_total_hits": true,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
	}

	return json.Marshal(body)
}

func buildAlertFilters(from, to time.Time, f models.QueryFilters) []map[string]any {
	filters := []map[string]any{
		{
			"range": map[string]any{
				"@timestamp": map[string]any{
					"gte": from.UTC().Format(time.RFC3339),
					"lte": to.UTC().Format(time.RFC3339),
				},
			},
		},
	}

	if len(f.AgentNames) > 0 {
		filters = append(filters, map[string]any{
			"terms": map[string]any{
				"agent.name": f.AgentNames,
			},
		})
	}

	if f.RuleLevelMin != nil || f.RuleLevelMax != nil {
		levelRange := map[string]any{}
		if f.RuleLevelMin != nil {
			levelRange["gte"] = *f.RuleLevelMin
		}
		if f.RuleLevelMax != nil {
			levelRange["lte"] = *f.RuleLevelMax
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{
				"rule.level": levelRange,
			},
		})
	}

	if len(f.RuleGroups) > 0 {
		filters = append(filters, map[string]any{
			"terms": map[string]any{
				"rule.groups": f.RuleGroups,
			},
		})
	}

	return filters
}

func fixedInterval(from, to time.Time, maxPoints int64) string {
	if maxPoints < 1 {
		maxPoints = 300
	}
	duration := to.Sub(from)
	if duration <= 0 {
		return "1m"
	}
	interval := duration / time.Duration(maxPoints)
	if interval < time.Minute {
		seconds := int(interval.Seconds())
		if seconds < 1 {
			seconds = 1
		}
		return fmt.Sprintf("%ds", seconds)
	}
	if interval < time.Hour {
		minutes := int(interval.Minutes())
		if minutes < 1 {
			minutes = 1
		}
		return fmt.Sprintf("%dm", minutes)
	}
	hours := int(interval.Hours())
	if hours < 1 {
		hours = 1
	}
	return fmt.Sprintf("%dh", hours)
}

type searchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source map[string]any `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Histogram struct {
			Buckets []struct {
				Key         int64  `json:"key"`
				KeyAsString string `json:"key_as_string"`
				DocCount    int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"histogram"`
	} `json:"aggregations"`
}

func ParseAlertsTimeSeriesFrame(raw []byte, refID string) (*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	times := make([]time.Time, 0, len(resp.Aggregations.Histogram.Buckets))
	values := make([]int64, 0, len(resp.Aggregations.Histogram.Buckets))

	for _, bucket := range resp.Aggregations.Histogram.Buckets {
		ts := time.UnixMilli(bucket.Key).UTC()
		if bucket.KeyAsString != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, bucket.KeyAsString); err == nil {
				ts = parsed.UTC()
			}
		}
		times = append(times, ts)
		values = append(values, bucket.DocCount)
	}

	frame := data.NewFrame("alerts")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField("Alerts", nil, values),
	)
	frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})

	return frame, nil
}

func ParseAlertsTableFrames(raw []byte, refID string) ([]*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	times := make([]time.Time, 0, len(resp.Hits.Hits))
	agents := make([]string, 0, len(resp.Hits.Hits))
	agentIDs := make([]string, 0, len(resp.Hits.Hits))
	ruleIDs := make([]string, 0, len(resp.Hits.Hits))
	levels := make([]int64, 0, len(resp.Hits.Hits))
	descriptions := make([]string, 0, len(resp.Hits.Hits))
	groups := make([]string, 0, len(resp.Hits.Hits))

	for _, hit := range resp.Hits.Hits {
		src := hit.Source
		times = append(times, parseTimestamp(src["@timestamp"]))
		agents = append(agents, nestedString(src, "agent", "name"))
		agentIDs = append(agentIDs, nestedString(src, "agent", "id"))
		ruleIDs = append(ruleIDs, nestedString(src, "rule", "id"))
		levels = append(levels, nestedInt64(src, "rule", "level"))
		descriptions = append(descriptions, nestedString(src, "rule", "description"))
		groups = append(groups, joinGroups(nestedStringSlice(src, "rule", "groups")))
	}

	frame := data.NewFrame("alerts")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField("agent", nil, agents),
		data.NewField("agent_id", nil, agentIDs),
		data.NewField("rule_id", nil, ruleIDs),
		data.NewField("severity_level", nil, levels),
		data.NewField("rule_description", nil, descriptions),
		data.NewField("rule_groups", nil, groups),
	)

	return []*data.Frame{frame}, nil
}

func ParseAlertsStatFrame(raw []byte, refID string) (*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	frame := data.NewFrame("alerts")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("value", nil, []int64{resp.Hits.Total.Value}),
	)

	return frame, nil
}

func nestedString(src map[string]any, keys ...string) string {
	current := any(src)
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = obj[key]
	}
	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func nestedInt64(src map[string]any, keys ...string) int64 {
	current := any(src)
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return 0
		}
		current = obj[key]
	}
	switch v := current.(type) {
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

func nestedStringSlice(src map[string]any, keys ...string) []string {
	current := any(src)
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = obj[key]
	}
	items, ok := current.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%v", item))
	}
	return out
}

func joinGroups(groups []string) string {
	if len(groups) == 0 {
		return ""
	}
	result := groups[0]
	for i := 1; i < len(groups); i++ {
		result += ", " + groups[i]
	}
	return result
}

func parseTimestamp(value any) time.Time {
	switch v := value.(type) {
	case string:
		if ts, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return ts.UTC()
		}
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			return ts.UTC()
		}
	case float64:
		return time.UnixMilli(int64(v)).UTC()
	}
	return time.Time{}
}

// AlertQueryParamsFrom builds params from a Grafana data query.
func AlertQueryParamsFrom(q backend.DataQuery, query models.Query) alertQueryParams {
	return alertQueryParams{
		From:          q.TimeRange.From,
		To:            q.TimeRange.To,
		MaxDataPoints: q.MaxDataPoints,
		Limit:         query.Limit,
		Filters:       query.Filters,
	}
}
