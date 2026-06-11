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
npm ci && npm run build
go run github.com/magefile/mage@latest -v build:linux
make lab-up          # first boot ~1–2 min; see Podman note below
make lab-dev-config  # provisions lab URLs (not the Kubernetes dev-config)
make dev             # Grafana at http://localhost:3000 (keep running)
make lab-connect     # separate terminal — attach Grafana to lab network
```

**Podman / Fedora:** If the manager API does not respond after `make lab-up`, run `make lab-reset` and wait until the script prints `Manager API is up.` Verify on the host:

```bash
curl -k -u 'wazuh-wui:MyS3cr37P450r.*-' -X POST \
  'https://127.0.0.1:55000/security/user/authenticate?raw=true'
curl -k -u 'admin:SecretPassword' 'https://127.0.0.1:9200/_cluster/health'
```

**Lab credentials** (default wazuh-docker single-node, also in `deploy/wazuh-lab/examples/datasource.yaml.example`):

| Field | Value |
|-------|--------|
| Manager URL | `https://wazuh.manager:55000` |
| Indexer URL | `https://wazuh.indexer:9200` |
| API username | `wazuh-wui` |
| API password | `MyS3cr37P450r.*-` |
| Indexer username | `admin` |
| Indexer password | `SecretPassword` |
| Skip TLS verify | **on** |

`make lab-dev-config` writes `deploy/dev/provisioning/datasources/wazuh.yaml` with these values. Do **not** use `make dev-config` for Path A — that file targets Kubernetes port-forward URLs.

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
