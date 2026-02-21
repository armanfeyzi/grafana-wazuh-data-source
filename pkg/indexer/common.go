package indexer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

// MaxResponseBytes caps the OpenSearch response body read to protect memory.
const MaxResponseBytes = 32 << 20 // 32 MB

type queryParams struct {
	From          time.Time
	To            time.Time
	MaxDataPoints int64
	Limit         int
	Filters       models.QueryFilters
}

// QueryParams is the shared indexer query input.
type QueryParams = queryParams

// AlertQueryParamsFrom builds params from a Grafana data query.
func AlertQueryParamsFrom(q backend.DataQuery, query models.Query) queryParams {
	return queryParams{
		From:          q.TimeRange.From,
		To:            q.TimeRange.To,
		MaxDataPoints: q.MaxDataPoints,
		Limit:         query.Limit,
		Filters:       query.Filters,
	}
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func buildTimeRangeFilter(from, to time.Time, field string) map[string]any {
	return map[string]any{
		"range": map[string]any{
			field: map[string]any{
				"gte": from.UTC().Format(time.RFC3339),
				"lte": to.UTC().Format(time.RFC3339),
			},
		},
	}
}

func buildAgentNameFilter(names []string) map[string]any {
	return map[string]any{
		"terms": map[string]any{
			"agent.name": names,
		},
	}
}

func buildAlertFilters(from, to time.Time, f models.QueryFilters) []map[string]any {
	filters := []map[string]any{
		buildTimeRangeFilter(from, to, "@timestamp"),
	}

	if len(f.AgentNamesForQuery()) > 0 {
		filters = append(filters, buildAgentNameFilter(f.AgentNamesForQuery()))
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

func buildVulnerabilityFilters(f models.QueryFilters) []map[string]any {
	filters := make([]map[string]any, 0, 2)
	if len(f.AgentNamesForQuery()) > 0 {
		filters = append(filters, buildAgentNameFilter(f.AgentNamesForQuery()))
	}
	if len(f.SeverityForQuery()) > 0 {
		filters = append(filters, map[string]any{
			"terms": map[string]any{
				"vulnerability.severity": f.SeverityForQuery(),
			},
		})
	}
	return filters
}

func appendRuleGroupFilter(filters []map[string]any, group string) []map[string]any {
	return append(filters, map[string]any{
		"term": map[string]any{
			"rule.groups": group,
		},
	})
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
		Severity struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"severity"`
		AvgScore struct {
			Value *float64 `json:"value"`
		} `json:"avg_score"`
	} `json:"aggregations"`
}

func parseHistogramFrame(raw []byte, refID, frameName, valueField string) (*data.Frame, error) {
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

	frame := data.NewFrame(frameName)
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField(valueField, nil, values),
	)
	frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})

	return frame, nil
}

func parseTotalStatFrame(raw []byte, refID, frameName string) (*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	frame := data.NewFrame(frameName)
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

func nestedFloat64(src map[string]any, keys ...string) float64 {
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
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
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
