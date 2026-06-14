# Grafana Plugin Signing Guide

This document explains how Grafana plugin signing works for public and private plugins, why signing is bypassed in the automated release pipeline, and how to test and run the plugin in your Kubernetes/Grafana cluster.

---

## 1. Why `policy_token` is Commented Out in CI/CD

For the release pipeline to package the plugin successfully, the `policy_token` input in `.github/workflows/release.yml` is omitted until catalog approval:

```yaml
      - name: Build and package plugin
        id: build
        uses: grafana/plugin-actions/package-plugin@package-plugin/v1.2.0
        with:
          go-version: '1.25.10'
```

Release builds install `govulncheck` and attach a [GitHub build provenance attestation](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations) to the ZIP, as required by Grafana catalog review.

### The Root Cause: API 409 Conflict
When a plugin ID is configured as a public community plugin (e.g., `armanfeyzi-wazuh-datasource`), Grafana Cloud's signing API will reject all public signature requests with a `409 Conflict / InvalidArgument` error **until the plugin has been formally approved and registered in the Grafana Plugin Catalog**.

* **For Public Submission:** You do **not** need to sign the plugin yourself. You submit the **unsigned** packaged ZIP archive from your GitHub Releases page to the Grafana review team. Once approved, Grafana's publishing system automatically signs it.
* **For Private Distribution:** If you wanted to distribute the plugin privately (not via the public catalog), you would sign it as `private` by providing the `SIGN_ROOT_URLS` setting pointing to your private Grafana instances.

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

| Field | Example for version `0.2.8` |
|-------|-----------------------------|
| **URL** (plugin ZIP) | `https://github.com/armanfeyzi/grafana-wazuh-data-source/releases/download/v0.2.8/armanfeyzi-wazuh-datasource-0.2.8.zip` |
| **Source code URL** | `https://github.com/armanfeyzi/grafana-wazuh-data-source/tree/v0.2.8` |
| **SHA1** | Contents of `armanfeyzi-wazuh-datasource-0.2.8.zip.sha1` from the same GitHub release |

Do **not** use the plugin ZIP URL as the source code URL. The source field must point at the Git repository (tagged release branch) so Grafana can run `govulncheck` against `go.mod` and verify GitHub build provenance attestation.

Release builds are produced by GitHub Actions (`.github/workflows/release.yml`) with `govulncheck` and [build provenance attestation](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations) enabled.

Regenerate catalog screenshots:

```bash
GRAFANA_URL=https://grafana.example.com \
GRAFANA_USER=admin GRAFANA_PASSWORD=secret \
node scripts/capture-catalog-screenshots.mjs
npm run build
```
