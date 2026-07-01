# Wazuh Datasource for Grafana

Connect Grafana to your Wazuh deployment so security data appears in dashboards and Explore — without manual OpenSearch configuration.

Query **alerts**, **vulnerabilities**, **file integrity monitoring (FIM)**, **SCA compliance**, and **agent status** from Wazuh's manager API and indexer.

## Requirements

- Grafana 10.4 or newer
- Wazuh 4.7 or newer with the manager API and indexer reachable from Grafana

## Getting started

### 1. Install the plugin

In Grafana, go to **Administration → Plugins**, search for **Wazuh**, and click **Install**.

### 2. Add a datasource

1. Go to **Connections → Data sources → Add data source**.
2. Select **Wazuh**.
3. Configure:
   - **Manager URL** — Wazuh manager API base URL (e.g. `https://wazuh.example.com:55000`)
   - **Indexer URL** — Wazuh indexer (OpenSearch) URL (e.g. `https://wazuh-indexer.example.com:9200`)
   - **Username** / **Password** — credentials for the manager API
   - **Indexer username** / **Indexer password** — optional; defaults to manager credentials
   - **Skip TLS verify** — enable only for lab environments with self-signed certificates
4. Click **Save & test**. A successful connection confirms both the manager API and indexer are reachable.

### 3. Explore your data

Open **Explore**, select the Wazuh datasource, choose a data type (alerts, vulnerabilities, FIM, SCA, or agent status), and run a query.

## Bundled dashboards

The plugin includes ready-made dashboards:

- Security Overview
- Vulnerabilities
- FIM
- SCA
- Agent Status

Find them under the **Wazuh** folder after installation. If you provision dashboards via YAML, set the datasource **`uid: wazuh`**.

## Further reading

- [Installation and configuration](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/installation.md)
- [RBAC permissions](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/rbac.md)
- [Kubernetes deployment](https://github.com/armanfeyzi/grafana-wazuh-data-source/blob/main/docs/kubernetes.md)
- [Source repository](https://github.com/armanfeyzi/grafana-wazuh-data-source)

## License

Apache 2.0
