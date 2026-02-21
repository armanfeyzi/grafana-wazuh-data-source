# RBAC guide

Minimum required permissions for the Wazuh datasource plugin to function correctly. Use these as the baseline; your security policy may restrict further.

---

## Wazuh Manager API

The plugin authenticates to the Manager API via JWT (`POST /security/user/authenticate`). The user needs read access to the resources the plugin queries.

### Minimum API roles

| Resource | Permission | Used for |
|----------|-----------|----------|
| `agent:read` | Read agent list and details | Agent status panels, `$agent` variable, SCA live scores |
| `sca:read` | Read SCA scan results | SCA table panels (live scores via `/sca/{agent_id}`) |

### Creating a least-privilege API user

In the Wazuh UI (**Server management → Security → Users**) or via API:

```bash
# Create a read-only role
curl -u admin:admin -X POST "https://MANAGER:55000/security/roles" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "grafana-readonly",
    "policies": [
      {"id": 1},   # agent:read
      {"id": 10}   # sca:read
    ]
  }'

# Create the user and assign the role
curl -u admin:admin -X POST "https://MANAGER:55000/security/users" \
  -H 'Content-Type: application/json' \
  -d '{"username": "grafana-wui", "password": "StrongPassword1!"}'
```

> See the [Wazuh RBAC reference](https://documentation.wazuh.com/current/user-manual/api/rbac/reference.html) for the full policy ID list.

### Using the built-in `wazuh-wui` user

Wazuh ships a pre-configured `wazuh-wui` user with broad dashboard read access. This works for development but has more permissions than strictly needed. For production, create a dedicated user with only `agent:read` and `sca:read`.

---

## Wazuh Indexer (OpenSearch)

The plugin connects to the indexer using Basic Auth. The indexer user needs read access to Wazuh data indices.

### Minimum index permissions

| Index pattern | Permission | Used for |
|--------------|-----------|----------|
| `wazuh-alerts-*` | `read` | Alerts, FIM events, SCA history |
| `wazuh-states-vulnerabilities-*` | `read` | Vulnerability panels |

### Creating a least-privilege indexer user

In OpenSearch Dashboards (**Security → Internal users**) or via the REST API:

```json
PUT _plugins/_security/api/roles/grafana-readonly
{
  "index_permissions": [
    {
      "index_patterns": ["wazuh-alerts-*", "wazuh-states-vulnerabilities-*"],
      "allowed_actions": ["read", "indices:data/read/search"]
    }
  ]
}
```

Then create a user and map it to this role.

### Using the built-in `admin` user

The `admin` user works for development but has superuser privileges. In production, create a dedicated indexer user with read-only access to only the Wazuh indices.

---

## Credential storage

The plugin stores credentials in Grafana's **secure JSON** (`secureJsonData`). They are:

- Encrypted at rest in the Grafana database
- Never returned to the browser after saving
- Only decrypted server-side by the Go backend plugin

Do not store credentials in plain text in `provisioning/datasources/` files. Use Grafana's secret injection mechanisms instead:

```yaml
# provisioning/datasources/wazuh.yaml
secureJsonData:
  password: $__env{WAZUH_API_PASSWORD}
  indexerPassword: $__env{WAZUH_INDEXER_PASSWORD}
```

---

## TLS

- Use HTTPS for both the Manager API (`:55000`) and Indexer (`:9200`) in production
- The plugin defaults to **verifying TLS certificates**
- "Skip TLS verify" is provided for self-signed certificates in development — a visible warning is shown in the config UI when enabled
- In production, configure your Wazuh deployment with certificates signed by a trusted CA, or add your CA certificate to Grafana's trusted CA store

---

## Summary checklist

- [ ] Manager API user has `agent:read` and `sca:read`
- [ ] Indexer user has read access to `wazuh-alerts-*` and `wazuh-states-vulnerabilities-*`
- [ ] Credentials stored via `secureJsonData`, not plain text
- [ ] TLS verify enabled in production
- [ ] Datasource UID set to `wazuh` for bundled dashboard compatibility
