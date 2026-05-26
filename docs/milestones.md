# Milestones

Task list for the Wazuh Grafana datasource plugin. Architecture detail: [project-roadmap.md](./project-roadmap.md). **Current status:** [status.md](./status.md).

**Targets:** Wazuh 4.7+, Grafana 10.4+, plugin id `armanfeyzi-wazuh-datasource`.  
**Last updated:** 2026-05-26

---

## Progress overview

| Phase | Status |
|-------|--------|
| 0 — Setup | ✅ Complete |
| 1 — Core infrastructure | ✅ Complete |
| 2 — Query engine MVP | ✅ Complete |
| 3 — Query editor | ✅ Complete |
| 4 — Extended data types | ✅ Complete |
| 5 — Dashboards | ✅ Complete |
| 6 — Correlation | ✅ Complete |
| 7 — Release prep | ✅ Complete |
| Deployment refactor | ✅ Complete |

---

## Phase 0 — Setup ✅

| Task | Done when | Status |
|------|-----------|--------|
| Scaffold with `@grafana/create-plugin` (backend on) | `npm run dev` loads plugin in Grafana | ✅ |
| Pin Go module + plugin SDK | `mage -v build:linux` succeeds | ✅ |
| CI on PR | Lint, Go test, backend build pass | ✅ |
| README dev section | Prerequisites and local run steps documented | ✅ |
| Shared query types | TS + Go structs match for config and query model | ✅ |

---

## Phase 1 — Core infrastructure ✅

| Task | Done when | Status |
|------|-----------|--------|
| ConfigEditor | Manager URL, Indexer URL, user/pass, TLS skip option; saves correctly | ✅ |
| Secure credentials | Password in `secureJsonData`; not exposed to browser | ✅ |
| API auth | JWT fetch + cache; refresh on 401 | ✅ |
| Indexer client | Authenticated ping against indexer | ✅ |
| Health check | Save & Test green when both paths work; clear errors otherwise | ✅ |
| Test harness | Mock or compose-based health check in CI | ✅ |

---

## Phase 2 — Query engine MVP ✅

| Task | Done when | Status |
|------|-----------|--------|
| Query router | Dispatches on `dataType`; useful error for unknown types | ✅ |
| Alerts (time series) | Count over time from `wazuh-alerts-*` | ✅ |
| Alerts (table) | Latest N events with rule, agent, level, description | ✅ |
| Alert filters | Agent name, rule level, rule group | ✅ |
| Agent status | Table from `GET /agents` | ✅ |
| Normalizer | Consistent frame field names | ✅ |
| Unit tests | Query builder fixtures | ✅ |

---

## Phase 3 — Query editor ✅

| Task | Done when | Status |
|------|-----------|--------|
| Data type picker | Alerts, Vuln, FIM, SCA, Agents | ✅ |
| Format picker | Time series / table / stat — per data type | ✅ |
| Filters | Agent, severity/level, inherits Grafana time range | ✅ |
| Aggregations | Count over time, top N, latest events | ✅ |
| Validation | Bad combos blocked before run | ✅ |
| Persistence | Panel JSON round-trips on save | ✅ |

---

## Phase 4 — Extended data types ✅

| Task | Done when | Status |
|------|-----------|--------|
| Vulnerabilities | `wazuh-states-vulnerabilities-*` — severity + per-agent packages | ✅ |
| FIM | Recent changes; top paths; user breakdown | ✅ |
| SCA (live) | Per-agent scores via API | ✅ |
| SCA (history) | Score trend from indexer | ✅ |
| Index version handling | Works with `wazuh-alerts-4.x-*` and legacy patterns | ✅ |
| Enable all types in UI | All five data types selectable and working | ✅ |

---

## Phase 5 — Dashboards ✅

| Task | Done when | Status |
|------|-----------|--------|
| Security overview | Alerts today, severity over time, top rules, active agents | ✅ |
| Vulnerabilities | Severity split, packages per agent, recent CVEs | ✅ |
| FIM | Changes by agent, top paths, by user | ✅ |
| SCA | Score per agent, failed checks, trend | ✅ |
| Agent status | Status counts, OS + version breakdown | ✅ |
| Bundle in plugin | JSON in `src/dashboards/` → `dist/dashboards/` | ✅ |
| Variables | `agent` (query), `severity` (custom), `$__interval` | ✅ |
| Panel variable interpolation | `$agent` / `$severity` resolve before backend query | ✅ |

---

## Phase 6 — Correlation ✅

| Task | Done when | Status |
|------|-----------|--------|
| Agent variable | Template var from datasource agent list | ✅ |
| Panel filter interpolation | `applyTemplateVariables` for `$agent`, `$severity` | ✅ |
| Namespace variable | Exposed when k8s metadata exists in data | ✅ |
| Mixed example dashboard | Prometheus CPU + Wazuh alerts on shared `$agent` | ✅ |
| Field mapping doc | Normalized labels ↔ Prometheus labels | ✅ |
| Provisioning sample | Example datasource YAML for GitOps | ✅ |

**Phase done when:** one variable filters Prometheus and Wazuh panels on the same row. ✅

---

## Phase 7 — Release prep ✅

| Task | Done when | Status |
|------|-----------|--------|
| Error messages | Auth, timeout, missing index, RBAC — all readable | ✅ |
| Perf guardrails | Server-side 25s timeout; 32 MB cap; clampLimit | ✅ |
| Security pass | No creds in logs; TLS warning UI; RBAC guide | ✅ |
| User docs | Install, config, data types, dashboards, troubleshooting | ✅ |
| Contributor docs | Dev setup, adding a data type (CONTRIBUTING.md) | ✅ |
| v0.1.0 | CHANGELOG, signed zip workflow, plugin.json metadata | ✅ |

---

## Deployment architecture ✅ (2026-05-22)

| Task | Status |
|------|--------|
| `deploy/dev/` Grafana-only compose | ✅ |
| `deploy/wazuh-lab/` optional lab | ✅ |
| `deploy/kubernetes/` examples | ✅ |
| `provisioning/examples/` templates | ✅ |
| `make dev`, `make k8s-forward`, `make lab-*` | ✅ |
| Persistent Grafana volume + dev provisioning | ✅ |
| Docs: development, installation, kubernetes, status | ✅ |
