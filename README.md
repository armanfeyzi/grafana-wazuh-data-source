# Wazuh Datasource for Grafana

Open-source Grafana plugin that connects to Wazuh (manager API + indexer) so security data appears alongside your existing dashboards — without manual OpenSearch configuration.

## Quick start

```bash
npm install
npm run build
go run github.com/magefile/mage@latest -v build:linux
make dev    # Grafana at http://localhost:3000
```

Add the **Wazuh** datasource in the UI with your manager and indexer URLs. See [Installation Guide](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/installation.md).

## Documentation

| Guide | Audience |
|-------|----------|
| [Installation](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/installation.md) | Install plugin + configure datasource |
| [Development](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/development.md) | Local plugin hacking |
| [Kubernetes](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/kubernetes.md) | Production / in-cluster Wazuh |
| [RBAC guide](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/rbac.md) | Minimum required API + indexer permissions |
| [Field mapping](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/field-mapping.md) | Normalized field names ↔ Prometheus/Loki labels |
| [Optional Wazuh lab](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/deploy/wazuh-lab/README.md) | Local wazuh-docker stack |
| [Project brief](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/project-brief.md) | Goals and scope |
| [Status](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/status.md) | What's done and what's next |
| [Roadmap](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/project-roadmap.md) | Phases and milestones |
| [Contributing](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/CONTRIBUTING.md) | Dev setup, architecture, adding a data type |

## Requirements

- Node.js 22+, Go 1.22+
- Grafana 10.4+, Wazuh 4.7+
- Docker/Podman (local Grafana dev only)

## Bundled dashboards

Security Overview, Vulnerabilities, FIM, SCA, and Agent Status — import from the plugin or provision from `dist/dashboards/`. Require datasource **`uid: wazuh`** when using provisioning (see `provisioning/examples/`).

## Deploy paths

```text
Plugin dev     →  make dev              (deploy/dev/ — dashboards in Wazuh folder)
Optional lab   →  make lab-up           (deploy/wazuh-lab/)
Kubernetes     →  deploy/kubernetes/    (https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/docs/kubernetes.md)
```

After `make dev`, open **Dashboards → Wazuh** for bundled dashboards. Datasource UID must be **`wazuh`** (set in Connections → Data sources → Wazuh → General).

## License

Apache 2.0 — see [LICENSE](https://github.com/armanfeyzi/grafana-wazuh-data-source-plugin/blob/main/LICENSE).
