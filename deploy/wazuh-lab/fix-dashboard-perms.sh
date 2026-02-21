#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SINGLE_NODE="${ROOT}/lab/wazuh-docker/single-node"
COMPOSE_FILE="docker-compose.yml"

if docker info 2>/dev/null | grep -qi podman; then
  COMPOSE_FILE="docker-compose.podman.yml"
  # Shared SELinux label for wazuh.yml — :Z gives each container a private MCS category
  # and the dashboard loses read/write access after restarts.
  if [[ -f "${SINGLE_NODE}/${COMPOSE_FILE}" ]]; then
    sed -i 's|/wazuh/config/wazuh.yml:Z|/wazuh/config/wazuh.yml:z|g' "${SINGLE_NODE}/${COMPOSE_FILE}"
  fi
fi

WAZUH_YML="${SINGLE_NODE}/config/wazuh_dashboard/wazuh.yml"
if [[ ! -f "${WAZUH_YML}" ]]; then
  echo "Missing ${WAZUH_YML} — run ./deploy/wazuh-lab/setup.sh first."
  exit 1
fi

chcon -t container_file_t -l s0 "${WAZUH_YML}" 2>/dev/null || true
chmod 666 "${WAZUH_YML}" 2>/dev/null || {
  echo "chmod failed — reclaim ownership, then re-run:"
  echo "  sudo chown \$(id -un):\$(id -gn) ${WAZUH_YML}"
  exit 1
}

echo "Fixed ${WAZUH_YML}"
ls -laZ "${WAZUH_YML}"

if podman ps --format '{{.Names}}' 2>/dev/null | grep -q 'wazuh.dashboard'; then
  cd "${SINGLE_NODE}"
  docker compose -f "${COMPOSE_FILE}" restart wazuh.dashboard
  echo "Restarted wazuh.dashboard — open https://127.0.0.1:8443 and click Check connection."
else
  echo "Dashboard container is not running. Start the lab, then retry."
fi
