package indexer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func buildFIMFilters(from, to time.Time, f models.QueryFilters) []map[string]any {
	filters := buildAlertFilters(from, to, f)
	return appendRuleGroupFilter(filters, "syscheck")
}

func BuildFIMTimeSeriesQuery(p queryParams) ([]byte, error) {
	filters := buildFIMFilters(p.From, p.To, p.Filters)
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

func BuildFIMTableQuery(p queryParams) ([]byte, error) {
	limit := clampLimit(p.Limit)
	filters := buildFIMFilters(p.From, p.To, p.Filters)

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
			"syscheck.path",
			"syscheck.event",
			"syscheck.mode",
			"syscheck.uname_after",
			"syscheck.uid_after",
			"rule.description",
		},
	}

	return json.Marshal(body)
}

func BuildFIMStatQuery(p queryParams) ([]byte, error) {
	filters := buildFIMFilters(p.From, p.To, p.Filters)
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

func ParseFIMTimeSeriesFrame(raw []byte, refID string) (*data.Frame, error) {
	return parseHistogramFrame(raw, refID, "fim", "FIM events")
}

func ParseFIMTableFrames(raw []byte, refID string) ([]*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	times := make([]time.Time, 0, len(resp.Hits.Hits))
	agents := make([]string, 0, len(resp.Hits.Hits))
	agentIDs := make([]string, 0, len(resp.Hits.Hits))
	paths := make([]string, 0, len(resp.Hits.Hits))
	events := make([]string, 0, len(resp.Hits.Hits))
	users := make([]string, 0, len(resp.Hits.Hits))
	descriptions := make([]string, 0, len(resp.Hits.Hits))

	for _, hit := range resp.Hits.Hits {
		src := hit.Source
		times = append(times, parseTimestamp(src["@timestamp"]))
		agents = append(agents, nestedString(src, "agent", "name"))
		agentIDs = append(agentIDs, nestedString(src, "agent", "id"))
		paths = append(paths, nestedString(src, "syscheck", "path"))
		event := nestedString(src, "syscheck", "event")
		if event == "" {
			event = nestedString(src, "syscheck", "mode")
		}
		events = append(events, event)
		user := nestedString(src, "syscheck", "uname_after")
		if user == "" {
			user = nestedString(src, "syscheck", "uid_after")
		}
		users = append(users, user)
		descriptions = append(descriptions, nestedString(src, "rule", "description"))
	}

	frame := data.NewFrame("fim")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("Time", nil, times),
		data.NewField("agent", nil, agents),
		data.NewField("agent_id", nil, agentIDs),
		data.NewField("path", nil, paths),
		data.NewField("event", nil, events),
		data.NewField("user", nil, users),
		data.NewField("description", nil, descriptions),
	)

	return []*data.Frame{frame}, nil
}

func ParseFIMStatFrame(raw []byte, refID string) (*data.Frame, error) {
	return parseTotalStatFrame(raw, refID, "fim")
}
