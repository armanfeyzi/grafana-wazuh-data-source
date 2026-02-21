# Provisioning examples

Copy these templates into your Grafana provisioning path. They are **not** mounted by default in local dev (`deploy/dev/`).

| File | Purpose |
|------|---------|
| `datasources.yaml.example` | Generic datasource — set your URLs and credentials |
| `dashboards.yaml.example` | File provider for dashboard JSON |

For the optional Wazuh docker lab, see `deploy/wazuh-lab/examples/datasource.yaml.example`.

For Kubernetes, see `deploy/kubernetes/`.

Bundled dashboards require datasource **`uid: wazuh`**.
