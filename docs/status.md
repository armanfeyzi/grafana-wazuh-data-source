# Project status

**Last updated:** 2026-05-26 (v0.2.5)

Living summary of what is implemented, what was fixed recently, and what comes next. Task detail: [milestones.md](./milestones.md). Architecture: [project-roadmap.md](./project-roadmap.md).

---

## At a glance

| Phase | Topic | Status |
|-------|--------|--------|
| 0 | Scaffold, CI, dev loop | **Done** |
| 1 | Config, auth, health check | **Done** |
| 2 | Alerts + agents query engine | **Done** |
| 3 | Query editor UI | **Done** |
| 4 | Vulnerabilities, FIM, SCA | **Done** |
| 5 | Bundled dashboards | **Done** |
| 6 | Correlation & variables | **Done** |
| 7 | Release hardening | **Done** |
| — | Deployment architecture refactor | **Done** |

**Latest release:** [v0.2.5](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/tag/v0.2.5)

---

## What works today

### Plugin
- Backend Go plugin: Manager API (JWT) + Indexer (OpenSearch) dual path
- Plugin ID: **`armanfeyzi-wazuh-datasource`**
- Data types: **alerts**, **vulnerabilities**, **FIM**, **SCA**, **agent status**
- Formats: time series, table, stat (per data type)
- Explore query editor with dynamic agent dropdown (from Manager API)
- Save & Test validates both manager and indexer

### Dashboards (bundled)
| Dashboard | UID |
|-----------|-----|
| Security Overview | `wazuh-security-overview` |
| Vulnerabilities | `wazuh-vulnerabilities` |
| File Integrity (FIM) | `wazuh-fim` |
| Security Configuration (SCA) | `wazuh-sca` |
| Agent Status | `wazuh-agent-status` |
| Correlation with Prometheus (Example) | `wazuh-mixed-prometheus-example` |

All panels use datasource type **`armanfeyzi-wazuh-datasource`** and UID **`wazuh`**. Provision from `dist/dashboards/` or import from plugin.

### Template variables
- **Agent** — dynamic query variable from Wazuh datasource (lists all registered agents)
- **Namespace** — dynamic query variable from alert data (empty on non-k8s deployments)
- **Severity** — custom list on Vulnerabilities dashboard (includes unscored `-` CVEs)
- Panel filters (`$agent`, `$severity`) resolved via `applyTemplateVariables` before backend query

### Local development
```bash
make k8s-forward   # both manager + indexer port-forwards (keep running)
make dev-config    # first time: create local datasource yaml
make dev           # build plugin + start Grafana at :3000
```

Grafana container reaches cluster Wazuh via `https://host.containers.internal:55000` and `:9200`.

---

## Recent fixes/additions (v0.2.2 – v0.2.5)

| Version | Change | Details |
|---------|--------|---------|
| v0.2.5 | Dashboard datasource type | Bundled panels now reference `armanfeyzi-wazuh-datasource` (was `wazuh-datasource`) so imports resolve without manual re-selection |
| v0.2.5 | Security Overview layout | Alert annotations; refined severity stat panels and grid layout |
| v0.2.4 | Security Overview redesign | Per-severity stat row (Critical / High / Medium / Low) matching Wazuh UI |
| v0.2.3 | Vulnerability count mismatch | Unscored CVEs (`severity = "-"`) included when "All" severities selected |
| v0.2.2 | SCA rate limiting | Singleflight dedup + 45 s cache on `ListSCAForAgents` to prevent HTTP 429 |
| v0.2.1 | SCA + Overview fixes | Agent 000 included in SCA tables; pie chart field mapping corrected |
| v0.1.0 | Initial release | All five data types, six bundled dashboards, template variables, mixed Prometheus example |

---

## Deployment layout

```text
src/, pkg/              Plugin (no env-specific URLs)
deploy/dev/             Grafana-only dev (make dev)
deploy/wazuh-lab/       Optional local Wazuh docker lab
deploy/kubernetes/      K8s provisioning examples
provisioning/examples/  Copy-paste templates for production
```

Spec: [superpowers/specs/2026-05-22-deployment-architecture-design.md](./superpowers/specs/2026-05-22-deployment-architecture-design.md)

---

## What's next

### Near term
- **Grafana plugin catalog submission** — submit unsigned ZIP from GitHub Releases to [grafana/grafana-plugin-repository](https://github.com/grafana/grafana-plugin-repository) (see [signing.md](./signing.md))
- **Grafana 13 dependency upgrade** — Dependabot PRs #8–#9 open; CI passes but major bump — merge when ready to target Grafana 13
- **Remaining Dependabot PRs** — eslint/test/webpack-cli groups (#6, #7, #10)

### v1.1+
- Separate Manager API vs Indexer credentials (backend supports fallback; needs UI toggle)
- Dynamic `$severity` template variable (same pattern as `$namespace`)
- E2E Playwright tests against local Wazuh Docker lab (tests exist but are skipped in CI)
- Grafana Alerting on Wazuh queries
- Wazuh Cloud testing
- Index pattern version auto-detection improvements

---

## Troubleshooting quick reference

| Symptom | Likely cause |
|---------|----------------|
| `connection refused` on `:55000` | Manager port-forward stopped — run `make k8s-forward` |
| Explore OK, dashboards empty | Datasource UID ≠ `wazuh` |
| Panels show "datasource not found" on import | Datasource type must be `armanfeyzi-wazuh-datasource`; upgrade to v0.2.5+ bundled dashboards |
| Vuln count lower than Wazuh UI | Unscored CVEs excluded — upgrade to v0.2.3+ or add `-` to severity filter |
| SCA panels return HTTP 429 | Upgrade to v0.2.2+ (singleflight + cache) or reduce dashboard refresh rate |
| Vuln/FIM/SCA panels empty | No matching data indexed in Wazuh yet |
| Agent dropdown empty | Manager API unreachable or RBAC denied |
| Panel shows 0 with `$agent` set | Old plugin build without `applyTemplateVariables` — rebuild + restart |
| `$agent` chip visible in panel edit | Normal — variable resolves at query time |

See [development.md](./development.md) and [kubernetes.md](./kubernetes.md) for details.
