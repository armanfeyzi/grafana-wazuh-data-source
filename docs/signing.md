# Grafana Plugin Signing Guide

This document explains how Grafana plugin signing works for public and private plugins, why signing is bypassed in the automated release pipeline, and how to test and run the plugin in your Kubernetes/Grafana cluster.

---

## 1. Why `policy_token` is Commented Out in CI/CD

For the release pipeline to package the plugin successfully, the `policy_token` input in `.github/workflows/release.yml` has been commented out:

```yaml
      - name: Build and package plugin
        id: build
        uses: grafana/plugin-actions/build-plugin@build-plugin/v1.0.2
        with:
          go-version: '1.26'
          # policy_token: ${{ secrets.GRAFANA_ACCESS_POLICY_TOKEN }}
```

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

## 3. Workflow for Production Submission

Once you are done testing the plugin in your cluster and are ready to submit it for public review:

1. Download the unsigned `.zip` package from your latest GitHub Release.
2. Submit a Pull Request to the [Grafana Plugin Repository](https://github.com/grafana/grafana-plugin-repository) following their checklist.
3. Upon approval, the Grafana team will automatically sign, publish, and list your plugin under the official plugin catalog at `grafana.com/plugins`.
