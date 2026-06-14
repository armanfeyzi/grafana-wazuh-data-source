# Wazuh Datasource for Grafana

Open-source Grafana plugin that connects to Wazuh (manager API + indexer) so security data appears alongside your existing dashboards — without manual OpenSearch configuration.

**Latest release:** [v0.2.8](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/latest)

## Guides

| Guide | Description |
|-------|-------------|
| [Installation](installation/) | Install plugin, configure datasource, import dashboards |
| [Development](development/) | Local plugin development loop |
| [Kubernetes](kubernetes/) | Production / in-cluster Wazuh |
| [RBAC](rbac/) | Minimum API + indexer permissions |
| [Field mapping](field-mapping/) | Normalized fields ↔ Prometheus/Loki labels |
| [Signing & catalog](signing/) | Plugin signing and Grafana catalog submission |
| [Reviewer quickstart](reviewer-quickstart/) | 15-minute path for Grafana plugin reviewers |
| [Status](status/) | What's done and what's next |
| [Roadmap](project-roadmap/) | Phases and architecture |

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
