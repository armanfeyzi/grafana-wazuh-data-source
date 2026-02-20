package indexer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func TestBuildAlertsTimeSeriesQuery(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	minLevel := 5

	body, err := BuildAlertsTimeSeriesQuery(alertQueryParams{
		From:          from,
		To:            to,
		MaxDataPoints: 60,
		Filters: models.QueryFilters{
			AgentNames:   []string{"ubuntu"},
			RuleLevelMin: &minLevel,
			RuleGroups:   []string{"sshd"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed["size"].(float64) != 0 {
		t.Fatalf("expected size 0, got %v", parsed["size"])
	}

	filters := parsed["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	if len(filters) != 4 {
		t.Fatalf("expected 4 filters, got %d", len(filters))
	}
}

func TestParseAlertsTimeSeriesFrame(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"aggregations": {
			"histogram": {
				"buckets": [
					{"key": 1779374400000, "doc_count": 3},
					{"key": 1779374460000, "doc_count": 7}
				]
			}
		}
	}`)

	frame, err := ParseAlertsTimeSeriesFrame(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frame.Fields[1].Len() != 2 {
		t.Fatalf("expected 2 points, got %d", frame.Fields[1].Len())
	}
}

func TestParseAlertsTableFrames(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"hits": {
			"hits": [
				{
					"_source": {
						"@timestamp": "2026-05-21T10:00:00.000Z",
						"agent": {"name": "ubuntu", "id": "001"},
						"rule": {"id": "5710", "level": 5, "description": "SSH login", "groups": ["sshd", "authentication_failed"]}
					}
				}
			]
		}
	}`)

	frames, err := ParseAlertsTableFrames(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frames[0].Fields[1].At(0) != "ubuntu" {
		t.Fatalf("unexpected agent: %v", frames[0].Fields[1].At(0))
	}
}

func TestParseAlertsStatFrame(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"hits":{"total":{"value":42}}}`)
	frame, err := ParseAlertsStatFrame(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frame.Fields[0].At(0) != int64(42) {
		t.Fatalf("unexpected stat value: %v", frame.Fields[0].At(0))
	}
}

func TestFixedInterval(t *testing.T) {
	t.Parallel()

	from := time.Unix(0, 0)
	to := from.Add(time.Hour)
	if got := fixedInterval(from, to, 60); got != "1m" {
		t.Fatalf("expected 1m, got %s", got)
	}
}
