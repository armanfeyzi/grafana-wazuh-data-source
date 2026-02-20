package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData(t *testing.T) {
	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"aggregations":{"histogram":{"buckets":[{"key":1779374400000,"doc_count":2}]}}}`))
	}))
	defer indexerServer.Close()

	ds := newTestDatasource(t, "http://manager", indexerServer.URL)

	from := time.Now().Add(-time.Hour)
	to := time.Now()

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{
					RefID:         "A",
					JSON:          []byte(`{"dataType":"alerts","format":"time_series"}`),
					TimeRange:     backend.TimeRange{From: from, To: to},
					MaxDataPoints: 100,
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
	if resp.Responses["A"].Error != nil {
		t.Fatalf("unexpected query error: %v", resp.Responses["A"].Error)
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
