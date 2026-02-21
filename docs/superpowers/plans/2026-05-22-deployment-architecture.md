# Deployment Architecture Implementation Plan

> **Status:** Complete (2026-05-22)  
> **Live status:** [../../status.md](../../status.md)

**Goal:** Separate plugin code from environment-specific deployment so local dev, optional Wazuh lab, and Kubernetes production are clean, documented paths.

**Architecture:** Plugin-only default dev (`deploy/dev/`). Optional Wazuh lab isolated under `deploy/wazuh-lab/`. K8s examples under `deploy/kubernetes/`. Provisioning templates in `provisioning/examples/` only.

**Tech Stack:** Docker Compose, Grafana provisioning YAML, Kustomize examples, Make targets

**Spec:** [2026-05-22-deployment-architecture-design.md](../specs/2026-05-22-deployment-architecture-design.md)

---

## Phase D1 — Restructure ✅

### Task 1: Create `deploy/dev/` Grafana-only compose

**Files:**
- Create: `deploy/dev/docker-compose.yaml`
- Create: `deploy/dev/.env.example`
- Create: `deploy/dev/README.md`

- [x] Grafana-only compose with plugin mount from `dist/`
- [x] `extra_hosts: host.containers.internal:host-gateway` for K8s port-forward access
- [x] Persistent Grafana volume
- [x] Optional dev provisioning (datasource + dashboards)
- [x] Removed root `docker-compose.yaml`

### Task 2: Move Wazuh lab to `deploy/wazuh-lab/`

- [x] Lab under `deploy/wazuh-lab/`
- [x] `connect-grafana.sh`, `setup.sh`
- [x] `examples/datasource.yaml.example`
- [x] Podman/SELinux docs in lab README

### Task 3: Move provisioning to examples

- [x] `provisioning/examples/` templates
- [x] Dev-specific provisioning in `deploy/dev/provisioning/` (gitignored secrets)

### Task 4: Add `deploy/kubernetes/` skeleton

- [x] README, kustomization, ConfigMap, Secret example

### Task 5: Trim root README + add doc stubs

- [x] `README.md`, `docs/development.md`, `docs/installation.md`, `docs/kubernetes.md`

---

## Phase D2 — Dashboard & provisioning cleanup ✅

### Task 6: Verify dashboards use UID `wazuh`

- [x] All panels reference `"uid": "wazuh"`
- [x] No `${datasource}` template variable
- [x] Dynamic `$agent` query variable on dashboards
- [x] `applyTemplateVariables` for panel filter interpolation

---

## Phase D3 — Optional lab ergonomics ✅

### Task 7: Makefile targets

- [x] `make dev` — build + start Grafana
- [x] `make dev-config` — create local datasource yaml
- [x] `make k8s-forward` — start both port-forwards
- [x] `make lab-up` / `make lab-down` / `make lab-connect`

---

## Phase D4 — Verification ✅

### Task 8: Verify CI still passes

- [x] `npm run typecheck`, frontend tests, `go test ./...`, `npm run build`

---

## Phase D5 — Documentation ✅

- [x] Deployment hygiene in docs
- [x] `docs/status.md` — living progress summary
- [x] Updated milestones and roadmap status

---

## Follow-up (not part of this plan)

See [status.md](../../status.md) Phase 6–7:
- Mixed Prometheus + Wazuh dashboard
- Namespace template variable
- v0.1.0 release
