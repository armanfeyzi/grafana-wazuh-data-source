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
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

func TestCallResourceAgents(t *testing.T) {
	t.Parallel()

	manager := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/security/user/authenticate":
			_, _ = w.Write([]byte("token"))
		case "/agents":
			_, _ = w.Write([]byte(`{"data":{"affected_items":[{"id":"001","name":"fedora"}]}}`))
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

	var response *backend.CallResourceResponse
	err = ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Method: http.MethodGet,
		Path:   "agents",
	}, &testResourceSender{fn: func(resp *backend.CallResourceResponse) error {
		response = resp
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", response.Status, response.Body)
	}

	var options []wazuhapi.AgentOption
	if err := json.Unmarshal(response.Body, &options); err != nil {
		t.Fatal(err)
	}
	if len(options) != 1 || options[0].Value != "fedora" {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestCallResourceAgents_upstreamErrorSanitized(t *testing.T) {
	t.Parallel()

	manager := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/security/user/authenticate" {
			_, _ = w.Write([]byte("token"))
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"internal failure at 10.0.0.5:55000"}`))
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

	var response *backend.CallResourceResponse
	_ = ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Method: http.MethodGet,
		Path:   "agents",
	}, &testResourceSender{fn: func(resp *backend.CallResourceResponse) error {
		response = resp
		return nil
	}})

	if response.Status != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", response.Status)
	}
	body := string(response.Body)
	if strings.Contains(body, "10.0.0.5") {
		t.Fatalf("upstream IP leaked in response: %s", body)
	}
	if !strings.Contains(body, "unexpected") && !strings.Contains(body, "cannot connect") {
		t.Fatalf("expected sanitized message, got %s", body)
	}
}

type testResourceSender struct {
	fn func(*backend.CallResourceResponse) error
}

func (s *testResourceSender) Send(resp *backend.CallResourceResponse) error {
	return s.fn(resp)
}

func (s *testResourceSender) SendStatus(status int, body []byte) error {
	return s.fn(&backend.CallResourceResponse{Status: status, Body: body})
}
