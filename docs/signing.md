# Grafana Plugin Signing Guide

This document explains how Grafana plugin signing works for public and private plugins, how the release pipeline signs builds, and how to test the plugin before catalog listing.

---

## 1. Release pipeline signing

After Grafana catalog approval, release builds are signed automatically via `grafana/plugin-actions/package-plugin` using a Grafana Cloud Access Policy token:

```yaml
      - name: Build and package plugin
        id: build
        uses: grafana/plugin-actions/package-plugin@package-plugin/v1.2.0
        with:
          go-version: '1.26.5'
          node-version: '22'
          policy_token: ${{ secrets.GRAFANA_ACCESS_POLICY_TOKEN }}
```

Create the token at Grafana Cloud → **My Account → Security → Access Policies** with scope **`plugins:write`**, then store it as the GitHub Actions secret `GRAFANA_ACCESS_POLICY_TOKEN`.

Release builds also run `govulncheck` and attach a [GitHub build provenance attestation](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations) to the ZIP.

### Historical note (pre-approval)

Until the plugin was approved for the catalog, public signing returned `409 Conflict`. Unsigned ZIPs were submitted for review. That restriction is lifted after approval; new releases must include `MANIFEST.txt`.

* **Public / community plugins:** Sign with `plugins:write` (no `rootUrls`).
* **Private plugins:** Sign with `--rootUrls` pointing at your Grafana instance roots.

---

## 2. How to Test the Unsigned Plugin in your Kubernetes Cluster

To test the unsigned plugin inside your Kubernetes cluster (e.g. using the official Grafana Helm Chart or Kubernetes manifests), you must tell Grafana to allow loading unsigned plugins.

### Option A: Via Helm Chart (Recommended)
If you deploy Grafana using the standard Helm chart, update your `values.yaml` to include the plugin ID under `allowLoadingUnsignedPlugins`:

```yaml
env:
  GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS: "armanfeyzi-wazuh-datasource"
```

Or under the plugins configuration section:

```yaml
grafana.ini:
  plugins:
    allow_loading_unsigned_plugins: armanfeyzi-wazuh-datasource
```

### Option B: Via Environment Variables in Deployment Manifests
If you are deploying Grafana using standard Kubernetes deployment manifests, add the following environment variable to the Grafana container spec:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
spec:
  template:
    spec:
      containers:
        - name: grafana
          env:
            - name: GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS
              value: "armanfeyzi-wazuh-datasource"
```

---

## 3. Grafana Plugin Catalog

Community plugins are signed by Grafana after catalog approval. Until then, distribute the unsigned ZIP from [GitHub Releases](https://github.com/armanfeyzi/grafana-wazuh-data-source/releases).

Publishers submit through [Grafana Cloud → My Plugins](https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin). Requirements:

- Plugin ID prefix matches your Grafana Cloud organization slug (`armanfeyzi`)
- Release ZIP passes the [plugin validator](https://plugin-validator.grafana.net)
- Public source repository with `CHANGELOG.md`, `LICENSE`, and documentation

After approval, Grafana signs the plugin and lists it at `grafana.com/grafana/plugins/armanfeyzi-wazuh-datasource`. Users can then install without `allow_loading_unsigned_plugins`.

### Submission form fields

When submitting or updating via Grafana Cloud → My Plugins, use **separate URLs** for the packaged plugin and the source repository:

| Field | Example for version `0.2.10` |
|-------|-----------------------------|
| **URL** (plugin ZIP) | `https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/download/v0.2.10/armanfeyzi-wazuh-datasource-0.2.10.zip` |
| **Source code URL** | `https://github.com/armanfeyzi/grafana-wazuh-data-source/tree/v0.2.10` |
| **SHA1** | Contents of `armanfeyzi-wazuh-datasource-0.2.10.zip.sha1` from the same GitHub release |

Do **not** use the plugin ZIP URL as the source code URL. The source field must point at the Git repository (tagged release branch) so Grafana can run `govulncheck` against `go.mod` and verify GitHub build provenance attestation.

Release builds are produced by GitHub Actions (`.github/workflows/release.yml`) with `govulncheck` and [build provenance attestation](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations) enabled.

Regenerate catalog screenshots:

```bash
GRAFANA_URL=https://grafana.example.com \
GRAFANA_USER=admin GRAFANA_PASSWORD=secret \
node scripts/capture-catalog-screenshots.mjs
npm run build
```
