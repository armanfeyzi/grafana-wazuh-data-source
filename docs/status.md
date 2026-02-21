# Project status

**Last updated:** 2026-05-22 (Phase 6 complete)

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
| 7 | Release hardening | **Not started** |
| — | Deployment architecture refactor | **Done** |

---

## What works today

### Plugin
- Backend Go plugin: Manager API (JWT) + Indexer (OpenSearch) dual path
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

All panels use datasource UID **`wazuh`**. Provision from `dist/dashboards/` or import from plugin.

### Template variables
- **Agent** — dynamic query variable from Wazuh datasource (lists all registered agents)
- **Severity** — custom list on Vulnerabilities dashboard
- Panel filters (`$agent`, `$severity`) resolved via `applyTemplateVariables` before backend query

### Local development
```bash
make k8s-forward   # both manager + indexer port-forwards (keep running)
make dev-config    # first time: create local datasource yaml
make dev           # build plugin + start Grafana at :3000
```

Grafana container reaches cluster Wazuh via `https://host.containers.internal:55000` and `:9200`.

---

## Recent fixes/additions (2026-05-22)

| Change | Details |
|--------|---------|
| Namespace template variable | New `/namespaces` resource endpoint queries `kubernetes.namespace` agg from `wazuh-alerts-*`; returns empty list on non-k8s deployments |
| `VariableQueryEditor` picker | Replaced invisible placeholder with Select: **Agents** / **Namespaces** |
| Mixed correlation dashboard | New bundled dashboard `wazuh-mixed-prometheus-example` — Prometheus node CPU + Wazuh alerts sharing `$agent` variable |
| Field mapping doc | `docs/field-mapping.md` — all normalized plugin fields, Prometheus label correlation patterns |
| Dashboard agent dropdown hardcoded | Dashboard used custom variable; no plugin variable support | `CustomVariableSupport` + query-type `$agent` variable |
| Panels empty with `$agent` filter | Literal `$agent` sent to backend | `applyTemplateVariables` on datasource; backend ignores unresolved `$…` |
| Datasource lost on restart | No persistent volume / provisioning | Grafana data volume + file provisioning in `deploy/dev/` |
| `connection refused` on manager API | Manager port-forward not running (indexer still up) | `make k8s-forward`; docs updated |
| Explore works, dashboards empty | Wrong datasource UID or missing `uid: wazuh` | All bundled dashboards pinned to `wazuh` |

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

### Phase 7 — release prep (current focus)
1. **User-facing error messages** — auth, timeout, missing index, RBAC denied — all readable
2. **Performance check** — 7-day alert time series < 10s on medium deployment
3. **Security review** — no credentials in logs, TLS defaults, RBAC guide
4. **`CONTRIBUTING.md`** — dev setup, architecture overview, how to add a data type
5. **v0.1.0** — signed plugin ZIP, GitHub release with changelog, declared Grafana versions in `plugin.json`

### Later (v1.1+)
- Separate Manager API vs Indexer credentials
- Grafana Alerting on Wazuh queries
- Wazuh Cloud testing
- Index pattern version auto-detection improvements

---

## Troubleshooting quick reference

| Symptom | Likely cause |
|---------|----------------|
| `connection refused` on `:55000` | Manager port-forward stopped — run `make k8s-forward` |
| Explore OK, dashboards empty | Datasource UID ≠ `wazuh` |
| Vuln/FIM/SCA panels empty | No matching data indexed in Wazuh yet |
| Agent dropdown empty | Manager API unreachable or RBAC denied |
| Panel shows 0 with `$agent` set | Old plugin build without `applyTemplateVariables` — rebuild + restart |
| `$agent` chip visible in panel edit | Normal — variable resolves at query time |

See [development.md](./development.md) and [kubernetes.md](./kubernetes.md) for details.
