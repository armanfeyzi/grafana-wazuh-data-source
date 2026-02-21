# Installation

## Requirements

- Grafana **10.4+**
- Wazuh **4.7+** with Manager API and Indexer
- API user with `agent:read`, `sca:read` (minimum)
- Indexer user with read access to `wazuh-alerts-*`, `wazuh-states-vulnerabilities-*`

## Install the plugin

### From release ZIP (recommended)

1. Download the release asset from GitHub Releases.
2. Unzip into Grafana's plugin directory:

   ```text
   /var/lib/grafana/plugins/wazuh-datasource/
   ```

3. Allow unsigned plugins (until catalog listing):

   ```ini
   [plugins]
   allow_loading_unsigned_plugins = wazuh-datasource
   ```

4. Restart Grafana.

### From source

```bash
npm ci && npm run build
go run github.com/magefile/mage@latest -v build:linux
# Copy dist/ to Grafana plugins path
```

## Configure the datasource

In Grafana: **Connections → Data sources → Add → Wazuh**

| Field | Description |
|-------|-------------|
| Manager URL | Wazuh API base, e.g. `https://wazuh-manager.example.com:55000` |
| Indexer URL | OpenSearch base, e.g. `https://indexer.example.com:9200` |
| Username / Password | Manager API credentials |
| Indexer username / password | Indexer credentials (often different from API user) |
| Skip TLS verify | Dev/lab only — use proper CA in production |

Click **Save & test**. Both manager API and indexer must respond.

## Provision with GitOps

Copy `provisioning/examples/datasources.yaml.example` and set:

- `uid: wazuh` — **required** for bundled dashboards
- Your manager and indexer URLs
- Credentials via `secureJsonData` or Grafana secret injection

For Kubernetes, see [kubernetes.md](kubernetes.md) and `deploy/kubernetes/`.

## Import dashboards

Five dashboards ship with the plugin:

| Dashboard | UID |
|-----------|-----|
| Security Overview | `wazuh-security-overview` |
| Vulnerabilities | `wazuh-vulnerabilities` |
| File Integrity (FIM) | `wazuh-fim` |
| Security Configuration (SCA) | `wazuh-sca` |
| Agent Status | `wazuh-agent-status` |

Import via Grafana UI (Dashboards → New → Import from plugin) or file provisioning.

**Note:** Vulnerabilities, FIM, and SCA panels need matching data in Wazuh (vuln index, syscheck alerts, SCA scans). Alerts and agent status work as soon as agents are connected.

**Template variables:** Agent dropdown is populated dynamically from the Wazuh datasource. Datasource UID must be `wazuh`. See [status.md](status.md) if panels show zero data with `$agent` selected.
