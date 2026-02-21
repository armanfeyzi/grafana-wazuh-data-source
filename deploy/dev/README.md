# Local Grafana for plugin development

Runs **Grafana only** with the plugin mounted from `dist/`. No Wazuh stack, no auto-provisioned datasource.

## Prerequisites

- Node.js 22+, Go 1.22+, Docker or Podman
- Built plugin: `npm run build` and `go run github.com/magefile/mage@latest -v build:linux`

## Start

From the repository root:

```bash
make dev
# or:
docker compose -f deploy/dev/docker-compose.yaml up --build
```

Open http://localhost:3000.

### Datasource (auto-provisioned)

On first run, create your local config:

```bash
make dev-config
```

Edit `deploy/dev/provisioning/datasources/wazuh.yaml` with passwords from your cluster:

```bash
kubectl get secret wazuh-api-cred -n wazuh -o jsonpath='{.data.API_PASSWORD}' | base64 -d; echo
kubectl get secret indexer-cred -n wazuh -o jsonpath='{.data.INDEXER_PASSWORD}' | base64 -d; echo
```

Then `make dev`. The datasource and dashboards survive restarts (Grafana data volume + file provisioning).

Keep **both** port-forwards running (manager + indexer):

```bash
make k8s-forward
# or manually:
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh 55000:55000
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh-dev-indexer 9200:9200
```

If you see `connection refused` on `host.containers.internal:55000`, the **manager** forward is usually missing — only the indexer forward (9200) is still running.

## Connect to Wazuh

| Source | Manager URL | Indexer URL |
|--------|-------------|-------------|
| Kubernetes port-forward | `https://127.0.0.1:55000` | `https://127.0.0.1:9200` |
| Optional local lab | `https://wazuh.manager:55000` | `https://wazuh.indexer:9200` |

After starting the [optional Wazuh lab](../wazuh-lab/README.md), run `make lab-connect` so Grafana can reach `wazuh.manager` / `wazuh.indexer` on the lab Docker network.

See [docs/development.md](../../docs/development.md) for the full workflow.
