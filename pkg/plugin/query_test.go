package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

func TestExecuteQueryAlertsTimeSeries(t *testing.T) {
	t.Parallel()

	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wazuh-alerts-*/_search" && r.URL.Path != "/wazuh-alerts-4.x-2026.05.21/_search" {
			t.Logf("accepting generic path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"aggregations": {
				"histogram": {
					"buckets": [{"key": 1779374400000, "doc_count": 5}]
				}
			}
		}`))
	}))
	defer indexerServer.Close()

	hc, err := httpclient.NewTest(true)
	if err != nil {
		t.Fatal(err)
	}
	ds := &Datasource{
		settings: &models.PluginSettings{IndexerURL: indexerServer.URL},
		indexer:  indexer.NewClient(&models.PluginSettings{IndexerURL: indexerServer.URL, Username: "admin", Secrets: &models.SecretPluginSettings{Password: "x"}}, hc),
	}

	from := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	queryJSON, _ := json.Marshal(models.Query{
		DataType: models.DataTypeAlerts,
		Format:   models.QueryFormatTimeSeries,
	})

	resp, err := ds.executeQuery(context.Background(), "A", models.Query{
		DataType: models.DataTypeAlerts,
		Format:   models.QueryFormatTimeSeries,
	}, backend.DataQuery{
		RefID: "A",
		JSON:  queryJSON,
		TimeRange: backend.TimeRange{From: from, To: to},
		MaxDataPoints: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Frames) != 1 || resp.Frames[0].Fields[1].Len() != 1 {
		t.Fatalf("unexpected frames: %+v", resp.Frames)
	}
}

func TestExecuteQueryAgents(t *testing.T) {
	t.Parallel()

	manager := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/security/user/authenticate":
			_, _ = w.Write([]byte("token"))
		case "/agents":
			_, _ = w.Write([]byte(`{"data":{"affected_items":[{"id":"001","name":"ubuntu","status":"active","ip":"10.0.0.1","version":"v4.8.0","lastKeepAlive":"now","os":{"name":"Ubuntu"}}]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer manager.Close()

	settings := &models.PluginSettings{
		ManagerURL: manager.URL,
		Username:   "admin",
		Secrets:    &models.SecretPluginSettings{Password: "secret"},
	}
	hc, err := httpclient.NewTest(true)
	if err != nil {
		t.Fatal(err)
	}
	ds := &Datasource{
		settings: settings,
		wazuhAPI: wazuhapi.NewClient(settings, hc),
	}

	resp, err := ds.executeQuery(context.Background(), "A", models.Query{DataType: models.DataTypeAgents}, backend.DataQuery{RefID: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Frames[0].Fields[1].At(0) != "ubuntu" {
		t.Fatalf("unexpected agent: %v", resp.Frames[0].Fields[1].At(0))
	}
}

func TestExecuteQueryUnsupportedType(t *testing.T) {
	t.Parallel()

	ds := &Datasource{}
	_, err := ds.executeQuery(context.Background(), "A", models.Query{DataType: models.DataType("unknown")}, backend.DataQuery{})
	if err == nil {
		t.Fatal("expected error for unsupported data type")
	}
}
