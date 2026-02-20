package indexer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func TestBuildFIMTableQuery(t *testing.T) {
	t.Parallel()

	body, err := BuildFIMTableQuery(queryParams{
		From:  time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC),
		To:    time.Date(2026, 5, 21, 1, 0, 0, 0, time.UTC),
		Limit: 50,
		Filters: models.QueryFilters{
			AgentNames: []string{"fedora"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatal(err)
	}

	filters := parsed["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	if len(filters) != 3 {
		t.Fatalf("expected 3 filters including syscheck, got %d", len(filters))
	}
}

func TestParseVulnerabilitiesTableFrames(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"hits": {
			"hits": [{
				"_source": {
					"agent": {"name": "fedora", "id": "001"},
					"package": {"name": "openssl", "version": "3.0.0"},
					"vulnerability": {
						"id": "CVE-2024-0001",
						"severity": "High",
						"description": "Example",
						"detected_at": "2026-05-21T10:00:00Z"
					}
				}
			}]
		}
	}`)

	frames, err := ParseVulnerabilitiesTableFrames(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frames[0].Fields[4].At(0) != "CVE-2024-0001" {
		t.Fatalf("unexpected cve: %v", frames[0].Fields[4].At(0))
	}
}

func TestParseFIMTableFrames(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"hits": {
			"hits": [{
				"_source": {
					"@timestamp": "2026-05-21T10:00:00.000Z",
					"agent": {"name": "fedora", "id": "001"},
					"syscheck": {"path": "/etc/passwd", "event": "modified", "uname_after": "root"},
					"rule": {"description": "File modified"}
				}
			}]
		}
	}`)

	frames, err := ParseFIMTableFrames(raw, "A")
	if err != nil {
		t.Fatal(err)
	}
	if frames[0].Fields[3].At(0) != "/etc/passwd" {
		t.Fatalf("unexpected path: %v", frames[0].Fields[3].At(0))
	}
}
