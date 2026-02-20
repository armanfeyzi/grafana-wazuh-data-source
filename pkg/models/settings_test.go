package models

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestLoadPluginSettings(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{
			"managerUrl": "https://wazuh.example.com:55000",
			"indexerUrl": "https://indexer.example.com:9200",
			"username": "admin"
		}`),
		DecryptedSecureJSONData: map[string]string{
			"password": "secret",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if settings.ManagerURL != "https://wazuh.example.com:55000" {
		t.Fatalf("unexpected manager URL: %s", settings.ManagerURL)
	}
	if settings.Secrets.Password != "secret" {
		t.Fatalf("unexpected password: %s", settings.Secrets.Password)
	}
	if settings.IndexerUser() != "admin" {
		t.Fatalf("expected fallback indexer user admin, got %s", settings.IndexerUser())
	}
}

func TestIndexerCredentialsOverride(t *testing.T) {
	settings, err := LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{
			"username": "wazuh-wui",
			"indexerUsername": "admin"
		}`),
		DecryptedSecureJSONData: map[string]string{
			"password":        "api-pass",
			"indexerPassword": "indexer-pass",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if settings.IndexerUser() != "admin" {
		t.Fatalf("unexpected indexer user: %s", settings.IndexerUser())
	}
	if settings.IndexerPass() != "indexer-pass" {
		t.Fatalf("unexpected indexer password: %s", settings.IndexerPass())
	}
}

func TestQueryUnmarshal(t *testing.T) {
	var q Query
	if err := json.Unmarshal([]byte(`{
		"dataType": "alerts",
		"format": "time_series",
		"limit": 50
	}`), &q); err != nil {
		t.Fatal(err)
	}

	if q.DataType != DataTypeAlerts {
		t.Fatalf("unexpected data type: %s", q.DataType)
	}
	if q.Format != QueryFormatTimeSeries {
		t.Fatalf("unexpected format: %s", q.Format)
	}
	if q.Limit != 50 {
		t.Fatalf("unexpected limit: %d", q.Limit)
	}
}
