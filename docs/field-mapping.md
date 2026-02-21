# Field mapping reference

This document lists every field name emitted by the Wazuh datasource plugin, organised by data type. Use these names when building Grafana variable expressions or creating correlations with Prometheus/Loki datasources.

---

## Alerts (`dataType: alerts`)

### Time series format

| Field | Type | Description |
|-------|------|-------------|
| `Time` | time | Bucket timestamp |
| `Alerts` | number | Alert count per interval |

### Table format

| Field | Type | Wazuh source field | Notes |
|-------|------|--------------------|-------|
| `Time` | time | `@timestamp` | Alert event time |
| `agent` | string | `agent.name` | Use to link with Prometheus `instance` label |
| `agent_id` | string | `agent.id` | Wazuh internal agent ID |
| `rule_id` | string | `rule.id` | Wazuh rule identifier |
| `severity_level` | number | `rule.level` | 0–15 scale; 12+ = critical |
| `rule_description` | string | `rule.description` | Human-readable rule summary |
| `rule_groups` | string | `rule.groups` | Comma-separated, e.g. `syscheck, fim` |

### Stat format

| Field | Type | Description |
|-------|------|-------------|
| `value` | number | Total alert count matching filters |

---

## Vulnerabilities (`dataType: vulnerabilities`)

### Table format

| Field | Type | Wazuh source field | Notes |
|-------|------|--------------------|-------|
| `agent` | string | `agent.name` | |
| `package` | string | `package.name` | Affected package |
| `version` | string | `package.version` | Installed version |
| `cve` | string | `vulnerability.id` | CVE identifier |
| `severity` | string | `vulnerability.severity` | Critical / High / Medium / Low / None |
| `cvss3_score` | number | `vulnerability.score.base` | CVSS v3 base score |
| `description` | string | `vulnerability.description` | |

### Stat format

| Field | Type | Description |
|-------|------|-------------|
| `value` | number | Total vulnerability count |

---

## File Integrity Monitoring (`dataType: fim`)

### Time series format

| Field | Type | Description |
|-------|------|-------------|
| `Time` | time | Bucket timestamp |
| `FIM events` | number | Event count per interval |

### Table format

| Field | Type | Wazuh source field | Notes |
|-------|------|--------------------|-------|
| `Time` | time | `@timestamp` | |
| `agent` | string | `agent.name` | |
| `path` | string | `syscheck.path` | Monitored file path |
| `event` | string | `syscheck.event` | `added` / `modified` / `deleted` |
| `user` | string | `syscheck.uname_after` | User that triggered the change |
| `md5` | string | `syscheck.md5_after` | Post-change MD5 hash |

---

## Security Configuration Assessment (`dataType: sca`)

### Table format (live — from Manager API)

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Agent name |
| `policy` | string | SCA policy name |
| `pass` | number | Passing checks |
| `fail` | number | Failing checks |
| `score` | number | Compliance score (%) |
| `start_scan` | time | Last scan timestamp |

### Time series format (historical — from indexer)

| Field | Type | Description |
|-------|------|-------------|
| `Time` | time | Bucket timestamp |
| `SCA events` | number | Scan events per interval |

---

## Agent status (`dataType: agents`)

### Table format

| Field | Type | Wazuh source field | Notes |
|-------|------|--------------------|-------|
| `agent` | string | `name` | **Link target for `$agent` variable** |
| `agent_id` | string | `id` | |
| `status` | string | `status` | `active` / `disconnected` / `never_connected` |
| `ip` | string | `ip` | Use to correlate with Prometheus `instance` |
| `os` | string | `os.name` | |
| `os_version` | string | `os.version` | |
| `version` | string | `version` | Wazuh agent version |
| `last_keepalive` | time | `lastKeepAlive` | |

---

## Linking Wazuh to Prometheus / Loki

### Agent → Prometheus node exporter

The Wazuh `agent` field maps to the Prometheus `instance` label, but with different formats:

| Wazuh `agent` | Prometheus `instance` (typical) |
|---------------|----------------------------------|
| `web-01` | `web-01:9100` |
| `k8s-node-1` | `10.0.0.5:9100` |

**Recommended approach:** Use the `$agent` variable with a Prometheus label regex:

```
node_cpu_seconds_total{instance=~"$agent.*"}
```

The `.*` suffix handles the port suffix discrepancy.

### Namespace → kube-state-metrics

When Wazuh's Kubernetes integration is active, the `$namespace` variable (from the Wazuh datasource) lists distinct namespaces from Wazuh alert data. In Prometheus:

```
kube_pod_info{namespace="$namespace"}
```

### Example: mixed dashboard variable

See the bundled **Correlation with Prometheus (Example)** dashboard (`wazuh-mixed-prometheus-example`) for a working reference. It demonstrates:

- `$agent` variable from Wazuh datasource filtering both node exporter and Wazuh alert panels
- `$namespace` variable from Wazuh datasource (populated when k8s integration is active)
- Shared time range applied to all panels

### Field naming conventions

The plugin uses these normalized names consistently so Grafana transformations (join, merge) work without field renaming:

| Concept | Plugin field | Prometheus label | Loki label |
|---------|-------------|------------------|------------|
| Agent/node name | `agent` | `instance` (partial) | `host` |
| Agent IP | `ip` | `instance` (host part) | — |
| Kubernetes namespace | `$namespace` variable | `namespace` | `namespace` |
| Severity (numeric) | `severity_level` | — | — |
| Severity (string) | `severity` | — | — |

---

## Tips for production dashboards

1. **Use `$agent` as a single-value variable** (not multi-select) when linking to Prometheus — regex matching works best with one value.
2. **Pin the Wazuh datasource UID to `wazuh`** in provisioning to ensure bundled dashboards resolve correctly.
3. **The `$namespace` variable returns an empty list** on non-k8s Wazuh deployments — this is expected; the variable simply has no values and panels show all data.
4. **Time range is always inherited from Grafana** — you do not need to pass it in Wazuh query filters; the backend applies it automatically.
