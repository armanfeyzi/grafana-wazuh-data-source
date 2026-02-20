#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LAB="${ROOT}/lab"
REPO="${LAB}/wazuh-docker"
TAG="v4.8.0"
SINGLE_NODE="${REPO}/single-node"

uses_podman() {
  docker info 2>/dev/null | grep -qi podman
}

# Rootless Podman cannot bind host ports below 1024, or raise production ulimits.
patch_for_rootless_podman() {
  local file="$1"
  awk '
    /^    ulimits:/ { skip=1; next }
    skip && /^    [a-zA-Z0-9_.-]+:/ { skip=0 }
    skip { next }
    /514:514\/udp/ { next }
    /443:5601/ { print "      - \"8443:5601\""; next }
    /- \.\// && $0 !~ /:Z/ && $0 !~ /:z/ {
      if ($0 ~ /wazuh\.yml/) {
        print $0 ":z"
      } else {
        print $0 ":Z"
      }
      next
    }
    { print }
  ' "${file}" > "${file}.tmp" && mv "${file}.tmp" "${file}"
}

generate_certs() {
  local certs_dir="${SINGLE_NODE}/config/wazuh_indexer_ssl_certs"
  local certs_yml="${SINGLE_NODE}/config/certs.yml"

  mkdir -p "${certs_dir}"
  chmod -R a+rX "${SINGLE_NODE}/config"

  echo "==> Generating indexer certificates (first run)..."

  if uses_podman; then
    podman run --rm \
      -v "${certs_dir}:/certificates:Z" \
      -v "${certs_yml}:/config/certs.yml:Z" \
      docker.io/wazuh/wazuh-certs-generator:0.0.2
  else
    docker compose -f generate-indexer-certs.yml run --rm generator
  fi
}

echo "==> Wazuh local lab (${TAG})"

current_map_count="$(sysctl -n vm.max_map_count 2>/dev/null || echo 0)"
if [[ "${current_map_count}" -lt 262144 ]]; then
  echo "==> vm.max_map_count is ${current_map_count}; Wazuh indexer needs 262144"
  echo "    Run: sudo sysctl -w vm.max_map_count=262144"
  exit 1
fi

mkdir -p "${LAB}"

if [[ ! -d "${REPO}/.git" ]]; then
  echo "==> Cloning wazuh-docker ${TAG}..."
  git clone --depth 1 --branch "${TAG}" https://github.com/wazuh/wazuh-docker.git "${REPO}"
fi

cd "${SINGLE_NODE}"

if [[ ! -f config/wazuh_indexer_ssl_certs/root-ca.pem ]]; then
  generate_certs
fi

COMPOSE_FILE="docker-compose.yml"
DASHBOARD_PORT="443"
if uses_podman; then
  echo "==> Podman detected — patching compose for rootless + SELinux..."
  cp docker-compose.yml docker-compose.podman.yml
  patch_for_rootless_podman docker-compose.podman.yml
  # Shared SELinux label on wazuh.yml bind mount (see fix-dashboard-perms.sh).
  sed -i 's|/wazuh/config/wazuh.yml:Z|/wazuh/config/wazuh.yml:z|g' docker-compose.podman.yml
  COMPOSE_FILE="docker-compose.podman.yml"
  DASHBOARD_PORT="8443"
fi

fix_dashboard_config_permissions() {
  local wazuh_yml="${SINGLE_NODE}/config/wazuh_dashboard/wazuh.yml"
  [[ -f "${wazuh_yml}" ]] || return 0

  # Rootless Podman + :Z bind mounts can leave MCS categories that block the
  # dashboard process from reading/writing wazuh.yml (EACCES in the UI).
  # Keep host ownership as the lab user — do NOT chown from inside the container
  # (that maps to an unprivileged subuid on the host and breaks chmod/edit).
  chcon -t container_file_t -l s0 "${wazuh_yml}" 2>/dev/null || true
  chmod 666 "${wazuh_yml}" 2>/dev/null || {
    echo "    Could not chmod ${wazuh_yml} — if owned by a numeric UID, run:"
    echo "      sudo chown \$(id -un):\$(id -gn) ${wazuh_yml}"
    echo "    or delete and recreate the file, then re-run this script."
  }
}

wait_for_manager_api() {
  echo "==> Waiting for Wazuh manager API..."
  for _ in $(seq 1 36); do
    if curl -sk -m 3 -u 'wazuh-wui:MyS3cr37P450r.*-' -X POST \
      'https://127.0.0.1:55000/security/user/authenticate?raw=true' 2>/dev/null \
      | grep -qE '^[A-Za-z0-9._-]+$'; then
      echo "    Manager API is up."
      return 0
    fi
    sleep 5
  done
  echo "    Manager API not responding yet — check: podman logs single-node_wazuh.manager_1"
  return 1
}

if [[ "${1:-}" == "reset" ]]; then
  echo "==> Resetting Wazuh lab (removes volumes — all Wazuh data in volumes is lost)..."
  docker compose -f "${COMPOSE_FILE}" down -v || true
  echo "==> Starting fresh stack..."
  docker compose -f "${COMPOSE_FILE}" up -d
  if uses_podman; then
    sleep 5
    fix_dashboard_config_permissions
  fi
  wait_for_manager_api || true
  cat <<EOF

Reset complete. Dashboard: https://localhost:${DASHBOARD_PORT}
Click "Check connection" in the Wazuh app settings — status should show Online.

EOF
  exit 0
fi

echo "==> Starting Wazuh stack..."
docker compose -f "${COMPOSE_FILE}" up -d

if uses_podman; then
  sleep 5
  echo "==> Fixing wazuh.yml permissions for dashboard (Podman)..."
  fix_dashboard_config_permissions
fi

cat <<EOF

Wazuh is starting (first boot can take 1–2 minutes).

Endpoints on the host:
  Manager API : https://localhost:55000
  Indexer     : https://localhost:9200
  Dashboard   : https://localhost:${DASHBOARD_PORT}

Default credentials (wazuh-docker ${TAG} single-node):
  Manager API : wazuh-wui / MyS3cr37P450r.*-
  Indexer     : admin / SecretPassword

Grafana datasource (plugin runs inside a container):
  Manager URL       : https://host.containers.internal:55000
  Indexer URL       : https://host.containers.internal:9200
  API username      : wazuh-wui
  API password      : MyS3cr37P450r.*-
  Indexer username  : admin
  Indexer password  : SecretPassword
  Skip TLS verify   : on

Verify API:
  curl -k -u 'wazuh-wui:MyS3cr37P450r.*-' -X POST \\
    'https://127.0.0.1:55000/security/user/authenticate?raw=true'

Verify indexer:
  curl -k -u 'admin:SecretPassword' 'https://127.0.0.1:9200/_cluster/health'

Stop lab:
  cd ${SINGLE_NODE} && docker compose -f ${COMPOSE_FILE} down

EOF
