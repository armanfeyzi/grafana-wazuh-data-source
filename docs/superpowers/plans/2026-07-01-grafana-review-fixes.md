# Grafana Catalog Review Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address all Grafana catalog review feedback (SDK HTTP client, error sanitization, Go CVE rebuild, catalog README, SDK update) and ship v0.2.9 for resubmission.

**Architecture:** Replace the custom HTTP transport with Grafana SDK `backend/httpclient` wired through a thin `pkg/httpclient` wrapper. Unify user-facing errors through existing `WazuhError` types and `models.UserMessage`, extracting shared indexer classifiers into `pkg/indexer/errors.go`. Bump Go 1.25.11 and SDK v0.292.2; rewrite `src/README.md` for catalog users.

**Tech Stack:** Go 1.25.11, grafana-plugin-sdk-go v0.292.2, Grafana plugin backend (Mage build), npm/webpack frontend (version bump only).

**Spec:** `docs/superpowers/specs/2026-07-01-grafana-review-fixes-design.md`

---

## File map

| File | Responsibility |
|---|---|
| `pkg/httpclient/client.go` | Production + test HTTP client factories using SDK |
| `pkg/plugin/datasource.go` | Wire SDK client in `NewDatasource`; safe error fallbacks |
| `pkg/indexer/errors.go` | **New** — shared HTTP/network error classifiers |
| `pkg/indexer/client.go` | Ping uses shared classifiers |
| `pkg/indexer/search.go` | Search uses shared classifiers (imports from errors.go) |
| `pkg/wazuhapi/client.go` | Remove body excerpts from default HTTP error branch |
| `pkg/models/errors.go` | Safe `UserMessage` fallback; `WazuhError.Error()` returns Message only |
| `pkg/plugin/resource.go` | Use `UserMessage` in CallResource responses |
| `src/README.md` | Catalog-user documentation |
| `go.mod` / `go.sum` | Go 1.25.11, SDK v0.292.2 |
| `.github/workflows/release.yml` | Pin Go 1.25.11 |
| `package.json` / `CHANGELOG.md` | Version 0.2.9 |

---

### Task 1: Bump Go toolchain and Grafana SDK

**Files:**
- Modify: `go.mod`
- Modify: `go.sum` (regenerated)
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Update go.mod**

```bash
cd /home/arman/Projects/personal/grafana-wazuh-data-source-plugin
go mod edit -go=1.25.11
go get github.com/grafana/grafana-plugin-sdk-go@v0.292.2
go mod tidy
```

Expected: `go.mod` contains `go 1.25.11` and `github.com/grafana/grafana-plugin-sdk-go v0.292.2`.

- [ ] **Step 2: Pin release workflow Go version**

In `.github/workflows/release.yml`, change both `go-version: '1.25.10'` lines to `go-version: '1.25.11'` (Setup Go step and package-plugin `with.go-version`).

- [ ] **Step 3: Verify compile**

```bash
go build ./...
```

Expected: exit 0, no errors.

---

### Task 2: Migrate HTTP client to Grafana SDK

**Files:**
- Modify: `pkg/httpclient/client.go`
- Modify: `pkg/plugin/datasource.go`
- Modify: `pkg/indexer/client_test.go`
- Modify: `pkg/wazuhapi/client_test.go`
- Modify: `pkg/plugin/health_test.go`
- Modify: `pkg/plugin/query_test.go`
- Modify: `pkg/plugin/resource_test.go`

- [ ] **Step 1: Replace `pkg/httpclient/client.go`**

```go
package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	sdkhttpclient "github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
)

const defaultTimeout = 30 * time.Second

// New creates an HTTP client using Grafana SDK options from datasource settings.
// This applies managed proxy, TLS, and transport settings from Grafana.
func New(ctx context.Context, settings backend.DataSourceInstanceSettings) (*http.Client, error) {
	opts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, err
	}
	if opts.Timeouts == nil {
		opts.Timeouts = &sdkhttpclient.TimeoutOptions{}
	}
	opts.Timeouts.Timeout = defaultTimeout
	return sdkhttpclient.New(opts)
}

// NewTest creates an HTTP client for unit tests (no Grafana context required).
func NewTest(skipTLSVerify bool) (*http.Client, error) {
	return sdkhttpclient.New(sdkhttpclient.Options{
		Timeouts: &sdkhttpclient.TimeoutOptions{Timeout: defaultTimeout},
		TLS: &sdkhttpclient.TLSOptions{
			InsecureSkipVerify: skipTLSVerify, // #nosec G402 -- test helper for httptest TLS
		},
	})
}
```

- [ ] **Step 2: Update `NewDatasource` in `pkg/plugin/datasource.go`**

Change signature usage from:

```go
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// ...
	httpClient := httpclient.New(config.TlsSkipVerify)
```

To:

```go
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// ...
	httpClient, err := httpclient.New(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("create HTTP client: %w", err)
	}
```

Add `"fmt"` to imports if not present.

- [ ] **Step 3: Update all test files using `httpclient.New(true)`**

Replace every `httpclient.New(true)` with a pattern that handles the error return:

```go
hc, err := httpclient.NewTest(true)
if err != nil {
    t.Fatal(err)
}
```

Files to update (grep confirms these):
- `pkg/indexer/client_test.go` (2 places)
- `pkg/wazuhapi/client_test.go` (3 places)
- `pkg/plugin/health_test.go` (1 place)
- `pkg/plugin/query_test.go` (1 place)
- `pkg/plugin/resource_test.go` (1 place)

- [ ] **Step 4: Run backend tests**

```bash
go test ./pkg/...
```

Expected: all PASS.

---

### Task 3: Safe error messages in models layer

**Files:**
- Modify: `pkg/models/errors.go`
- Modify: `pkg/models/errors_test.go`

- [ ] **Step 1: Write failing tests in `pkg/models/errors_test.go`**

Replace `TestUserMessage_plainError`:

```go
func TestUserMessage_plainError(t *testing.T) {
	err := fmt.Errorf("dial tcp 10.0.0.5:443: connection refused")
	if got := UserMessage(err); got != "An unexpected error occurred" {
		t.Errorf("expected generic message, got %q", got)
	}
}
```

Add test for `WazuhError.Error()` without cause leak:

```go
func TestWazuhError_Error_doesNotExposeCause(t *testing.T) {
	cause := fmt.Errorf("dial tcp 192.168.1.1:55000: connection refused")
	we := NewWazuhError(ErrUnreachable, "cannot connect — check the URL", cause)
	if got := we.Error(); got != "cannot connect — check the URL" {
		t.Errorf("expected message only, got %q", got)
	}
	if contains(we.Error(), "192.168") {
		t.Errorf("cause leaked into Error(): %q", we.Error())
	}
}
```

Update `TestWazuhError_Error_withCause` to expect Message only (no cause suffix):

```go
func TestWazuhError_Error_withCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	we := NewWazuhError(ErrUnreachable, "cannot reach indexer", cause)
	if we.Error() != "cannot reach indexer" {
		t.Errorf("expected %q, got %q", "cannot reach indexer", we.Error())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/models/... -run 'TestUserMessage_plainError|TestWazuhError_Error' -v
```

Expected: FAIL (plain error still returns raw string; cause still appended).

- [ ] **Step 3: Implement in `pkg/models/errors.go`**

Add constant:

```go
const genericUserMessage = "An unexpected error occurred"
```

Change `WazuhError.Error()`:

```go
func (e *WazuhError) Error() string {
	return e.Message
}
```

Change `UserMessage` fallback:

```go
func UserMessage(err error) string {
	if we, ok := AsWazuhError(err); ok {
		return we.Message
	}
	return genericUserMessage
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/models/... -v
```

Expected: PASS.

---

### Task 4: Extract indexer error classifiers and fix Ping

**Files:**
- Create: `pkg/indexer/errors.go`
- Modify: `pkg/indexer/search.go`
- Modify: `pkg/indexer/client.go`
- Modify: `pkg/indexer/client_test.go`

- [ ] **Step 1: Create `pkg/indexer/errors.go`**

Move `classifyHTTPError`, `classifyNetworkError`, and `sanitizeExcerpt` from `search.go` into this file. Update default HTTP branch to **omit body excerpt**:

```go
default:
	return models.NewWazuhError(models.ErrBadResponse,
		fmt.Sprintf("Wazuh indexer returned an unexpected response (HTTP %d)", status), nil)
```

Keep 401/403/404 cases unchanged (already user-safe).

- [ ] **Step 2: Remove moved functions from `pkg/indexer/search.go`**

`Search` continues calling `classifyHTTPError` and `classifyNetworkError` (same package).

- [ ] **Step 3: Rewrite `Ping` in `pkg/indexer/client.go`**

```go
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_cluster/health", nil)
	if err != nil {
		return models.NewWazuhError(models.ErrBadResponse,
			"failed to build indexer health request", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return classifyNetworkError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return classifyHTTPError(resp.StatusCode, body, c.password)
	}

	return nil
}
```

- [ ] **Step 4: Update `pkg/indexer/client_test.go` auth failure test**

```go
err := client.Ping(context.Background())
if err == nil {
	t.Fatal("expected error")
}
if !models.IsWazuhError(err, models.ErrAuth) {
	t.Fatalf("expected ErrAuth, got %v", err)
}
if !strings.Contains(models.UserMessage(err), "authentication failed") {
	t.Fatalf("expected auth message, got %v", err)
}
```

- [ ] **Step 5: Run indexer tests**

```bash
go test ./pkg/indexer/... -v
```

Expected: PASS.

---

### Task 5: Sanitize wazuhapi default HTTP errors

**Files:**
- Modify: `pkg/wazuhapi/client.go`
- Test: `pkg/wazuhapi/client_test.go`

- [ ] **Step 1: Update `classifyHTTPError` default branch in `pkg/wazuhapi/client.go`**

```go
default:
	return models.NewWazuhError(models.ErrBadResponse,
		fmt.Sprintf("%s returned an unexpected response (HTTP %d)", component, status), nil)
```

Remove `excerpt` variable usage from default branch (variable can remain for potential future logging but is unused in default — delete excerpt line if linter complains).

- [ ] **Step 2: Run wazuhapi tests**

```bash
go test ./pkg/wazuhapi/... -v
```

Expected: PASS. If any test asserts body excerpt in error message, update assertion to check HTTP status only.

---

### Task 6: Sanitize plugin resource and query error paths

**Files:**
- Modify: `pkg/plugin/resource.go`
- Modify: `pkg/plugin/datasource.go`
- Modify: `pkg/plugin/resource_test.go`

- [ ] **Step 1: Add failing resource test in `pkg/plugin/resource_test.go`**

```go
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
```

Add `"strings"` import.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/plugin/... -run TestCallResourceAgents_upstreamErrorSanitized -v
```

Expected: FAIL (body contains `10.0.0.5`).

- [ ] **Step 3: Update `pkg/plugin/resource.go`**

Replace every:

```go
Body: []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
```

With:

```go
Body: []byte(fmt.Sprintf(`{"message":%q}`, models.UserMessage(err))),
```

Add import for `pkg/models` if not present (package alias not needed).

- [ ] **Step 4: Update `wazuhErrToDataResponse` in `pkg/plugin/datasource.go`**

```go
func wazuhErrToDataResponse(err error) backend.DataResponse {
	we, ok := models.AsWazuhError(err)
	if !ok {
		return backend.ErrDataResponse(backend.StatusInternal, models.UserMessage(err))
	}
	// ... rest unchanged
}
```

Note: `UserMessage` on non-WazuhError now returns generic message (from Task 3).

- [ ] **Step 5: Run plugin tests**

```bash
go test ./pkg/plugin/... -v
```

Expected: PASS.

---

### Task 7: Rewrite catalog README

**Files:**
- Modify: `src/README.md`

- [ ] **Step 1: Replace `src/README.md` content**

```markdown
# Wazuh Datasource for Grafana

Connect Grafana to your Wazuh deployment so security data appears in dashboards and Explore — without manual OpenSearch configuration.

Query **alerts**, **vulnerabilities**, **file integrity monitoring (FIM)**, **SCA compliance**, and **agent status** from Wazuh's manager API and indexer.

## Requirements

- Grafana 10.4 or newer
- Wazuh 4.7 or newer with the manager API and indexer reachable from Grafana

## Getting started

### 1. Install the plugin

In Grafana, go to **Administration → Plugins**, search for **Wazuh**, and click **Install**.

### 2. Add a datasource

1. Go to **Connections → Data sources → Add data source**.
2. Select **Wazuh**.
3. Configure:
   - **Manager URL** — Wazuh manager API base URL (e.g. `https://wazuh.example.com:55000`)
   - **Indexer URL** — Wazuh indexer (OpenSearch) URL (e.g. `https://wazuh-indexer.example.com:9200`)
   - **Username** / **Password** — credentials for the manager API
   - **Indexer username** / **Indexer password** — optional; defaults to manager credentials
   - **Skip TLS verify** — enable only for lab environments with self-signed certificates
4. Click **Save & test**. A successful connection confirms both the manager API and indexer are reachable.

### 3. Explore your data

Open **Explore**, select the Wazuh datasource, choose a data type (alerts, vulnerabilities, FIM, SCA, or agent status), and run a query.

## Bundled dashboards

The plugin includes ready-made dashboards:

- Security Overview
- Vulnerabilities
- FIM
- SCA
- Agent Status

Find them under the **Wazuh** folder after installation. If you provision dashboards via YAML, set the datasource **`uid: wazuh`**.

## Further reading

- [Installation and configuration](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/installation.md)
- [RBAC permissions](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/rbac.md)
- [Kubernetes deployment](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/kubernetes.md)
- [Source repository](https://github.com/armanfeyzi/grafana-wazuh-data-source)

## License

Apache 2.0
```

- [ ] **Step 2: Verify no dev jargon remains**

```bash
grep -E 'npm |go run|GF_PLUGINS_ALLOW|unsigned|reviewer-quickstart|development\.md' src/README.md
```

Expected: no matches (exit 1 from grep = success).

---

### Task 8: Version bump and changelog

**Files:**
- Modify: `package.json`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Bump `package.json` version to `0.2.9`**

- [ ] **Step 2: Add CHANGELOG entry at top (below header)**

```markdown
## [0.2.9] — 2026-07-01

### Changed

- **HTTP client** — Use Grafana SDK `backend/httpclient` for proxy, TLS, and transport consistency.
- **Error messages** — Sanitize all user-facing errors; no raw upstream host, port, or response bodies in the UI.
- **Go toolchain** — Bump to Go 1.25.11 (CVE fixes GO-2026-5037, GO-2026-5038, GO-2026-5039).
- **SDK** — Update `grafana-plugin-sdk-go` to v0.292.2.
- **Catalog README** — Rewrite `src/README.md` for Grafana plugin catalog users.
```

Note: `src/plugin.json` uses `%VERSION%` substituted at build time from `package.json` — no direct edit needed.

---

### Task 9: Full verification

**Files:** none (commands only)

- [ ] **Step 1: Run all Go tests**

```bash
go test ./... -count=1
```

Expected: all PASS.

- [ ] **Step 2: Run govulncheck**

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

Expected: no vulnerabilities affecting Go 1.25.11 stdlib (or clean scan).

- [ ] **Step 3: Build frontend and backend**

```bash
npm run build
go run github.com/magefile/mage@latest -v build:linux
```

Expected: exit 0, `dist/` populated.

- [ ] **Step 4: Run plugin validator (matches CI)**

```bash
PLUGIN_ID=$(cat dist/plugin.json | jq -r .id)
PLUGIN_VERSION=$(cat dist/plugin.json | jq -r .info.version)
mv dist ${PLUGIN_ID}
zip ${PLUGIN_ID}-${PLUGIN_VERSION}.zip ${PLUGIN_ID} -r
docker run --pull=always \
  -v $PWD/${PLUGIN_ID}-${PLUGIN_VERSION}.zip:/archive.zip \
  grafana/plugin-validator-cli /archive.zip
```

Expected: validator passes (SDK age warning resolved, Go CVEs clean).

---

### Task 10: Release and resubmit (manual)

- [ ] **Step 1: Commit all changes** (when user requests)

- [ ] **Step 2: Tag and push `v0.2.9`**

```bash
git tag v0.2.9
git push origin main --tags
```

- [ ] **Step 3: Confirm GitHub Release artifacts**

Verify release contains:
- `armanfeyzi-wazuh-datasource-0.2.9.zip`
- `.sha1` checksum file
- Build provenance attestation

- [ ] **Step 4: Update Grafana Cloud submission**

Grafana Cloud → My Plugins → Update Submission:
- Plugin ZIP URL from GitHub Release
- Source URL: `https://github.com/armanfeyzi/grafana-wazuh-data-source/tree/v0.2.9`
- SHA1 from release

- [ ] **Step 5: Reply to ticket #232003** confirming all review items addressed.

- [ ] **Step 6: Update Linear ARM-64** with resubmission note.

---

## Spec coverage checklist

| Spec requirement | Task |
|---|---|
| SDK HTTP client | Task 2 |
| Error sanitization (indexer/client.go, resource.go) | Tasks 3–6 |
| Consistent error audit (search, wazuhapi, datasource) | Tasks 3–6 |
| Go 1.25.11 CVE fix | Task 1 |
| SDK v0.292.2 | Task 1 |
| Catalog README | Task 7 |
| v0.2.9 release | Task 8–10 |
