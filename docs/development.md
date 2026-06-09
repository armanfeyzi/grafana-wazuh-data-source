# Development

## Plugin-only workflow (default)

For plugin frontend/backend work. Wazuh can be anywhere — your Kubernetes cluster (port-forward), a remote server, or the optional local lab.

```bash
npm install
npm run dev          # webpack watch → dist/
go run github.com/magefile/mage@latest -v build:linux
make dev             # Grafana at http://localhost:3000
```

Or without Make:

```bash
docker compose -f deploy/dev/docker-compose.yaml up --build
```

**Configure the datasource** in Grafana (Settings → Data sources → Add → Wazuh), or copy `provisioning/examples/datasources.yaml.example` and mount it in Grafana.

### Connect to Wazuh in Kubernetes (recommended for production-like dev)

**1. Port-forward on all interfaces** (required when Grafana runs in Docker/Podman):

```bash
make k8s-forward
# or manually:
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh 55000:55000
kubectl port-forward --address 0.0.0.0 -n wazuh svc/wazuh-dev-indexer 9200:9200
```

**Both** forwards must stay running. If you see `connection refused` on port 55000, the manager forward is usually missing while the indexer (9200) still works.

**2. Configure the datasource** (`make dev` — plugin runs inside the container):

| Field | Value |
|-------|--------|
| Manager URL | `https://host.containers.internal:55000` |
| Indexer URL | `https://host.containers.internal:9200` |
| Skip TLS verify | on |

Do **not** use `127.0.0.1` (that is the container itself) or `10.89.x.1` (Podman bridge gateway — connection refused).

Credentials come from your cluster secrets (`wazuh-api-cred`, `indexer-cred`). Example:

```bash
kubectl get secret wazuh-api-cred -n wazuh -o jsonpath='{.data.API_USERNAME}' | base64 -d; echo
kubectl get secret indexer-cred -n wazuh -o jsonpath='{.data.INDEXER_USERNAME}' | base64 -d; echo
```

If Grafana runs **on the host** (not in Docker), use `https://127.0.0.1:55000` and `:9200` instead.

### First-time dev setup

```bash
make dev-config   # creates deploy/dev/provisioning/datasources/wazuh.yaml from example
# edit passwords from kubectl secrets, then:
make dev
```

Bundled dashboards appear under **Dashboards → Wazuh**. Datasource UID must be **`wazuh`**.

### Dashboard template variables

Bundled dashboards use a dynamic **Agent** variable (query type → Wazuh datasource). Panel filters reference `$agent` and `$severity`. The plugin implements:

- `CustomVariableSupport` — populates agent dropdown from Manager API
- `applyTemplateVariables` — resolves `$agent` / `$severity` before the backend query runs

If panels show zero data with `$agent` set but work when the filter is cleared, rebuild the plugin and restart Grafana (`make dev`).

See [status.md](status.md) for troubleshooting.

---

## Optional full stack (local Wazuh lab)

For contributors without cluster access. **Not required** for plugin development.

```bash
sudo sysctl -w vm.max_map_count=262144
make lab-up          # starts wazuh-docker single-node
make dev             # Grafana (separate terminal)
make lab-connect     # attach Grafana to lab network
```

Then provision using `deploy/wazuh-lab/examples/datasource.yaml.example` or add the datasource manually with `https://wazuh.manager:55000` and `https://wazuh.indexer:9200`.

Podman, SELinux, and lab troubleshooting: [deploy/wazuh-lab/README.md](../deploy/wazuh-lab/README.md).

---

## Checks

```bash
npm run typecheck
npm run lint
npm run test:ci
go run github.com/magefile/mage@latest -v test
```

---

## Architecture note

| Path | Purpose |
|------|---------|
| `src/`, `pkg/` | Plugin — no environment-specific URLs |
| `deploy/dev/` | Grafana for local hacking |
| `deploy/wazuh-lab/` | Optional Wazuh docker lab |
| `deploy/kubernetes/` | Production provisioning examples |
| `provisioning/examples/` | Copy-paste templates |

See [CONTRIBUTING.md](../CONTRIBUTING.md#project-structure) for the full layout.
