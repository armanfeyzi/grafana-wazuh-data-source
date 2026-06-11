# Reviewer quickstart

Guide for Grafana plugin reviewers to validate **armanfeyzi-wazuh-datasource** in under 15 minutes.

## What this plugin does

Connects Grafana to **Wazuh Manager API** (JWT auth) and **Wazuh Indexer** (OpenSearch) so users can query alerts, vulnerabilities, FIM, SCA, and agent status without configuring a separate OpenSearch datasource.

- **Plugin ID:** `armanfeyzi-wazuh-datasource`
- **Display name:** Wazuh
- **Recommended datasource UID:** `wazuh` (required for bundled dashboards)

## Prerequisites

| Component | Version |
|-----------|---------|
| Grafana | 10.4+ |
| Wazuh | 4.7+ with Manager API + Indexer |

## Path A — Local Wazuh Docker lab (fastest, self-contained)

From the plugin repository root:

```bash
npm ci
make lab-up          # ~5 min first run; clones wazuh-docker 4.8.0
make dev-config      # first time only
make dev             # Grafana at http://localhost:3000 (terminal 2)
make lab-connect     # attach Grafana container to lab network (terminal 3)
```

**Lab credentials** (default wazuh-docker single-node):

| Field | Value |
|-------|--------|
| Manager URL | `https://wazuh.manager:55000` |
| Indexer URL | `https://wazuh.indexer:9200` |
| API username | `wazuh-wui` |
| API password | `MyS3cr37P450r.*-` |
| Indexer username | `admin` |
| Indexer password | `SecretPassword` |
| Skip TLS verify | **on** |

Configure via **Connections → Data sources → Wazuh** or copy `deploy/wazuh-lab/examples/datasource.yaml.example`.

**Allow unsigned plugin** (required until catalog signing):

```text
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=armanfeyzi-wazuh-datasource
```

(`make dev` sets this automatically in `deploy/dev/docker-compose.yaml`.)

### Expected results (Path A)

1. **Save & test** → green: `Connected to Wazuh manager API and indexer`
2. **Dashboards → Wazuh → Security Overview** → panels show agent/alert data after agents register (~2–5 min)
3. **Explore → Wazuh** → data type **Alerts**, format **Table** → rows returned

## Path B — Kubernetes port-forward (production-like)

If you have kubectl access to a cluster running Wazuh:

```bash
npm ci && npm run build
go run github.com/magefile/mage@latest -v build:linux
make dev-config && make dev

# separate terminal — keep running:
make k8s-forward
```

Datasource URLs when Grafana runs in Docker (`make dev`):

| Field | Value |
|-------|--------|
| Manager URL | `https://host.containers.internal:55000` |
| Indexer URL | `https://host.containers.internal:9200` |
| Skip TLS verify | **on** |

Use credentials from your cluster secrets. See [development.md](./development.md) for `kubectl` examples.

## Path C — Reviewer-provided environment

The maintainer can provide temporary access to a live Grafana + Wazuh lab on request (in-cluster Wazuh 4.x, datasource UID `wazuh`, bundled dashboards provisioned). Mention this in the submission notes if needed.

## Validation checklist

| Step | Expected |
|------|----------|
| Install unsigned ZIP into `/var/lib/grafana/plugins/armanfeyzi-wazuh-datasource/` | Plugin loads after restart |
| Add datasource, **Save & test** | Status OK, both manager + indexer reachable |
| Open **Security Overview** dashboard (`uid: wazuh-security-overview`) | Stat panels and charts populate |
| Explore → data type **Agents** | Table lists registered agents |
| Explore → data type **Vulnerabilities** | Table rows (may be empty on fresh lab) |
| Bundled dashboards import | Datasource type `armanfeyzi-wazuh-datasource`, UID `wazuh` |

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Plugin not loading | Set `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=armanfeyzi-wazuh-datasource` |
| Save & test: connection refused | Manager/indexer URL wrong; for Docker dev use `host.containers.internal`, not `127.0.0.1` |
| Dashboards empty, Explore works | Datasource UID must be `wazuh` |
| Panels show "datasource not found" | Datasource `type` must be `armanfeyzi-wazuh-datasource` |
| SCA panels HTTP 429 | Reduce refresh rate; plugin caches SCA requests (v0.2.2+) |

## More documentation

- [Installation](./installation.md)
- [RBAC minimum permissions](./rbac.md)
- [Kubernetes deployment](./kubernetes.md)
- [Local Wazuh lab](../deploy/wazuh-lab/README.md)
