# Contributing to grafana-wazuh-data-source-plugin

Thank you for contributing! This guide covers everything you need to build, test, and extend the plugin.

---

## Development environment

**Requirements:**
- Node.js 22+
- Go 1.22+
- Docker or Podman (for local Grafana)
- A running Wazuh deployment (cluster or local lab — see below)

```bash
git clone https://github.com/armanfeyzi/grafana-wazuh-data-source
cd grafana-wazuh-data-source
npm ci
```

---

## Local dev loop

### Option A — Against a Kubernetes Wazuh cluster

```bash
# Terminal 1: keep port-forwards running
make k8s-forward

# Terminal 2: build and start Grafana
make dev-config   # first time only — creates deploy/dev/provisioning/datasources/wazuh.yaml
make dev          # builds plugin + starts Grafana at http://localhost:3000
```

Edit `deploy/dev/provisioning/datasources/wazuh.yaml` with your Wazuh credentials. See [docs/development.md](docs/development.md).

### Option B — Local Wazuh lab (Docker)

```bash
make lab-up       # pulls wazuh-docker and starts single-node Wazuh
make dev          # builds plugin + connects Grafana to the local lab
```

### Iterating on backend (Go)

```bash
make backend      # rebuild Go binary only
# Then restart the Grafana container:
docker compose -f deploy/dev/docker-compose.yaml restart grafana
```

### Iterating on frontend (TypeScript)

The frontend live-reloads automatically with `npm run dev`. The Grafana container picks up changes in `dist/` via the volume mount.

---

## Running tests

```bash
# Backend (Go)
go test ./pkg/...

# Frontend (TypeScript)
npx jest --ci
npm run typecheck
npm run lint
```

All tests must pass before submitting a PR.

---

## Architecture overview

The plugin has two layers:

```
Browser (React/TS)           Go Backend Plugin
─────────────────────        ─────────────────────────────────
ConfigEditor.tsx     ──────▶  CheckHealth (manager + indexer)
QueryEditor.tsx      ──────▶  QueryData → executeQuery()
VariableQueryEditor  ──────▶  CallResource (agents, namespaces)
```

### Data-path routing

`pkg/plugin/query.go` dispatches on `models.DataType`:

| DataType | Primary backend | Function |
|----------|----------------|----------|
| `alerts` | Indexer (`wazuh-alerts-*`) | `queryAlerts()` |
| `vulnerabilities` | Indexer (`wazuh-states-vulnerabilities-*`) | `queryVulnerabilities()` |
| `fim` | Indexer (`wazuh-alerts-*`) | `queryFIM()` |
| `sca` | REST API (live) or Indexer (history) | `querySCA()` |
| `agents` | REST API | `queryAgents()` |

### Query model (structured, not raw JSON)

The frontend sends a typed `WazuhQuery` object. The backend translates it to OpenSearch DSL or Wazuh API calls. Users never write queries by hand.

### Error handling

Backend errors are classified via `models.WazuhError` with typed codes (`ErrAuth`, `ErrForbidden`, `ErrUnreachable`, `ErrTimeout`, `ErrIndexMissing`, `ErrBadResponse`). `datasource.go` maps these to Grafana HTTP status codes. Credentials are never included in user-visible error messages.

---

## How to add a new data type

Follow these steps to add, for example, a `logs` data type:

### 1. Add the constant (Go model)

**`pkg/models/query.go`**
```go
const (
    // existing…
    DataTypeLogs DataType = "logs"
)
```

### 2. Add the indexer query builder + parser

**`pkg/indexer/logs.go`** — implement:
- `BuildLogsTimeSeriesQuery(p QueryParams) ([]byte, error)`
- `BuildLogsTableQuery(p QueryParams) ([]byte, error)`
- `ParseLogsTimeSeriesFrame(raw []byte, refID string) (*data.Frame, error)`
- `ParseLogsTableFrames(raw []byte, refID string) ([]*data.Frame, error)`

Follow the pattern in `pkg/indexer/alerts.go`.

**`pkg/indexer/logs_test.go`** — cover the query builders with fixture JSON and the parsers with representative responses.

### 3. Wire into the query router

**`pkg/plugin/query.go`** — add a `case models.DataTypeLogs:` in `executeQuery()`.

### 4. Add the format validator

**`pkg/plugin/validate.go`** — add `models.DataTypeLogs` to the `validateFormat()` switch.

### 5. Add TypeScript types

**`src/types.ts`** — add `'logs'` to `WazuhDataType`.

**`src/queryUtils.ts`** — add:
- `DATA_TYPE_LABELS['logs'] = 'Log events'`
- Format list: `const LOGS_FORMATS: WazuhQueryFormat[] = ['time_series', 'table']`
- Cases in `formatsForDataType()`, `defaultFormatForDataType()`, `showAgentFilter()`

### 6. Frontend query editor

The `QueryEditor.tsx` reads from `queryUtils.ts` — no changes needed unless you need a new filter control specific to `logs`.

### 7. Field mapping doc

**`docs/field-mapping.md`** — document the normalized field names emitted by the new parser.

### 8. Bundled dashboard (optional)

Add `src/dashboards/wazuh-logs.json` and register it in `src/plugin.json`.

---

## PR checklist

- [ ] `go test ./pkg/...` passes
- [ ] `npx jest --ci` passes
- [ ] `npm run typecheck && npm run lint` — 0 errors
- [ ] New backend code has unit tests with fixture JSON
- [ ] Error paths return `models.WazuhError` (not raw `fmt.Errorf`)
- [ ] No credentials or internal paths in user-visible error messages
- [ ] `docs/field-mapping.md` updated if new fields are emitted
- [ ] `docs/status.md` updated if a phase or milestone changes

---

## Project structure

```
src/                         Frontend (React/TypeScript)
  components/
    ConfigEditor.tsx         Datasource configuration UI
    QueryEditor.tsx          Panel query builder UI
    VariableQueryEditor.tsx  Template variable type picker
  dashboards/                Bundled JSON dashboards
  datasource.ts              Frontend datasource class
  queryUtils.ts              Format/filter helpers
  types.ts                   Shared query + config types
  variableSupport.ts         Custom variable query support

pkg/
  httpclient/                Shared HTTP client (TLS, timeout)
  indexer/                   OpenSearch query builders + parsers
  models/                    Shared Go types (query, settings, errors)
  plugin/                    Grafana plugin handlers (query, health, resource)
  wazuhapi/                  Wazuh REST API client

deploy/
  dev/                       Grafana-only local development (make dev)
  wazuh-lab/                 Optional local Wazuh Docker lab
  kubernetes/                Kubernetes provisioning examples

docs/                        User and contributor documentation
provisioning/examples/       GitOps datasource + dashboard templates
```

---

## Questions?

Open a GitHub issue. Include the Grafana version, Wazuh version, and the exact error message from the panel or browser console.
