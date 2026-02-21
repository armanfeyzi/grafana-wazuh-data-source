package wazuhapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func TestClientPing(t *testing.T) {
	t.Parallel()

	const token = "test-jwt-token"
	authCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/security/user/authenticate":
			authCalls++
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user != "admin" || pass != "secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(token))
		case "/agents":
			if r.Header.Get("Authorization") != "Bearer "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"affected_items":[],"total_affected_items":0}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(testSettings(server.URL), httpclient.New(true))
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if authCalls != 1 {
		t.Fatalf("expected one auth call, got %d", authCalls)
	}
}

func TestClientPingRefreshesTokenOn401(t *testing.T) {
	t.Parallel()

	const firstToken = "expired-token"
	const secondToken = "fresh-token"
	agentCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/security/user/authenticate":
			_, _ = w.Write([]byte(secondToken))
		case "/agents":
			agentCalls++
			auth := r.Header.Get("Authorization")
			if agentCalls == 1 && auth == "Bearer "+firstToken {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if strings.HasSuffix(auth, secondToken) {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(testSettings(server.URL), httpclient.New(true))
	client.cachedToken = firstToken
	client.tokenExpiry = time.Now().Add(time.Hour)

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if agentCalls != 2 {
		t.Fatalf("expected two agent calls, got %d", agentCalls)
	}
}

func TestAuthenticateFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(testSettings(server.URL), httpclient.New(true))
	err := client.Ping(context.Background())
	if err == nil || !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected authentication error, got %v", err)
	}
}

func testSettings(baseURL string) *models.PluginSettings {
	return &models.PluginSettings{
		ManagerURL: baseURL,
		Username:   "admin",
		Secrets:    &models.SecretPluginSettings{Password: "secret"},
	}
}
