# Kubernetes deployment

Use this when Wazuh is already running in your cluster (official Helm chart or equivalent) and you want Grafana to query it via this plugin.

## Overview

```text
Grafana pod  →  plugin backend  →  wazuh-manager:55000 (API)
                               →  indexer:9200 (OpenSearch)
```

No local Docker lab required. Configure URLs and credentials via provisioning.

## 1. Install the plugin in Grafana

Choose one approach:

| Method | Notes |
|--------|--------|
| Custom Grafana image | `COPY dist/` into `/var/lib/grafana/plugins/wazuh-datasource` |
| Init container | Download ZIP to a shared `emptyDir` volume |
| Manual mount | ConfigMap/volume with built plugin files |

Set `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=wazuh-datasource` until the plugin is signed/catalog-listed.

## 2. Find Wazuh service names

```bash
kubectl get svc -n wazuh
```

Typical names (adjust for your release):

| Service | Port | Use as |
|---------|------|--------|
| `wazuh-manager` or `wazuh` | 55000 | Manager URL |
| `indexer` or `wazuh-indexer` | 9200 | Indexer URL |

Example in-cluster URLs:

```text
https://wazuh-manager.wazuh.svc.cluster.local:55000
https://indexer.wazuh.svc.cluster.local:9200
```

## 3. Provision the datasource

Examples live in [`deploy/kubernetes/`](../deploy/kubernetes/).

```bash
# Edit URLs and namespace in configmap-datasource.yaml
kubectl apply -f deploy/kubernetes/configmap-datasource.yaml
```

Mount the ConfigMap in your Grafana deployment:

```yaml
volumeMounts:
  - name: wazuh-datasource
    mountPath: /etc/grafana/provisioning/datasources/wazuh.yaml
    subPath: wazuh.yaml
volumes:
  - name: wazuh-datasource
    configMap:
      name: grafana-datasource-wazuh
```

Wire passwords via Kubernetes Secret and Grafana's env expansion, or use the Grafana Operator `GrafanaDatasource` CR.

**Important:** Keep `uid: wazuh` so bundled dashboards resolve the datasource.

## 4. Credentials

Create a Secret (see `deploy/kubernetes/secret-datasource.yaml.example`). Use the same RBAC-aware users as your Wazuh deployment:

- API: e.g. `wazuh-wui` or a dedicated read-only API user
- Indexer: e.g. `admin` or a read-only OpenSearch user

## 5. TLS

In production, set `tlsSkipVerify: false` and trust the Wazuh/indexer CA inside the Grafana pod (mount CA bundle or use cert-manager).

## 6. Import dashboards

Import from the plugin bundle or mount dashboard JSON via file provisioning (`provisioning/examples/dashboards.yaml.example`).

## Local dev against cluster Wazuh

Port-forward **must bind to all interfaces** when Grafana runs in Docker/Podman (`make dev`):

```bash
make k8s-forward
# or manually:
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh 55000:55000
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh-dev-indexer 9200:9200
make dev
```

Datasource URLs (inside container):

```text
https://host.containers.internal:55000
https://host.containers.internal:9200
```

Use `127.0.0.1` only if Grafana runs directly on the host, not via `make dev`.

Verify both ports are listening: `ss -tlnp | rg ':55000|:9200'`

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Save & Test fails (manager) | Service name, port, API credentials, network policy; **manager port-forward running on 55000** |
| Save & Test fails (indexer) | Indexer URL, separate indexer credentials |
| `connection refused` on `host.containers.internal:55000` | Manager port-forward stopped — run `make k8s-forward` |
| Dashboards empty, Explore works | Datasource `uid` must be `wazuh` |
| Panels empty with `$agent` filter | Rebuild plugin (`make dev`); needs `applyTemplateVariables` support |
| No vulnerabilities / FIM / SCA | Wazuh data not indexed yet — not a Grafana wiring issue |
| Agent dropdown empty | Manager API unreachable or RBAC denied |

See [installation.md](installation.md) and Wazuh docs for module-specific data requirements.
