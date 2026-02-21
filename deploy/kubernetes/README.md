# Wazuh datasource — Kubernetes examples

Examples for Grafana already running in your cluster with Wazuh deployed via the [official Wazuh Kubernetes/Helm documentation](https://documentation.wazuh.com/current/deployment-options/wazuh-kubernetes.html).

**Adjust service names** to match your Helm release and namespace.

## Typical service URLs

| Component | Example URL (in-cluster) |
|-----------|--------------------------|
| Manager API | `https://wazuh-manager.wazuh.svc.cluster.local:55000` |
| Indexer | `https://indexer.wazuh.svc.cluster.local:9200` |

Run `kubectl get svc -n wazuh` (or your namespace) to confirm names.

## Install steps

1. **Install the plugin** in Grafana (custom image, init container, or volume mount of the plugin ZIP).

2. **Create a Secret** from the example:

   ```bash
   cp secret-datasource.yaml.example secret-datasource.yaml
   # edit base64 or use kubectl create secret ...
   kubectl apply -f secret-datasource.yaml
   ```

3. **Apply datasource ConfigMap** (edit URLs first):

   ```bash
   kubectl apply -f configmap-datasource.yaml
   ```

4. **Mount provisioning** in your Grafana deployment so it reads:
   - `/etc/grafana/provisioning/datasources/wazuh.yaml` (from ConfigMap)
   - Credentials via Grafana's env or secret mounting

   Or use the [Grafana Operator](https://grafana.github.io/grafana-operator/) `GrafanaDatasource` CR instead of file provisioning.

5. **Import dashboards** from the plugin UI, or provision JSON from the plugin bundle (`dist/dashboards/`).

Bundled dashboards require datasource **`uid: wazuh`** (set in `configmap-datasource.yaml`).

## Kustomize

```bash
kubectl apply -k deploy/kubernetes/
```

Edit `configmap-datasource.yaml` before applying.

## Dev against cluster Wazuh

Port-forward manager and indexer, then use `https://127.0.0.1:55000` and `https://127.0.0.1:9200` in the Grafana UI with **Skip TLS verify** enabled.

See [docs/kubernetes.md](../../docs/kubernetes.md) for the full guide.
