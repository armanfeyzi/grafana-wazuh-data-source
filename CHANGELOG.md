# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.2.8] — 2026-06-14

### Changed

- **Go toolchain** — Set `go 1.25.7` and `grafana-plugin-sdk-go v0.291.0` so Grafana's catalog validator (Go 1.25) can run `govulncheck` source scans.
- **Release pipeline** — Build releases with Go 1.25.10 to match Grafana's review environment.

---

## [0.2.7] — 2026-06-11

### Changed

- **Release pipeline** — Install `govulncheck` before packaging and attach GitHub build provenance attestation to release ZIPs (Grafana catalog review requirements).
- **Plugin metadata** — Added sponsorship link in `plugin.json`.

---

## [0.2.6] — 2026-06-11

### Added

- **Catalog screenshots** — Security Overview dashboard, datasource configuration, and Explore query editor images in `plugin.json` for the Grafana plugin catalog.
- **Reviewer quickstart** — [docs/reviewer-quickstart.md](docs/reviewer-quickstart.md) for Grafana plugin reviewers validating the datasource in under 15 minutes.
- **Screenshot capture script** — `scripts/capture-catalog-screenshots.mjs` to regenerate catalog images from a Grafana instance.

### Changed

- **Plugin metadata** — Added repository, documentation, and reviewer quickstart links in `plugin.json`.

---

## [0.2.5] — 2026-05-26

### Fixed

- **Bundled dashboards referenced wrong datasource type** — All bundled dashboard panels used `wazuh-datasource` while the plugin ID is `armanfeyzi-wazuh-datasource`. Panels now resolve the datasource correctly on import without manual re-selection.

### Changed

- **Security Overview dashboard annotations and layout** — Added alert annotations for improved visibility; refined severity stat panel grid positions, thresholds, and titles for consistency with the Wazuh UI.

---

## [0.2.4] — 2026-05-26

### Changed

- **Security Overview dashboard redesigned to match Wazuh Overview layout** — Added four dedicated severity stat panels at the top of the dashboard (Critical / High / Medium / Low), each coloured with its standard severity colour and filtered by the corresponding Wazuh rule level range (Critical ≥15, High 12–14, Medium 7–11, Low 0–6). Removed the redundant "Alerts By Level" pie chart and tightened the secondary row to show total alerts, registered agents, alerts by agent, and alerts by rule.

---

## [0.2.3] — 2026-05-25

### Fixed

- **Vulnerability count mismatch vs Wazuh interface** — CVEs that Wazuh has not yet assigned a CVSS score to are stored in the indexer with `vulnerability.severity = "-"`. Because `"-"` was absent from the `$severity` variable's option list (`Critical,High,Medium,Low,None`), selecting "All" expanded to those five values and applied a `terms` filter that silently excluded all unscored CVEs (5,676 documents in the dev cluster). Added `"-"` to the dashboard variable and to `VULNERABILITY_SEVERITIES` so unscored CVEs are included by default. The query editor now shows these as "Unscored (-)" rather than a bare dash.

---

## [0.2.2] — 2026-05-25

### Fixed

- **SCA dashboard HTTP 429 rate-limit errors** — The SCA dashboard was making one Wazuh Manager API call per agent for every panel that renders (N+1 pattern). With multiple panels loading simultaneously and a 1-minute auto-refresh, this burst would exceed the Wazuh default request rate limit. Fixed by adding two layers of protection to `ListSCAForAgents`:
  - **Singleflight deduplication**: concurrent calls with identical parameters (e.g. three SCA panels loading at the same time) now share a single in-flight set of API requests instead of each making independent N+1 calls.
  - **45-second TTL cache**: the result is cached so the 1-minute dashboard auto-refresh re-uses the previous response rather than triggering a fresh burst.

---

## [0.1.0] — 2026-05-22

First public release. All five Wazuh data types, five bundled dashboards, template variable support, and a mixed Prometheus correlation dashboard.

### Added

#### Plugin core
- Backend Go plugin with dual-path routing: Wazuh Manager REST API (JWT) and Wazuh Indexer (OpenSearch)
- Secure credential storage via Grafana `secureJsonData` — credentials never reach the browser
- TLS skip-verify option for self-signed lab certificates (with UI warning in production)
- `Save & Test` health check validates both manager API and indexer connectivity
- Structured error classification: auth, forbidden, unreachable, timeout, index-missing — each returns a user-readable message with appropriate HTTP status code
- Server-side OpenSearch query timeout (25s) to guard against long-running queries
- Response body read cap (32 MB) to protect memory on large deployments
- Credential sanitization — passwords are never included in panel error messages

#### Data types
- **Alerts** — time series (date histogram), table (latest N events), stat (total count)
  - Filters: agent name, rule level min/max, rule groups
- **Vulnerabilities** — table (packages + CVEs + severity), stat (count), time series (detections over time)
  - Filters: agent name, severity (Critical / High / Medium / Low / None)
- **File Integrity Monitoring (FIM)** — time series, table (file changes with agent/path/user/action), stat
  - Filters: agent name
- **Security Configuration Assessment (SCA)** — table (live compliance scores via Manager API), time series (historical scan activity), stat
  - Filters: agent name
- **Agent status** — table (agent name, status, IP, OS, version, last keepalive)

#### Query editor
- Data type dropdown with format options per type (time series / table / stat)
- Dynamic agent filter: multi-select populated live from the Wazuh Manager API
- Severity filter for vulnerability panels
- Rule level range and rule groups filter for alert panels
- Inline validation errors before query runs
- Template variable interpolation (`$agent`, `$severity`) resolved server-side

#### Template variables
- **Agents** — dynamic query variable listing all registered agent names
- **Namespaces** — dynamic query variable listing distinct Kubernetes namespaces from alert data (empty on non-k8s deployments)

#### Bundled dashboards
- **Security Overview** (`wazuh-security-overview`) — alerts today, alert trend, latest alerts table; `$agent` variable
- **Vulnerabilities** (`wazuh-vulnerabilities`) — severity breakdown, packages per agent, recent CVEs; `$agent` + `$severity` variables
- **File Integrity (FIM)** (`wazuh-fim`) — FIM events over time, recent changes table, top paths; `$agent` variable
- **Security Configuration (SCA)** (`wazuh-sca`) — live compliance scores, score trend; `$agent` variable
- **Agent Status** (`wazuh-agent-status`) — connection status counts, OS distribution
- **Correlation with Prometheus (Example)** (`wazuh-mixed-prometheus-example`) — Prometheus node CPU + Wazuh alerts linked by `$agent`; `$namespace` variable

#### Configuration
- Manager URL + API username/password
- Indexer URL + optional separate indexer credentials (falls back to API credentials)
- Custom index prefix override
- Skip TLS verify (dev only)

#### Deployment
- `make dev` — Grafana-only local development with Docker Compose
- `make k8s-forward` — port-forwards both Wazuh API and Indexer from Kubernetes
- `make lab-up` — local Wazuh Docker lab (optional)
- `provisioning/examples/` — GitOps datasource and dashboard YAML templates
- `deploy/kubernetes/` — Kubernetes Secret + ConfigMap examples

#### Documentation
- [Installation guide](docs/installation.md) — requirements, install, configure, provision
- [Development guide](docs/development.md) — local dev loop, build commands
- [Kubernetes guide](docs/kubernetes.md) — in-cluster deployment
- [RBAC guide](docs/rbac.md) — minimum required API and indexer permissions
- [Field mapping reference](docs/field-mapping.md) — all normalized field names; Prometheus/Loki correlation patterns
- [CONTRIBUTING.md](CONTRIBUTING.md) — dev setup, architecture, how to add a new data type

### Requirements

- Grafana **10.4+**
- Wazuh **4.7+** (Manager API + Indexer)
- Go 1.22+ / Node.js 22+ (build only)
