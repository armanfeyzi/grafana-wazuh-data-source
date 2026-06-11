# Wazuh Datasource for Grafana

Open-source Grafana plugin that connects to Wazuh (manager API + indexer) so security data appears alongside your existing dashboards — without manual OpenSearch configuration.

## Installation

Download the latest release ZIP from [GitHub Releases](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases), extract it into your Grafana plugins directory, and restart Grafana.

Until the plugin is signed and listed in the Grafana catalog, allow the unsigned plugin ID:

```text
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=armanfeyzi-wazuh-datasource
```

See the [Installation Guide](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/installation.md) for datasource configuration, Kubernetes notes, and troubleshooting.

## Documentation

| Guide | Audience |
|-------|----------|
| [Installation](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/installation.md) | Install plugin + configure datasource |
| [Development](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/development.md) | Local plugin hacking |
| [Kubernetes](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/kubernetes.md) | Production / in-cluster Wazuh |
| [RBAC guide](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/rbac.md) | Minimum required API + indexer permissions |
| [Field mapping](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/field-mapping.md) | Normalized field names ↔ Prometheus/Loki labels |
| [Optional Wazuh lab](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/deploy/wazuh-lab/README.md) | Local wazuh-docker stack |
| [Project brief](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/project-brief.md) | Goals and scope |
| [Status](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/status.md) | What's done and what's next |
| [Roadmap](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/project-roadmap.md) | Phases and architecture |
| [Contributing](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/CONTRIBUTING.md) | Dev setup, architecture, adding a data type |
| [Reviewer quickstart](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/reviewer-quickstart.md) | Grafana catalog reviewers — 15 min test path |

## Requirements

- Grafana 10.4+
- Wazuh 4.7+ with manager API and indexer reachable from Grafana

## Data types

Query **alerts**, **vulnerabilities**, **FIM**, **SCA**, and **agent status** from the Explore query editor or bundled dashboards.

## Bundled dashboards

Security Overview, Vulnerabilities, FIM, SCA, and Agent Status ship with the plugin. When provisioning, set datasource **`uid: wazuh`** (see `provisioning/examples/` in the repository).

## License

Apache 2.0 — see [LICENSE](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/LICENSE).
