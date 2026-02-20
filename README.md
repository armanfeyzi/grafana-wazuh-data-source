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

The plugin loads from `dist/` via Docker Compose. Provisioning example is in `provisioning/datasources/`.

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
