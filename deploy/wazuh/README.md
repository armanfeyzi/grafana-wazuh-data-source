# Local Wazuh lab

Official [wazuh-docker](https://github.com/wazuh/wazuh-docker) single-node stack for plugin development. Requires ~8 GB RAM and Docker/Podman.

## One-time host setup

```bash
sudo sysctl -w vm.max_map_count=262144
# persist across reboots (Fedora):
# echo 'vm.max_map_count=262144' | sudo tee /etc/sysctl.d/99-wazuh.conf && sudo sysctl --system
```

On **Fedora/Podman (rootless)**, the setup script patches the compose file to:
- Apply SELinux volume labels (`:Z`) on bind mounts
- Drop syslog port **514** and map dashboard to **8443** (ports below 1024 need root)
- Remove `ulimits` blocks (nofile/memlock cannot be raised rootless)

## Start Wazuh

```bash
./deploy/wazuh/setup.sh
```

First run clones `wazuh-docker` v4.8.0 into `deploy/wazuh/lab/` (gitignored), generates TLS certs, and starts manager + indexer + dashboard.

Wait until both respond:

```bash
curl -k -u 'wazuh-wui:MyS3cr37P450r.*-' -X POST \
  'https://127.0.0.1:55000/security/user/authenticate?raw=true'

curl -k -u 'admin:SecretPassword' 'https://127.0.0.1:9200/_cluster/health'
```

## Grafana datasource

The plugin backend runs inside the Grafana container — use `host.containers.internal`, not `localhost`.

| Field | Value |
|-------|--------|
| Manager URL | `https://host.containers.internal:55000` |
| Indexer URL | `https://host.containers.internal:9200` |
| API username | `wazuh-wui` |
| API password | `MyS3cr37P450r.*-` |
| Indexer username | `admin` |
| Indexer password | `SecretPassword` |
| Skip TLS verify | on |

Manager API and indexer use **different users** in Wazuh — set both in the config editor.

## Wazuh dashboard: `EACCES` on `wazuh.yml`

On **Fedora + rootless Podman**, the dashboard UI may show:

```text
2001 - EACCES: permission denied, open '.../wazuh/config/wazuh.yml'
```

### Prerequisites

1. **Stack must be running** — `exec` and the dashboard both need live containers:

```bash
cd deploy/wazuh/lab/wazuh-docker/single-node
docker compose -f docker-compose.podman.yml up -d
```

2. **Use the full path** (do not rely on `$WAZUH_YML` unless you exported it):

```bash
WAZUH_YML=config/wazuh_dashboard/wazuh.yml
```

### Fix permissions (on the host)

The API connection is usually **already configured** in this file — fix access, then click **Test the configuration** in the UI.

```bash
cd deploy/wazuh/lab/wazuh-docker/single-node
WAZUH_YML=config/wazuh_dashboard/wazuh.yml

chcon -t container_file_t -l s0 "${WAZUH_YML}"
chmod 666 "${WAZUH_YML}"
```

Do **not** run `chown` from inside the dashboard container — on rootless Podman that reassigns the file to a subuid (e.g. `525287`) on the host and you will get `Operation not permitted` when editing or chmod'ing from your user.

### If `chmod` fails with Operation not permitted

The file was likely chowned by a container subuid. Either reclaim it:

```bash
sudo chown "$(id -un):$(id -gn)" config/wazuh_dashboard/wazuh.yml
chmod 666 config/wazuh_dashboard/wazuh.yml
```

Or delete and recreate (directory is yours):

```bash
rm -f config/wazuh_dashboard/wazuh.yml
cat > config/wazuh_dashboard/wazuh.yml <<'EOF'
hosts:
  - 1513629884013:
      url: "https://wazuh.manager"
      port: 55000
      username: wazuh-wui
      password: "MyS3cr37P450r.*-"
      run_as: false
EOF
chcon -t container_file_t -l s0 config/wazuh_dashboard/wazuh.yml
chmod 666 config/wazuh_dashboard/wazuh.yml
docker compose -f docker-compose.podman.yml restart wazuh.dashboard
```

### `docker compose down` errors

If you see `no container with name ... found`, the stack is already stopped — that is harmless. Use the podman compose file consistently:

```bash
docker compose -f docker-compose.podman.yml down
docker compose -f docker-compose.podman.yml up -d
```

Re-running `./deploy/wazuh/setup.sh` applies the permission fix automatically after startup on Podman.

## Wazuh dashboard: API shows **Offline**

The `wazuh.yml` host `https://wazuh.manager` is **correct** for the Docker/Podman network. **Offline** almost always means the **manager container is down or crash-looping**, not bad dashboard config.

### 1. Check manager on the host

```bash
curl -k -u 'wazuh-wui:MyS3cr37P450r.*-' -X POST \
  'https://127.0.0.1:55000/security/user/authenticate?raw=true'
```

- **JWT token returned** → API is up; wait ~1 min for dashboard to catch up, then click **Check connection**.
- **Empty / connection refused** → manager is broken. Check logs:

```bash
podman ps --filter name=wazuh.manager
podman logs single-node_wazuh.manager_1 2>&1 | tail -40
```

If you see `Permission denied` on `/var/ossec/api/configuration` or `/var/ossec/etc/ossec.conf`, named volumes were corrupted (often after `chown` from inside a container on rootless Podman).

### 2. Reset the lab (fixes corrupted volumes)

This wipes Wazuh data in Docker volumes and starts clean:

```bash
./deploy/wazuh/setup.sh reset
```

Or manually:

```bash
cd deploy/wazuh/lab/wazuh-docker/single-node
docker compose -f docker-compose.podman.yml down -v
docker compose -f docker-compose.podman.yml up -d
chcon -t container_file_t -l s0 config/wazuh_dashboard/wazuh.yml
chmod 666 config/wazuh_dashboard/wazuh.yml
```

Wait until the curl test above returns a token, then refresh the dashboard and click **Check connection**.

## Generate test alerts

The single-node stack ships with **one agent only**: the manager itself (`000`). Port **1514** is the Wazuh **agent enrollment** port — it is **not SSH**. Running `ssh … -p 1514` will not create security alerts.

The manager container also does **not** monitor `/var/log/auth.log` or `sshd`, so failed-login rules never fire in this default setup.

### Quick smoke test (no extra agent)

Restart the manager to emit a known alert (`Wazuh server started.`):

```bash
podman exec single-node_wazuh.manager_1 /var/ossec/bin/wazuh-control restart
sleep 15
curl -k -u 'admin:SecretPassword' 'https://127.0.0.1:9200/wazuh-alerts-*/_count'
```

You should see the count increase. In Grafana **Explore**:

- **Data type:** Alerts
- **Format:** Table (easier to read than a single time-series point)
- **Time range:** Last 24 hours

### Realistic alerts (SSH, syscheck, etc.)

Install a Wazuh agent on your **host** (or another VM) and point it at the manager:

```bash
# Example: agent registers to manager on localhost:1514
# WAZUH_MANAGER='127.0.0.1' in /var/ossec/etc/ossec.conf
sudo systemctl enable --now wazuh-agent
```

Then trigger events on that host, e.g. a failed SSH login on port **22**:

```bash
ssh baduser@127.0.0.1
```

Verify in the indexer:

```bash
curl -k -u 'admin:SecretPassword' \
  'https://127.0.0.1:9200/wazuh-alerts-*/_search?size=5&sort=@timestamp:desc&pretty'
```

Or open **Wazuh dashboard** at `https://127.0.0.1:8443` → Security events.

## Stop

```bash
cd deploy/wazuh/lab/wazuh-docker/single-node && docker compose down
```

## Resources

- [Wazuh Docker docs](https://documentation.wazuh.com/current/deployment-options/docker/wazuh-container.html)
