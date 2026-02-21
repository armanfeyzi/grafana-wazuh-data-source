# Deployment Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Separate plugin code from environment-specific deployment so local dev, optional Wazuh lab, and Kubernetes production are clean, documented paths.

**Architecture:** Plugin-only default dev (`deploy/dev/`). Optional Wazuh lab isolated under `deploy/wazuh-lab/`. K8s examples under `deploy/kubernetes/`. Provisioning templates in `provisioning/examples/` only.

**Tech Stack:** Docker Compose, Grafana provisioning YAML, Kustomize examples, Make targets

**Spec:** [2026-05-22-deployment-architecture-design.md](../specs/2026-05-22-deployment-architecture-design.md)

---

## Phase D1 — Restructure

### Task 1: Create `deploy/dev/` Grafana-only compose

**Files:**
- Create: `deploy/dev/docker-compose.yaml`
- Create: `deploy/dev/.env.example`
- Create: `deploy/dev/README.md`
- Modify: `.config/docker-compose-base.yaml` (no change — reuse via extends)

- [ ] Copy root compose to `deploy/dev/` without Wazuh network or extra_hosts
- [ ] Mount only `./dist` and optional `deploy/dev/provisioning/` if user opts in
- [ ] Remove root `docker-compose.yaml` or replace with pointer

### Task 2: Move Wazuh lab to `deploy/wazuh-lab/`

**Files:**
- Move: `deploy/wazuh/*` → `deploy/wazuh-lab/*`
- Create: `deploy/wazuh-lab/connect-grafana.sh`
- Create: `deploy/wazuh-lab/examples/datasource.yaml.example`

- [ ] Update internal paths in `setup.sh`
- [ ] Keep Podman/SELinux docs in lab README only

### Task 3: Move provisioning to examples

**Files:**
- Move: `provisioning/datasources/datasources.yml` → `provisioning/examples/datasources.yaml.example`
- Move: `provisioning/dashboards/dashboards.yml` → `provisioning/examples/dashboards.yaml.example`
- Remove: auto-mount from dev compose

### Task 4: Add `deploy/kubernetes/` skeleton

**Files:**
- Create: `deploy/kubernetes/README.md`
- Create: `deploy/kubernetes/kustomization.yaml`
- Create: `deploy/kubernetes/configmap-datasource.yaml`
- Create: `deploy/kubernetes/secret-datasource.yaml.example`

### Task 5: Trim root README + add doc stubs

**Files:**
- Modify: `README.md`
- Create: `docs/development.md`
- Create: `docs/installation.md`
- Create: `docs/kubernetes.md`

---

## Phase D2 — Dashboard & provisioning cleanup

### Task 6: Verify dashboards use UID `wazuh`, no `${datasource}` variable

**Files:** `src/dashboards/*.json`

- [ ] Confirm all panels reference `"uid": "wazuh"`
- [ ] Remove datasource template variable if present
- [ ] Rebuild dist

---

## Phase D3 — Optional lab ergonomics

### Task 7: Makefile targets

**Files:** Create `Makefile`

- [ ] `make dev` — build + start Grafana
- [ ] `make lab-up` / `make lab-down` — Wazuh lab
- [ ] `make lab-connect` — attach Grafana to lab network

---

## Phase D4 — Verification

### Task 8: Verify CI still passes

- [ ] `npm run typecheck && npm run test:ci && go test ./...`

---

## Phase D5 — Update roadmap reference

- [ ] Add deployment hygiene note to `docs/project-roadmap.md` Phase 7
