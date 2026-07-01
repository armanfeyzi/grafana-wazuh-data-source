package indexer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func TestClientPing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_cluster/health" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"green"}`))
	}))
	defer server.Close()

	hc, err := httpclient.NewTest(true)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(testSettings(server.URL), hc)
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestClientPingAuthFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	hc, err := httpclient.NewTest(true)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(testSettings(server.URL), hc)
	err = client.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !models.IsWazuhError(err, models.ErrAuth) {
		t.Fatalf("expected ErrAuth, got %v", err)
	}
	if !strings.Contains(models.UserMessage(err), "authentication failed") {
		t.Fatalf("expected auth message, got %v", err)
	}
}

func testSettings(baseURL string) *models.PluginSettings {
	return &models.PluginSettings{
		IndexerURL: baseURL,
		Username:   "admin",
		Secrets:    &models.SecretPluginSettings{Password: "secret"},
	}
}
