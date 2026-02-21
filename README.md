# Wazuh datasource for Grafana

Grafana plugin that connects to Wazuh (manager API + indexer) so security data shows up alongside your existing dashboards.

## Requirements

- Node.js 22+
- Go 1.22+
- Docker (for local Grafana)

## Development

```bash
npm install
npm run dev          # frontend watch build → dist/
go run github.com/magefile/mage@latest -v build:linux   # backend binary → dist/
docker compose up    # Grafana at http://localhost:3000
```

Build `dist/` before starting Grafana. After code changes, keep `npm run dev` running and rebuild the backend when Go files change (`go run github.com/magefile/mage@latest -v build:linux`), then restart Grafana or the plugin process.

**Save & Test** checks connectivity to both the Wazuh manager API (JWT auth) and the indexer (`/_cluster/health`). Use real URLs and credentials from your Wazuh deployment; enable **Skip TLS verify** for self-signed certificates.

When Grafana runs in Docker/Podman, the plugin backend also runs **inside the container**. Do not use `localhost` for Wazuh on your machine.

For the **local Wazuh lab**, start `./deploy/wazuh/setup.sh` first, then `docker compose up`. Grafana joins the Wazuh Docker network and provisioning uses `https://wazuh.manager:55000` and `https://wazuh.indexer:9200`.

On **rootless Podman**, `host.containers.internal` often resolves to a host-gateway address that cannot reach published ports (`connection refused`). Prefer the Wazuh service hostnames on a shared network, or `host.docker.internal` on Docker Desktop.

Before Phase 1, Save & Test only checked that required fields were set (always green). It now performs a real connection test, so you need Wazuh reachable at the configured URLs.

### Local Wazuh lab

For development without a remote cluster, run the official single-node stack:

```bash
sudo sysctl -w vm.max_map_count=262144
chmod +x deploy/wazuh/setup.sh
./deploy/wazuh/setup.sh
```

See [deploy/wazuh/README.md](deploy/wazuh/README.md) for credentials and Grafana settings. Manager API and indexer use different users (`wazuh-wui` vs `admin`).

On Fedora with Podman, if image pulls fail, ensure `docker.io` is allowed in `/etc/containers/registries.conf` or pull manually:

```bash
podman pull docker.io/grafana/grafana-enterprise:12.4.0
```

The plugin loads from `dist/` via Docker Compose. Provisioning example is in `provisioning/datasources/`.

### Bundled dashboards

Five dashboards ship with the plugin (also auto-provisioned in local dev under the **Wazuh** folder):

| Dashboard | UID | Contents |
|-----------|-----|----------|
| Security Overview | `wazuh-security-overview` | Alert volume, latest events, agent count |
| Vulnerabilities | `wazuh-vulnerabilities` | CVE table, severity filter, detection trend |
| File Integrity (FIM) | `wazuh-fim` | Syscheck events over time and by path |
| Security Configuration (SCA) | `wazuh-sca` | Live scores (API) + historical scans |
| Agent Status | `wazuh-agent-status` | Full agent inventory |

Template variables: **datasource**, **agent** (where applicable), **severity** (vulnerabilities). Edit the agent variable options to match your agent names (dynamic agent variables are Phase 6).

After `npm run build`, dashboards are in `dist/dashboards/`. Restart Grafana to pick them up.

## Checks

```bash
npm run typecheck
npm run lint
npm run test:ci
npm run build
go run github.com/magefile/mage@latest -v build:linux
go run github.com/magefile/mage@latest -v test
```

## Docs

- [Project brief](project-brief.md)
- [Roadmap](docs/project-roadmap.md)
- [Milestones](docs/milestones.md)

## License

Apache 2.0 — see [LICENSE](LICENSE).
