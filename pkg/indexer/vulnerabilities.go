package indexer

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func BuildVulnerabilitiesTableQuery(p queryParams) ([]byte, error) {
	limit := clampLimit(p.Limit)
	filters := buildVulnerabilityFilters(p.Filters)

	body := map[string]any{
		"size": limit,
		"sort": []map[string]any{
			{"vulnerability.detected_at": map[string]any{"order": "desc", "unmapped_type": "date"}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"_source": []string{
			"agent.name",
			"agent.id",
			"package.name",
			"package.version",
			"vulnerability.id",
			"vulnerability.severity",
			"vulnerability.description",
			"vulnerability.detected_at",
		},
	}

	return json.Marshal(body)
}

func BuildVulnerabilitiesStatQuery(p queryParams) ([]byte, error) {
	filters := buildVulnerabilityFilters(p.Filters)
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

func BuildVulnerabilitiesSeverityTableQuery(p queryParams) ([]byte, error) {
	filters := buildVulnerabilityFilters(p.Filters)
	body := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
		"aggs": map[string]any{
			"severity": map[string]any{
				"terms": map[string]any{
					"field": "vulnerability.severity",
					"size":  10,
				},
			},
		},
	}

	return json.Marshal(body)
}

func ParseVulnerabilitiesTableFrames(raw []byte, refID string) ([]*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	agents := make([]string, 0, len(resp.Hits.Hits))
	agentIDs := make([]string, 0, len(resp.Hits.Hits))
	packages := make([]string, 0, len(resp.Hits.Hits))
	versions := make([]string, 0, len(resp.Hits.Hits))
	cves := make([]string, 0, len(resp.Hits.Hits))
	severities := make([]string, 0, len(resp.Hits.Hits))
	descriptions := make([]string, 0, len(resp.Hits.Hits))
	detectedAt := make([]string, 0, len(resp.Hits.Hits))

	for _, hit := range resp.Hits.Hits {
		src := hit.Source
		agents = append(agents, nestedString(src, "agent", "name"))
		agentIDs = append(agentIDs, nestedString(src, "agent", "id"))
		packages = append(packages, nestedString(src, "package", "name"))
		versions = append(versions, nestedString(src, "package", "version"))
		cves = append(cves, nestedString(src, "vulnerability", "id"))
		severities = append(severities, nestedString(src, "vulnerability", "severity"))
		descriptions = append(descriptions, nestedString(src, "vulnerability", "description"))
		detectedAt = append(detectedAt, nestedString(src, "vulnerability", "detected_at"))
	}

	frame := data.NewFrame("vulnerabilities")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("agent", nil, agents),
		data.NewField("agent_id", nil, agentIDs),
		data.NewField("package", nil, packages),
		data.NewField("package_version", nil, versions),
		data.NewField("cve", nil, cves),
		data.NewField("severity", nil, severities),
		data.NewField("description", nil, descriptions),
		data.NewField("detected_at", nil, detectedAt),
	)

	return []*data.Frame{frame}, nil
}

func ParseVulnerabilitiesStatFrame(raw []byte, refID string) (*data.Frame, error) {
	return parseTotalStatFrame(raw, refID, "vulnerabilities")
}

func ParseVulnerabilitiesSeverityTableFrames(raw []byte, refID string) ([]*data.Frame, error) {
	var resp searchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	severities := make([]string, 0, len(resp.Aggregations.Severity.Buckets))
	counts := make([]int64, 0, len(resp.Aggregations.Severity.Buckets))
	for _, bucket := range resp.Aggregations.Severity.Buckets {
		severities = append(severities, bucket.Key)
		counts = append(counts, bucket.DocCount)
	}

	frame := data.NewFrame("vulnerabilities")
	frame.RefID = refID
	frame.Fields = append(frame.Fields,
		data.NewField("severity", nil, severities),
		data.NewField("count", nil, counts),
	)

	return []*data.Frame{frame}, nil
}
