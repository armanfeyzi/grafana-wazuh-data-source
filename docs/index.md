# Wazuh Datasource for Grafana

Open-source Grafana plugin that connects to Wazuh (manager API + indexer) so security data appears alongside your existing dashboards — without manual OpenSearch configuration.

**Latest release:** [v0.2.5](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/latest)

## Guides

| Guide | Description |
|-------|-------------|
| [Installation](installation.html) | Install plugin, configure datasource, import dashboards |
| [Development](development.html) | Local plugin development loop |
| [Kubernetes](kubernetes.html) | Production / in-cluster Wazuh |
| [RBAC](rbac.html) | Minimum API + indexer permissions |
| [Field mapping](field-mapping.html) | Normalized fields ↔ Prometheus/Loki labels |
| [Signing & catalog](signing.html) | Plugin signing and Grafana catalog submission |
| [Status](status.html) | What's done and what's next |
| [Roadmap](project-roadmap.html) | Phases and architecture |
| [Milestones](milestones.html) | Task checklist |

## Quick start

```bash
npm install && npm run build
go run github.com/magefile/mage@latest -v build:linux
make dev    # Grafana at http://localhost:3000
```

## Repository

- [Source on GitHub](https://github.com/armanfeyzi/grafana-wazuh-data-source)
- [Releases](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases)
- [Contributing](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/CONTRIBUTING.md)
