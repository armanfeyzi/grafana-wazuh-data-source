package indexer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func buildSCAHistoryFilters(from, to time.Time, f models.QueryFilters) []map[string]any {
	filters := buildAlertFilters(from, to, f)
	return appendRuleGroupFilter(filters, "sca")
}

func BuildSCAHistoryTimeSeriesQuery(p queryParams) ([]byte, error) {
	filters := buildSCAHistoryFilters(p.From, p.To, p.Filters)
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
				"aggs": map[string]any{
					"avg_score": map[string]any{
						"avg": map[string]any{
							"field": "data.sca.score",
						},
					},
				},
			},
		},
	}

	return json.Marshal(body)
}

func BuildSCAHistoryTableQuery(p queryParams) ([]byte, error) {
	limit := clampLimit(p.Limit)
	filters := buildSCAHistoryFilters(p.From, p.To, p.Filters)

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
			"data.sca.policy",
			"data.sca.score",
			"data.sca.pass",
			"data.sca.fail",
			"data.sca.total_checks",
			"rule.description",
		},
	}

	return json.Marshal(body)
}

func BuildSCAHistoryStatQuery(p queryParams) ([]byte, error) {
	filters := buildSCAHistoryFilters(p.From, p.To, p.Filters)
	body := map[string]any{
		"size":             0,
		"track_total_hits": true,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
	}

	return json.Marshal(body)
}

type scaHistoryResponse struct {
	searchResponse
	Aggregations struct {
		Histogram struct {
			Buckets []struct {
				Key         int64  `json:"key"`
				KeyAsString string `json:"key_as_string"`
				DocCount    int64  `json:"doc_count"`
				AvgScore    struct {
					Value *float64 `json:"value"`
				} `json:"avg_score"`
			} `json:"buckets"`
		} `json:"histogram"`
	} `json:"aggregations"`
}

func ParseSCAHistoryTimeSeriesFrame(raw []byte, refID string) (*data.Frame, error) {
	var resp scaHistoryResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	times := make([]time.Time, 0, len(resp.Aggregations.Histogram.Buckets))
	scores := make([]float64, 0, len(resp.Aggregations.Histogram.Buckets))
	counts := make([]int64, 0, len(resp.Aggregations.Histogram.Buckets))

	for _, bucket := range resp.Aggregations.Histogram.Buckets {
		ts := time.UnixMilli(bucket.Key).UTC()
		if bucket.KeyAsString != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, bucket.KeyAsString); err == nil {
				ts = parsed.UTC()
			}
		}
		times = append(times, ts)
		counts = append(counts, bucket.DocCount)
		if bucket.AvgScore.Value != nil {
			scores = append(scores, *bucket.AvgScore.Value)
		} else {
			scores = append(scores, 0)
		}
	}

	frame := data.NewFrame("sca")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField("Avg score", nil, scores),
		data.NewField("Scans", nil, counts),
	)
	frame.SetMeta(&data.FrameMeta{Type: data.FrameTypeTimeSeriesMulti})

	return frame, nil
}

func ParseSCAHistoryTableFrames(raw []byte, refID string) ([]*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	times := make([]time.Time, 0, len(resp.Hits.Hits))
	agents := make([]string, 0, len(resp.Hits.Hits))
	agentIDs := make([]string, 0, len(resp.Hits.Hits))
	policies := make([]string, 0, len(resp.Hits.Hits))
	scores := make([]float64, 0, len(resp.Hits.Hits))
	passed := make([]int64, 0, len(resp.Hits.Hits))
	failed := make([]int64, 0, len(resp.Hits.Hits))
	totalChecks := make([]int64, 0, len(resp.Hits.Hits))

	for _, hit := range resp.Hits.Hits {
		src := hit.Source
		times = append(times, parseTimestamp(src["@timestamp"]))
		agents = append(agents, nestedString(src, "agent", "name"))
		agentIDs = append(agentIDs, nestedString(src, "agent", "id"))
		policies = append(policies, nestedString(src, "data", "sca", "policy"))
		scores = append(scores, nestedFloat64(src, "data", "sca", "score"))
		passed = append(passed, nestedInt64(src, "data", "sca", "pass"))
		failed = append(failed, nestedInt64(src, "data", "sca", "fail"))
		totalChecks = append(totalChecks, nestedInt64(src, "data", "sca", "total_checks"))
	}

	frame := data.NewFrame("sca")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField("agent", nil, agents),
		data.NewField("agent_id", nil, agentIDs),
		data.NewField("policy", nil, policies),
		data.NewField("score", nil, scores),
		data.NewField("pass", nil, passed),
		data.NewField("fail", nil, failed),
		data.NewField("total_checks", nil, totalChecks),
	)

	return []*data.Frame{frame}, nil
}

func ParseSCAHistoryStatFrame(raw []byte, refID string) (*data.Frame, error) {
	return parseTotalStatFrame(raw, refID, "sca")
}
