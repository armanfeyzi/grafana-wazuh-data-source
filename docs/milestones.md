# Milestones

Task list for the Wazuh Grafana datasource plugin. Architecture detail: [project-roadmap.md](./project-roadmap.md).

**Targets:** Wazuh 4.7+, Grafana 10.4+, plugin id `wazuh-datasource`.

---

## Phase 0 — Setup

| Task | Done when |
|------|-----------|
| Scaffold with `@grafana/create-plugin` (backend on) | `npm run dev` loads plugin in Grafana |
| Pin Go module + plugin SDK | `mage -v build:linux` succeeds |
| CI on PR | Lint, Go test, backend build pass |
| README dev section | Prerequisites and local run steps documented |
| Shared query types | TS + Go structs match for config and query model |

**Phase done:** clone → dev server → plugin visible in datasource list; CI green.

---

## Phase 1 — Core infrastructure

| Task | Done when |
|------|-----------|
| ConfigEditor | Manager URL, Indexer URL, user/pass, TLS skip option; saves correctly |
| Secure credentials | Password in `secureJsonData`; not exposed to browser |
| API auth | JWT fetch + cache; refresh on 401 |
| Indexer client | Authenticated ping against indexer |
| Health check | Save & Test green when both paths work; clear errors otherwise |
| Test harness | Mock or compose-based health check in CI |

**Phase done:** Save & Test works against a real Wazuh lab.

---

## Phase 2 — Query engine MVP

| Task | Done when |
|------|-----------|
| Query router | Dispatches on `dataType`; useful error for unknown types |
| Alerts (time series) | Count over time from `wazuh-alerts-*` |
| Alerts (table) | Latest N events with rule, agent, level, description |
| Alert filters | Agent name, rule level, rule group |
| Agent status | Table from `GET /agents` |
| Normalizer | Consistent frame field names |
| Unit tests | Query builder fixtures; ~80% on builder package |

**Phase done:** alert time-series panel + agent status table work end-to-end.

---

## Phase 3 — Query editor

| Task | Done when |
|------|-----------|
| Data type picker | Alerts, Vuln, FIM, SCA, Agents (unimplemented types disabled) |
| Format picker | Time series / table / stat — per data type |
| Filters | Agent, severity/level, inherits Grafana time range |
| Aggregations | Count over time, top N, latest events |
| Validation | Bad combos blocked before run |
| Persistence | Panel JSON round-trips on save |

**Phase done:** custom alert panel built without writing OpenSearch JSON.

---

## Phase 4 — Extended data types

| Task | Done when |
|------|-----------|
| Vulnerabilities | `wazuh-states-vulnerabilities-*` — severity + per-agent packages |
| FIM | Recent changes; top paths; user breakdown |
| SCA (live) | Per-agent scores via API |
| SCA (history) | Score trend from indexer |
| Index version handling | Works with `wazuh-alerts-4.x-*` and legacy patterns |
| Enable all types in UI | All five data types selectable and working |

**Phase done:** every brief data type returns table + aggregated results.

---

## Phase 5 — Dashboards

| Task | Done when |
|------|-----------|
| Security overview | Alerts today, severity over time, top rules, active agents |
| Vulnerabilities | Severity split, packages per agent, recent CVEs |
| FIM | Changes by agent, top paths, by user |
| SCA | Score per agent, failed checks, trend |
| Agent status | Status counts, OS + version breakdown |
| Bundle in plugin | JSON in `dashboards/`; prefixed UIDs |
| Variables | `agent`, `severity`, `$__interval` on each |

**Phase done:** import all five dashboards; panels populate within 5 min of adding datasource.

---

## Phase 6 — Correlation

| Task | Done when |
|------|-----------|
| Agent variable | Template var from datasource agent list |
| Namespace variable | Exposed when k8s metadata exists in data |
| Mixed example dashboard | Prometheus CPU + Wazuh alerts on shared `$agent` |
| Field mapping doc | Normalized labels ↔ Prometheus labels |
| Provisioning sample | Example datasource YAML for GitOps |

**Phase done:** one variable filters Prometheus and Wazuh panels on the same row.

---

## Phase 7 — Release prep

| Task | Done when |
|------|-----------|
| Error messages | Auth, timeout, missing index, RBAC — all readable |
| Perf check | 7-day alert time series < 10s on medium deployment |
| Security pass | No creds in logs; TLS docs; RBAC guide |
| User docs | Install, config, data types, dashboards, troubleshooting |
| Contributor docs | Dev setup, adding a data type |
| v0.1.0 | Signed zip, GitHub release, `plugin.json` declares Grafana versions |

**Phase done:** docs-only path from zero to working dashboards; tagged release.
