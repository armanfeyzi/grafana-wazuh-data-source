package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

func TestCheckHealthSuccess(t *testing.T) {
	t.Parallel()

	manager := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/security/user/authenticate":
			_, _ = w.Write([]byte("token"))
		case "/agents":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer manager.Close()

	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_cluster/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer indexerServer.Close()

	ds := newTestDatasource(t, manager.URL, indexerServer.URL)
	result, err := ds.CheckHealth(context.Background(), testHealthRequest(t, manager.URL, indexerServer.URL))
	if err != nil {
		t.Fatalf("CheckHealth() error = %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Fatalf("expected ok status, got %v message=%q", result.Status, result.Message)
	}
}

func TestCheckHealthManagerFailure(t *testing.T) {
	t.Parallel()

	indexerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer indexerServer.Close()

	ds := newTestDatasource(t, "http://127.0.0.1:1", indexerServer.URL)
	result, err := ds.CheckHealth(context.Background(), testHealthRequest(t, "http://127.0.0.1:1", indexerServer.URL))
	if err != nil {
		t.Fatalf("CheckHealth() error = %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Fatalf("expected error status, got %v", result.Status)
	}
	if !strings.Contains(result.Message, "manager API") {
		t.Fatalf("expected manager API error, got %q", result.Message)
	}
}

func TestCheckHealthMissingPassword(t *testing.T) {
	t.Parallel()

	ds := newTestDatasource(t, "https://manager.example.com:55000", "https://indexer.example.com:9200")
	req := testHealthRequest(t, "https://manager.example.com:55000", "https://indexer.example.com:9200")
	req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData = map[string]string{}

	result, err := ds.CheckHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckHealth() error = %v", err)
	}
	if result.Status != backend.HealthStatusError || !strings.Contains(result.Message, "password is required") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func newTestDatasource(t *testing.T, managerURL, indexerURL string) *Datasource {
	t.Helper()

	settings := &models.PluginSettings{
		ManagerURL: managerURL,
		IndexerURL: indexerURL,
		Username:   "admin",
		Secrets:    &models.SecretPluginSettings{Password: "secret"},
	}
	hc, err := httpclient.NewTest(true)
	if err != nil {
		t.Fatal(err)
	}

	return &Datasource{
		settings: settings,
		wazuhAPI: wazuhapi.NewClient(settings, hc),
		indexer:  indexer.NewClient(settings, hc),
	}
}

func testHealthRequest(t *testing.T, managerURL, indexerURL string) *backend.CheckHealthRequest {
	t.Helper()

	jsonData, err := json.Marshal(map[string]any{
		"managerUrl": managerURL,
		"indexerUrl": indexerURL,
		"username":   "admin",
	})
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}

	return &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData: jsonData,
				DecryptedSecureJSONData: map[string]string{
					"password": "secret",
				},
			},
		},
	}
}
