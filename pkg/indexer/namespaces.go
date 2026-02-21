package indexer

import (
	"encoding/json"
	"fmt"
	"time"
)

// BuildNamespacesQuery returns an OpenSearch terms-aggregation query that
// collects distinct kubernetes.namespace values from the last 7 days.
// It tries both the nested field path used by the Wazuh k8s integration
// (kubernetes.namespace) and the flat path used by some Falco integrations
// (data.kubernetes.namespace). Whichever has data wins; the parser reads
// either location.
func BuildNamespacesQuery(limit int) ([]byte, error) {
	if limit <= 0 {
		limit = 200
	}

	to := time.Now().UTC()
	from := to.Add(-7 * 24 * time.Hour)

	body := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					buildTimeRangeFilter(from, to, "@timestamp"),
					{
						"bool": map[string]any{
							"should": []map[string]any{
								{
									"exists": map[string]any{
										"field": "kubernetes.namespace",
									},
								},
								{
									"exists": map[string]any{
										"field": "data.kubernetes.namespace",
									},
								},
							},
							"minimum_should_match": 1,
						},
					},
				},
			},
		},
		"aggs": map[string]any{
			// Primary path: Wazuh k8s integration / Falco integration
			"namespaces": map[string]any{
				"terms": map[string]any{
					"field": "kubernetes.namespace.keyword",
					"size":  limit,
					"order": map[string]any{"_key": "asc"},
				},
			},
			// Fallback path: some Wazuh modules nest under data.*
			"namespaces_data": map[string]any{
				"terms": map[string]any{
					"field": "data.kubernetes.namespace.keyword",
					"size":  limit,
					"order": map[string]any{"_key": "asc"},
				},
			},
		},
	}

	return json.Marshal(body)
}

type namespacesResponse struct {
	Aggregations struct {
		Namespaces struct {
			Buckets []struct {
				Key string `json:"key"`
			} `json:"buckets"`
		} `json:"namespaces"`
		NamespacesData struct {
			Buckets []struct {
				Key string `json:"key"`
			} `json:"buckets"`
		} `json:"namespaces_data"`
	} `json:"aggregations"`
}

// ParseNamespacesResponse extracts the sorted, deduplicated list of namespace
// names from the aggregation response. Returns an empty slice (not an error)
// when the kubernetes.namespace field doesn't exist in the index.
func ParseNamespacesResponse(raw []byte) ([]string, error) {
	var resp namespacesResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse namespaces response: %w", err)
	}

	seen := make(map[string]struct{})
	var result []string

	for _, b := range resp.Aggregations.Namespaces.Buckets {
		if b.Key == "" {
			continue
		}
		if _, ok := seen[b.Key]; !ok {
			seen[b.Key] = struct{}{}
			result = append(result, b.Key)
		}
	}

	// Merge in the fallback agg — deduplicate across both paths.
	for _, b := range resp.Aggregations.NamespacesData.Buckets {
		if b.Key == "" {
			continue
		}
		if _, ok := seen[b.Key]; !ok {
			seen[b.Key] = struct{}{}
			result = append(result, b.Key)
		}
	}

	return result, nil
}
