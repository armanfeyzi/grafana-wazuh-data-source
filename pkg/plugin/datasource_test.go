package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData(t *testing.T) {
	ds := newTestDatasource(t, "http://manager", "http://indexer")

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{
					RefID: "A",
					JSON:  []byte(`{"dataType":"alerts","format":"time_series"}`),
				},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestNewDatasource(t *testing.T) {
	jsonData, err := json.Marshal(map[string]any{
		"managerUrl": "https://manager.example.com:55000",
		"indexerUrl": "https://indexer.example.com:9200",
		"username":   "admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	instance, err := NewDatasource(context.Background(), backend.DataSourceInstanceSettings{
		JSONData: jsonData,
		DecryptedSecureJSONData: map[string]string{
			"password": "secret",
		},
	})
	if err != nil {
		t.Fatalf("NewDatasource() error = %v", err)
	}
	if instance == nil {
		t.Fatal("expected datasource instance")
	}
}
