# Grafana Catalog Review Fixes — Design Spec

**Date:** 2026-07-01  
**Ticket:** Grafana #232003  
**Linear:** ARM-64 (Submit plugin to Grafana catalog)  
**Target release:** v0.2.9 (patch)

## Summary

Address all required and suggested changes from the Grafana plugin review team so the Wazuh datasource can proceed to signing. The plugin already passes functional review (Save & test, dashboards render). This work is compliance, security hygiene, and catalog documentation.

## Decisions

| Decision | Choice |
|---|---|
| Error sanitization scope | **B — Consistent:** fix named files plus audit all user-facing error paths |
| Go version | **1.25.11** (CVE fix, minimal churn from 1.25.7/1.25.10) |
| Release version | **0.2.9** (patch bump) |
| HTTP client approach | **Thin wrapper** in `pkg/httpclient` using SDK `backend/httpclient` |
| README approach | **Rewrite `src/README.md` in place** for catalog users; keep dev content in root `README.md` and `docs/` |

## Requirements (from Grafana review)

### Required

1. **SDK HTTP client** — Replace direct `http.Client` + custom transport in `pkg/httpclient/client.go` with Grafana SDK `backend/httpclient` so managed proxy, TLS, and transport settings apply.
2. **Error sanitization** — In `pkg/indexer/client.go` and `pkg/plugin/resource.go`, stop returning raw `err.Error()` or upstream response bodies to users.
3. **Go CVE rebuild** — Rebuild with Go ≥ 1.25.11 (fixes GO-2026-5037, GO-2026-5038, GO-2026-5039).
4. **Catalog README** — Rewrite `src/README.md` for catalog users; remove developer jargon and manual install instructions.

### Suggested

5. **SDK update** — Bump `grafana-plugin-sdk-go` to latest (v0.292.2 at time of design).

---

## 1. HTTP Client Migration

### Current state

`pkg/httpclient/client.go` creates a raw `http.Client` with a cloned `http.DefaultTransport` and manual `InsecureSkipVerify`. `NewDatasource` calls `httpclient.New(config.TlsSkipVerify)` and ignores the context parameter.

### Target state

```go
// pkg/httpclient/client.go
func New(ctx context.Context, settings backend.DataSourceInstanceSettings) (*http.Client, error) {
    opts, err := settings.HTTPClientOptions(ctx)
    if err != nil {
        return nil, err
    }
    opts.Timeouts = &sdkhttpclient.TimeoutOptions{Timeout: 30 * time.Second}
    return sdkhttpclient.New(opts)
}
```

- `settings.HTTPClientOptions(ctx)` reads `tlsSkipVerify` from datasource JSON (field already present in `PluginSettings`).
- Proxy options come from Grafana datasource proxy configuration automatically.
- `NewDatasource(ctx, settings)` propagates errors from client creation.

### Test helper

```go
func NewTest(skipTLSVerify bool) (*http.Client, error) {
    return sdkhttpclient.New(sdkhttpclient.Options{
        Timeouts: &sdkhttpclient.TimeoutOptions{Timeout: 30 * time.Second},
        TLS:      &sdkhttpclient.TLSOptions{InsecureSkipVerify: skipTLSVerify},
    })
}
```

All test files currently calling `httpclient.New(true)` update to `httpclient.NewTest(true)`.

### Files touched

| File | Change |
|---|---|
| `pkg/httpclient/client.go` | Rewrite to use SDK |
| `pkg/plugin/datasource.go` | Pass `ctx` + `settings` to `httpclient.New` |
| `pkg/indexer/client_test.go` | Use `NewTest` |
| `pkg/wazuhapi/client_test.go` | Use `NewTest` |
| `pkg/plugin/health_test.go` | Use `NewTest` |
| `pkg/plugin/query_test.go` | Use `NewTest` |
| `pkg/plugin/resource_test.go` | Use `NewTest` |

---

## 2. Error Sanitization (Consistent)

### Principle

User-visible messages must never contain:

- Raw network/transport errors (host, port, IP, connection refused strings)
- Upstream HTTP response bodies
- Wrapped internal error chains via `err.Error()`

Internal causes remain on `WazuhError.Cause` and are available via `errors.Unwrap()` for future logging. `WazuhError.Error()` returns only the user-facing `Message` field.

### Changes by file

#### `pkg/indexer/errors.go` (new)

Extract `classifyHTTPError`, `classifyNetworkError`, and `sanitizeExcerpt` from `pkg/indexer/search.go`. Shared by `Ping` and `Search`.

#### `pkg/indexer/client.go`

Refactor `Ping` to return typed `WazuhError` via shared classifiers:

- Network errors → `ErrUnreachable` or `ErrTimeout` with generic message
- HTTP 401 → authentication failed message (no body)
- HTTP ≥ 400 default → `"Wazuh indexer returned an unexpected response (HTTP {status})"` (no body excerpt)

#### `pkg/indexer/search.go`

Remove body excerpt from default `classifyHTTPError` branch (status code only).

#### `pkg/wazuhapi/client.go`

Same default-branch change: no body excerpt in user message.

#### `pkg/plugin/resource.go`

Replace all `err.Error()` in JSON responses with `models.UserMessage(err)`.

#### `pkg/models/errors.go`

- `UserMessage(err)` fallback: return `"An unexpected error occurred"` instead of `err.Error()` for non-`WazuhError` errors.
- `WazuhError.Error()`: return `Message` only; do not append `Cause`.

#### `pkg/plugin/datasource.go`

- `wazuhErrToDataResponse` non-`WazuhError` path: use generic fallback message.
- `CheckHealth` validation messages: keep as-is (static validation strings, no upstream leakage).

### User-facing message reference

| Scenario | User message |
|---|---|
| Indexer/manager unreachable | `cannot connect — check the URL and that the service is running` |
| Timeout | `request timed out — check connectivity or try again` |
| Auth failure | `authentication failed — check username and password` |
| Unexpected HTTP | `returned an unexpected response (HTTP {status})` |
| Unknown internal error | `An unexpected error occurred` |

### Tests to update

| File | Change |
|---|---|
| `pkg/indexer/client_test.go` | Assert `WazuhError` messages, not raw upstream strings |
| `pkg/models/errors_test.go` | `UserMessage(plainError)` → generic; `WazuhError.Error()` without cause suffix |
| `pkg/plugin/resource_test.go` | Add failure-path test asserting no raw upstream text in response body |

---

## 3. Toolchain and SDK

| File | Change |
|---|---|
| `go.mod` | `go 1.25.11`; `grafana-plugin-sdk-go v0.292.2` |
| `go.sum` | Regenerated via `go mod tidy` |
| `.github/workflows/release.yml` | `go-version: '1.25.11'` in both Setup Go and package-plugin steps |
| `.github/workflows/ci.yml` | No change needed (`go-version-file: go.mod`) |
| `package.json` | Version `0.2.9` |
| `src/plugin.json` | Version `0.2.9` (or via build `%VERSION%` substitution) |
| `CHANGELOG.md` | New `## [0.2.9]` section |

### CHANGELOG entry (draft)

```markdown
## [0.2.9] — 2026-07-01

### Changed

- **HTTP client** — Use Grafana SDK `backend/httpclient` for proxy, TLS, and transport consistency.
- **Error messages** — Sanitize all user-facing errors; no raw upstream details in UI.
- **Go toolchain** — Bump to Go 1.25.11 (CVE fixes GO-2026-5037/5038/5039).
- **SDK** — Update `grafana-plugin-sdk-go` to v0.292.2.
- **Catalog README** — Rewrite `src/README.md` for Grafana plugin catalog users.
```

### Verification

```bash
go test ./...
npm run build
# CI plugin-validator-cli on packaged ZIP
govulncheck ./...
```

---

## 4. Catalog README (`src/README.md`)

### Audience

Users who install from the Grafana plugin catalog (Plugins → Install).

### Structure

1. **What it does** — Connects Grafana to Wazuh manager API and indexer for security dashboards.
2. **Requirements** — Grafana 10.4+, Wazuh 4.7+ with manager API and indexer reachable.
3. **Getting started**
   - Install from catalog
   - Configuration → Data sources → Add Wazuh datasource
   - Enter manager URL, indexer URL, credentials
   - Save & test
   - Open Explore or bundled dashboards
4. **Bundled dashboards** — Security Overview, Vulnerabilities, FIM, SCA, Agent Status; note `uid: wazuh` for provisioning.
5. **Data types** — alerts, vulnerabilities, FIM, SCA, agent status.
6. **Further reading** — Links to GitHub `docs/installation.md`, `docs/rbac.md`, `docs/kubernetes.md`.

### Remove from `src/README.md`

- Manual ZIP download and install
- `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`
- Developer setup table (development, reviewer quickstart, roadmap, status)
- npm/go build instructions

### Unchanged

- Root `README.md` — keeps dev quick start and unsigned install banner.
- `docs/` — all existing guides remain.

---

## 5. Resubmission Checklist

1. Implement all changes on a feature branch.
2. Run full CI locally (`go test ./...`, `npm run build`, plugin validator).
3. Merge to `main`, tag `v0.2.9`.
4. Confirm GitHub Release workflow produces ZIP + `.sha1` + attestation.
5. Grafana Cloud → My Plugins → Update Submission:
   - **Plugin ZIP URL:** `https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/download/v0.2.9/armanfeyzi-wazuh-datasource-0.2.9.zip`
   - **Source URL:** `https://github.com/armanfeyzi/grafana-wazuh-data-source/tree/v0.2.9`
   - **SHA1:** from release artifact
6. Reply to ticket #232003 confirming all items addressed.
7. Update Linear ARM-64 with progress note.

---

## Out of scope

- Grafana Alerting support (ARM-68)
- Structured backend logging of internal errors (option C from brainstorming)
- Signing the plugin locally (Grafana signs after approval)
- Changes to root `README.md` or `docs/` beyond cross-links if needed

## Risk and mitigation

| Risk | Mitigation |
|---|---|
| SDK httpclient breaks httptest-based unit tests | `NewTest` helper with SDK options for tests |
| Proxy behavior untested locally | Rely on SDK contract; document in CHANGELOG |
| SDK bump introduces compile errors | Run `go test ./...` after `go get -u` |
| Error message changes break frontend | Messages are display-only strings; no parsing in frontend |
