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

## Stop

```bash
cd deploy/wazuh/lab/wazuh-docker/single-node && docker compose down
```

## Resources

- [Wazuh Docker docs](https://documentation.wazuh.com/current/deployment-options/docker/wazuh-container.html)
