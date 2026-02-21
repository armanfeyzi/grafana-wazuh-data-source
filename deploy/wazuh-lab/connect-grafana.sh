#!/usr/bin/env bash
set -euo pipefail

# Attach the Grafana dev container to the Wazuh lab Docker network.
# Run after: make dev (or compose up) AND make lab-up

CONTAINER="${GRAFANA_CONTAINER:-wazuh-datasource}"
NETWORK="${WAZUH_NETWORK:-single-node_default}"

if ! docker ps --format '{{.Names}}' | grep -qx "${CONTAINER}"; then
  echo "Grafana container '${CONTAINER}' is not running."
  echo "Start it first: make dev"
  exit 1
fi

if docker network inspect "${NETWORK}" >/dev/null 2>&1; then
  docker network connect "${NETWORK}" "${CONTAINER}" 2>/dev/null || true
  echo "Connected ${CONTAINER} to ${NETWORK}."
else
  echo "Network '${NETWORK}' not found. Start the lab first: make lab-up"
  exit 1
fi

echo ""
echo "Next steps:"
echo "  1. Mount deploy/wazuh-lab/examples/datasource.yaml.example in Grafana provisioning, or"
echo "  2. Add the datasource in the UI with:"
echo "       Manager URL : https://wazuh.manager:55000"
echo "       Indexer URL : https://wazuh.indexer:9200"
echo "  3. Save & Test, then open bundled dashboards (datasource uid must be: wazuh)"
