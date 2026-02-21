package indexer

import (
	"encoding/json"
	"testing"
)

func TestBuildNamespacesQuery(t *testing.T) {
	raw, err := BuildNamespacesQuery(50)
	if err != nil {
		t.Fatalf("BuildNamespacesQuery: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must request zero hits (aggregation-only query).
	if size, ok := body["size"].(float64); !ok || size != 0 {
		t.Errorf("expected size=0, got %v", body["size"])
	}

	// Must include both aggregation buckets.
	aggs, ok := body["aggs"].(map[string]any)
	if !ok {
		t.Fatal("missing aggs field")
	}
	if _, ok := aggs["namespaces"]; !ok {
		t.Error("missing aggs.namespaces")
	}
	if _, ok := aggs["namespaces_data"]; !ok {
		t.Error("missing aggs.namespaces_data")
	}
}

func TestParseNamespacesResponse_populated(t *testing.T) {
	fixture := `{
		"aggregations": {
			"namespaces": {
				"buckets": [
					{"key": "default"},
					{"key": "monitoring"},
					{"key": "security"}
				]
			},
			"namespaces_data": {
				"buckets": [
					{"key": "kube-system"}
				]
			}
		}
	}`

	namespaces, err := ParseNamespacesResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("ParseNamespacesResponse: %v", err)
	}

	expected := map[string]bool{
		"default":    true,
		"monitoring": true,
		"security":   true,
		"kube-system": true,
	}

	if len(namespaces) != len(expected) {
		t.Errorf("expected %d namespaces, got %d: %v", len(expected), len(namespaces), namespaces)
	}

	for _, ns := range namespaces {
		if !expected[ns] {
			t.Errorf("unexpected namespace: %q", ns)
		}
	}
}

func TestParseNamespacesResponse_deduplication(t *testing.T) {
	// The same namespace appears in both aggs — should only appear once.
	fixture := `{
		"aggregations": {
			"namespaces": {
				"buckets": [
					{"key": "default"},
					{"key": "monitoring"}
				]
			},
			"namespaces_data": {
				"buckets": [
					{"key": "default"},
					{"key": "security"}
				]
			}
		}
	}`

	namespaces, err := ParseNamespacesResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("ParseNamespacesResponse: %v", err)
	}

	if len(namespaces) != 3 {
		t.Errorf("expected 3 deduplicated namespaces, got %d: %v", len(namespaces), namespaces)
	}
}

func TestParseNamespacesResponse_empty(t *testing.T) {
	// Non-k8s deployment — field not in index — both buckets are empty.
	fixture := `{
		"aggregations": {
			"namespaces": {"buckets": []},
			"namespaces_data": {"buckets": []}
		}
	}`

	namespaces, err := ParseNamespacesResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("ParseNamespacesResponse: %v", err)
	}

	if len(namespaces) != 0 {
		t.Errorf("expected empty slice, got %v", namespaces)
	}
}

func TestParseNamespacesResponse_skipsEmptyKeys(t *testing.T) {
	fixture := `{
		"aggregations": {
			"namespaces": {
				"buckets": [
					{"key": ""},
					{"key": "production"}
				]
			},
			"namespaces_data": {"buckets": []}
		}
	}`

	namespaces, err := ParseNamespacesResponse([]byte(fixture))
	if err != nil {
		t.Fatalf("ParseNamespacesResponse: %v", err)
	}

	if len(namespaces) != 1 || namespaces[0] != "production" {
		t.Errorf("expected [production], got %v", namespaces)
	}
}
