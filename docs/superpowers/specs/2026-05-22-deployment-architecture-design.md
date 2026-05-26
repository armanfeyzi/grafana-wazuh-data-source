# Deployment Architecture Design

> **Status:** Draft — awaiting review  
> **Date:** 2026-05-22  
> **Decision:** Approach **A (plugin-only repo)** + optional **B (full-stack lab profile)**

## Problem

Deployment concerns for one developer machine (Fedora, rootless Podman, wazuh-docker) leaked into the plugin repository root. That produced:

- Root `docker-compose.yaml` coupled to Wazuh lab network names
- Auto-mounted provisioning with demo credentials and Wazuh-docker hostnames
- OS-specific scripts and README troubleshooting in the main path
- Dashboard workarounds (`deleteDatasources`, hardcoded UID) to paper over provisioning bugs

The **plugin code** (`src/`, `pkg/`) is largely environment-agnostic. The **packaging and docs** are not. This blocks a clean open-source story and a straightforward Kubernetes deployment.

## Goals

1. **Plugin** remains a portable artifact — no URLs, credentials, or OS assumptions in core code.
2. **Local dev** — any contributor runs Grafana + plugin in minutes; Wazuh is external or optional.
3. **Production / Kubernetes** — first-class install path using provisioning + secrets; no Docker lab required.
4. **Optional full stack** — contributors who want a local Wazuh lab opt in explicitly (`make lab-up` or similar).
5. **CI** — green without a live Wazuh cluster (fixtures + unit tests only).

## Non-goals

- Replacing the Wazuh Helm chart or operating Wazuh itself
- Supporting every container runtime edge case in the default dev path
- Wazuh Cloud as a tested target in v0.1

---

## Target Repository Layout

```text
grafana-wazuh-data-source-plugin/
├── src/                          # Frontend (unchanged responsibility)
├── pkg/                          # Backend (unchanged responsibility)
├── src/dashboards/               # Portable dashboard JSON (plugin bundle)
│
├── provisioning/
│   └── examples/                 # Templates only — NOT mounted by default
│       ├── datasources.yaml.example
│       └── dashboards.yaml.example
│
├── deploy/
│   ├── dev/                      # Grafana-only local development
│   │   ├── docker-compose.yaml
│   │   ├── .env.example
│   │   └── README.md
│   │
│   ├── kubernetes/               # Production-oriented examples
│   │   ├── README.md
│   │   ├── kustomization.yaml
│   │   ├── configmap-datasource.yaml
│   │   ├── configmap-dashboards.yaml   # optional
│   │   └── secret-datasource.yaml.example
│   │
│   └── wazuh-lab/                # Optional local Wazuh (current deploy/wazuh)
│       ├── setup.sh
│       ├── fix-dashboard-perms.sh
│       ├── docker-compose.override.example.yaml
│       └── README.md             # Podman/SELinux notes live here only
│
├── docs/
│   ├── installation.md           # Plugin install (manual, Helm sidecar, etc.)
│   ├── development.md            # Local plugin hacking
│   ├── kubernetes.md             # K8s provisioning guide
│   └── project-roadmap.md        # existing
│
├── docker-compose.yaml             # REMOVED from root (or thin pointer to deploy/dev)
├── README.md                       # Short: what it is + links to docs/
└── Makefile                        # Optional: dev, lab-up, lab-down targets
```

### Boundary rules

| Layer | May contain | Must not contain |
|-------|-------------|------------------|
| `src/`, `pkg/` | Query logic, index patterns, health checks | Hostnames, K8s service names, demo passwords |
| `provisioning/examples/` | Documented placeholders | Active secrets or auto-mount in dev |
| `deploy/dev/` | Grafana compose, plugin volume mount | Wazuh service definitions |
| `deploy/wazuh-lab/` | wazuh-docker clone, Podman patches | Changes to plugin source |
| `deploy/kubernetes/` | ConfigMaps, Kustomize, Helm values | Application business logic |

---

## Deployment Modes

### Mode 1 — Plugin development (default)

**Audience:** Contributors, frontend/backend plugin work.

```bash
npm install
npm run dev
go run github.com/magefile/mage@latest -v build:linux
docker compose -f deploy/dev/docker-compose.yaml up
```

Grafana at `http://localhost:3000`. Plugin loaded from `./dist`.

**Wazuh connection:** User configures datasource manually in UI **or** copies `provisioning/examples/datasources.yaml.example` and fills in URLs. Typical sources:

- Kubernetes port-forward to manager/indexer
- Existing cluster-internal URLs (if Grafana runs in-cluster)
- Optional local lab (Mode 2)

`deploy/dev/docker-compose.yaml` mounts **only** `./dist` — no datasource provisioning by default. This avoids wrong URLs on every contributor machine.

### Mode 2 — Optional full stack (B)

**Audience:** Contributors without access to a Wazuh cluster.

```bash
make lab-up    # or: deploy/wazuh-lab/setup.sh && deploy/wazuh-lab/connect-grafana.sh
```

Two independent compose projects:

1. **Wazuh lab** — official wazuh-docker single-node (`deploy/wazuh-lab/`)
2. **Grafana dev** — `deploy/dev/docker-compose.yaml`

Connection between them is **explicit**:

- Documented `podman network connect` / compose `external` network in `deploy/wazuh-lab/README.md`
- Optional helper script `deploy/wazuh-lab/connect-grafana.sh` (not invoked automatically from plugin compose)
- Example provisioning snippet using Docker DNS names (`wazuh.manager`, `wazuh.indexer`) under `deploy/wazuh-lab/examples/`

Podman, SELinux, `vm.max_map_count`, and volume reset troubleshooting stay **only** in `deploy/wazuh-lab/README.md`.

### Mode 3 — Production / Kubernetes (primary user goal)

**Audience:** Platform teams with Wazuh already deployed (official Helm chart).

Assumptions (documented, overridable):

- Wazuh installed via [Wazuh Kubernetes documentation / Helm](https://documentation.wazuh.com/current/deployment-options/wazuh-kubernetes.html)
- Grafana installed separately (kube-prometheus-stack, Grafana Operator, or standalone Helm)
- Plugin installed via init container, sidecar, or baked Grafana image

**Provisioning contract:**

```yaml
# deploy/kubernetes/configmap-datasource.yaml
apiVersion: 1
datasources:
  - name: Wazuh
    uid: wazuh                    # stable UID — bundled dashboards depend on this
    type: armanfeyzi-wazuh-datasource
    access: proxy
    url: ""                       # plugin uses jsonData URLs
    jsonData:
      managerUrl: https://wazuh-manager.wazuh.svc:55000   # adjust to your release
      indexerUrl: https://indexer.wazuh.svc:9200
      username: ${WAZUH_API_USER}
      indexerUsername: ${WAZUH_INDEXER_USER}
      tlsSkipVerify: false        # use proper CA in production
    secureJsonData:
      password: ${WAZUH_API_PASSWORD}
      indexerPassword: ${WAZUH_INDEXER_PASSWORD}
```

Credentials from Kubernetes Secret via Grafana provisioning env expansion or Grafana Operator `GrafanaDatasource` CR.

**No** `deleteDatasources` in production examples — idempotent provisioning only.

Service name placeholders documented in `docs/kubernetes.md` with a table mapping common Helm release names to URLs.

---

## Dashboard Strategy

### Provisioning contract

Bundled dashboards use a **fixed datasource UID**: `wazuh`.

Install docs must state:

> When provisioning the Wazuh datasource, set `uid: wazuh`. Bundled dashboards reference this UID.

This is the same pattern used by many Grafana community dashboards and avoids the broken `${datasource}` variable when `current` is unset.

### Plugin bundle vs file provisioning

| Channel | Location | When used |
|---------|----------|-----------|
| Plugin `includes` in `plugin.json` | `dist/dashboards/` | Install via Grafana UI / plugin catalog — user picks datasource on import |
| File provisioning | `deploy/kubernetes/` or user ConfigMap | GitOps / K8s — auto-import with UID `wazuh` |

Remove root `provisioning/dashboards/` auto-mount from dev compose. Dashboards in dev are either:

- Imported once from plugin UI, or
- Optional `deploy/dev/provisioning/` example the user opts into

### Template variables

- **Agent / severity** — keep as custom variables until Phase 6 (dynamic agent query).
- **Remove datasource template variable** from bundled JSON — redundant when UID is fixed.

---

## Configuration Surface (unchanged in plugin)

The plugin ConfigEditor remains the single configuration UI:

| Field | Dev example | K8s example |
|-------|-------------|-------------|
| Manager URL | `https://localhost:55000` (port-forward) | `https://wazuh-manager.wazuh.svc:55000` |
| Indexer URL | `https://localhost:9200` | `https://indexer.wazuh.svc:9200` |
| TLS skip verify | `true` (lab) | `false` (prod) |
| Credentials | lab defaults | Kubernetes Secret |

No code changes required for K8s — only provisioning YAML and docs.

---

## Testing Strategy

| Layer | What | Requires live Wazuh? |
|-------|------|----------------------|
| Go unit tests | Query builders, parsers, filters | No — JSON fixtures |
| Frontend tests | queryUtils, components | No |
| CI (GitHub Actions) | lint, test, build | No |
| Manual integration | Explore + dashboards against real cluster | Yes — user's responsibility |
| Optional CI job | `integration` workflow, manual dispatch | Yes — secrets for lab URL |

Do **not** block CI on Wazuh lab availability.

---

## Migration Plan (from current repo)

### Phase D1 — Restructure (no plugin logic changes)

1. Create `deploy/dev/docker-compose.yaml` — Grafana only, mount `./dist`, no Wazuh network.
2. Move `deploy/wazuh/` → `deploy/wazuh-lab/` (update paths in scripts).
3. Move `provisioning/datasources/datasources.yml` → `provisioning/examples/datasources.yaml.example`.
4. Move `provisioning/dashboards/` → `provisioning/examples/` or `deploy/dev/examples/`.
5. Remove root `docker-compose.yaml` or replace with one-line pointer comment to `deploy/dev/`.
6. Add `deploy/kubernetes/` skeleton (ConfigMap + Secret example + README).
7. Split README: root stays short; move dev/lab/Podman content to `docs/development.md` and `deploy/wazuh-lab/README.md`.

### Phase D2 — Dashboard & provisioning cleanup

1. Ensure all dashboards use `"uid": "wazuh"` (already done).
2. Remove `deleteDatasources` from examples; document UID migration for existing installs.
3. Remove datasource template variable from dashboard JSON.
4. Update `plugin.json` includes — dashboards remain in bundle.

### Phase D3 — Optional lab ergonomics

1. Add `Makefile` targets: `dev`, `lab-up`, `lab-down`, `lab-connect`.
2. Add `deploy/wazuh-lab/examples/datasource.yaml` for lab DNS names.
3. Add `deploy/wazuh-lab/connect-grafana.sh`.

### Phase D4 — Documentation & K8s guide

1. Write `docs/installation.md` — plugin ZIP, Grafana version, unsigned plugin flag.
2. Write `docs/kubernetes.md` — Helm service names, Secret wiring, port-forward dev flow from cluster.
3. Write `docs/development.md` — plugin-only and full-stack paths.
4. Update roadmap Phase 7 to include this deployment hygiene work.

---

## What Gets Deleted or Detached

| Current path | Action |
|--------------|--------|
| Root `docker-compose.yaml` Wazuh external network | Remove — move to lab docs/script |
| Root `provisioning/datasources/datasources.yml` with secrets | Move to `examples/`; remove from dev mount |
| `deleteDatasources` in default provisioning | Remove |
| Podman/SELinux in root README | Move to `deploy/wazuh-lab/README.md` only |
| `extra_hosts: host.containers.internal` in default dev | Remove unless documented opt-in for Docker Desktop users |

---

## Success Criteria

1. New contributor clones repo, runs `deploy/dev` compose, sees plugin in Grafana **without** starting Wazuh.
2. User with existing K8s Wazuh follows `docs/kubernetes.md` and has working Save & Test in under 10 minutes.
3. No file in `src/` or `pkg/` references Docker/Podman/K8s hostnames.
4. CI passes with zero external services.
5. Optional lab path documented and isolated — failures do not affect plugin development.

---

## Open Items (resolve during implementation)

1. **Exact K8s service names** — verify against official Wazuh Helm chart for v4.8; document overrides in `deploy/kubernetes/README.md`.
2. **Grafana plugin install in K8s** — document two options: (a) custom Grafana image with plugin, (b) initContainer + emptyDir mount.
3. **Separate API vs indexer credentials** — v1.1; examples use same secret for now with RBAC note.

---

## Approval

Once this spec is approved, the next step is an implementation plan at:

`docs/superpowers/plans/2026-05-22-deployment-architecture.md`
