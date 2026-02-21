#!/usr/bin/env bash
# Start kubectl port-forwards for Wazuh manager + indexer (bind 0.0.0.0 for Podman/Docker Grafana).
set -euo pipefail

NAMESPACE="${WAZUH_NAMESPACE:-wazuh}"
MANAGER_PORT="${WAZUH_MANAGER_PORT:-55000}"
INDEXER_PORT="${WAZUH_INDEXER_PORT:-9200}"

start_forward() {
  local svc="$1"
  local local_port="$2"
  local remote_port="$3"

  if ss -tln 2>/dev/null | rg -q ":${local_port}\\b"; then
    echo "OK  ${svc}:${remote_port} already forwarded on 0.0.0.0:${local_port}"
    return
  fi

  echo "Starting port-forward: ${svc} ${local_port}:${remote_port} (namespace ${NAMESPACE})"
  kubectl port-forward --address 0.0.0.0 -n "${NAMESPACE}" "svc/${svc}" "${local_port}:${remote_port}" &
  sleep 1

  if ss -tln 2>/dev/null | rg -q ":${local_port}\\b"; then
    echo "OK  ${svc} listening on 0.0.0.0:${local_port}"
  else
    echo "FAIL could not bind ${local_port} — check kubectl context and namespace" >&2
    exit 1
  fi
}

start_forward wazuh "${MANAGER_PORT}" 55000
start_forward wazuh-dev-indexer "${INDEXER_PORT}" 9200

echo ""
echo "Grafana datasource URLs (containerized dev):"
echo "  Manager: https://host.containers.internal:${MANAGER_PORT}"
echo "  Indexer: https://host.containers.internal:${INDEXER_PORT}"
echo ""
echo "Leave this terminal open, or run forwards in the background."
