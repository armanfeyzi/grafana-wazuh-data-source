# Wazuh Data Source Plugin for Grafana — Project Brief

## Background

Modern security teams running Kubernetes-based infrastructure typically operate
multiple security tools in parallel — each covering a different layer of the stack.
Tools like Wazuh (SIEM/XDR), Trivy (container scanning), Kyverno (policy enforcement),
SonarQube (code quality), and OWASP ZAP (application testing) each do their job well.
But they are isolated. Every tool has its own UI, its own login, its own concept of
what a "finding" looks like, and its own way of presenting data.

This means a security engineer or platform engineer has to context-switch across
multiple dashboards to get a complete picture of what is happening in the cluster.
There is no single place to ask: "what is the overall security posture right now?"

Grafana is already widely adopted for infrastructure observability. Most teams running
Kubernetes already have Grafana open in a browser tab showing Prometheus metrics,
node health, and resource usage. It is the natural home for a unified security view —
but it currently has no meaningful native integration with Wazuh, the most widely
used open-source SIEM in this space.


## The Problem

Wazuh stores all its data — security alerts, vulnerability findings, File Integrity
Monitoring (FIM) events, Security Configuration Assessment (SCA) results, and agent
status — in an OpenSearch indexer backend. This data is rich and actionable, but
today it is only accessible in two ways:

- The Wazuh built-in dashboard (a separate UI, separate login, separate mental model)
- Manually, by configuring the generic Grafana OpenSearch datasource to point at
  Wazuh's indices — which requires knowing Wazuh's internal index naming conventions,
  field schema, and query patterns from memory

Neither option is practical for day-to-day use. The Wazuh UI is isolated from
infrastructure metrics. The manual OpenSearch approach requires significant upfront
configuration work, produces no pre-built dashboards, and has to be repeated from
scratch on every new Grafana instance.

There is currently no dedicated Grafana plugin for Wazuh.


## Our Idea

We want to build a dedicated **Wazuh datasource plugin for Grafana** — an
open-source plugin that makes Wazuh a first-class citizen in Grafana, the same
way Prometheus or Loki are today.

The plugin would connect Grafana directly to a Wazuh deployment and expose its
data in a way that feels native to Grafana: queryable, composable with other
datasources, and ready to use without manual index configuration.

The core idea is simple: a security engineer should be able to open Grafana,
add "Wazuh" as a datasource (just like they add Prometheus), and immediately
start building panels or loading pre-built dashboards — without needing to know
anything about OpenSearch index patterns or Wazuh's internal data structure.


## Who This Is For

The primary users are platform engineers and security engineers who:

- Run Wazuh on a Kubernetes cluster (or on-premise) with a significant agent deployment
- Already use Grafana for infrastructure observability (Prometheus, Loki, etc.)
- Want to see security events and infrastructure health in the same place
- Are tired of switching between the Wazuh dashboard and Grafana to correlate events
- May be adding more security tools over time and want a stable, extensible foundation
  for a unified security view


## What It Should Look Like at the End

### Adding the datasource

The experience should feel identical to adding any other datasource in Grafana.
The user goes to Settings → Data Sources → Add, searches for "Wazuh", and sees
it in the list. They enter their Wazuh server URL and credentials, click
"Save & Test", and get a green confirmation that the connection works.

No index patterns. No manual field mapping. No knowledge of OpenSearch required.

### Pre-built dashboards

On first install, the plugin should offer a set of ready-to-use dashboards that
cover the most common Wazuh use cases:

- **Security overview** — total alerts today, alert severity distribution over time,
  top triggered rules, most active agents
- **Vulnerability detection** — vulnerable packages per agent, severity breakdown
  (critical / high / medium / low), newest vulnerabilities detected
- **File Integrity Monitoring (FIM)** — recent file changes by agent, most modified
  paths, changes by user
- **Security Configuration Assessment (SCA)** — compliance scores per agent,
  failed checks, score trends over time
- **Agent status** — which agents are active / disconnected / never connected,
  agent distribution by OS and version

These dashboards should be importable in one click, just like community dashboards
from grafana.com.

### The query editor

When building a custom panel, the user should see a Wazuh-aware query editor —
not a raw OpenSearch JSON box. It should let them:

- Select a data type (alerts, vulnerabilities, FIM events, SCA results, agent list)
- Filter by agent name, rule level, severity, time range, and other relevant fields
  using dropdowns and text fields — not hand-written queries
- Choose how to aggregate results (count over time, top N by field, latest events)

The query editor should feel like the Prometheus query editor does to someone who
knows PromQL — familiar enough to build on, but with guardrails that prevent
common mistakes.

### Correlation with infrastructure metrics

A key goal is to make it easy to combine Wazuh data with infrastructure data in
the same dashboard. For example:

- A panel showing CPU spike on a node next to a panel showing Wazuh alerts from
  that same node in the same time window
- A dashboard row per Kubernetes namespace showing both resource usage (from
  Prometheus) and policy violations or vulnerabilities (from Wazuh/Trivy)

This should work naturally because Grafana already supports mixing datasources
on a single dashboard — the plugin just needs to expose Wazuh data using the
same field names (agent name, namespace, node name) that infrastructure tools
use, so panels can be linked with Grafana variables.

### Two connection modes

Wazuh exposes data in two different ways, and the plugin should support both:

- **Wazuh Indexer (OpenSearch)** — for querying historical alert data, logs,
  vulnerability scan results, and FIM events. This is the high-volume path.
- **Wazuh REST API** — for querying live state: which agents are currently
  connected, current SCA scores, rule and decoder information. This is the
  low-volume, real-time path.

From the user's perspective, they do not need to think about which path is being
used. The query editor abstracts this — selecting "agent status" automatically
uses the API, selecting "alerts over time" automatically uses the indexer.


## What This Is Not

- It is not a replacement for the Wazuh dashboard. The Wazuh UI has features
  (active response, rule editing, agent management) that are out of scope here.
  This plugin is for *observability and correlation*, not administration.
- It is not a full SIEM replacement inside Grafana. The goal is to surface Wazuh's
  most useful security signals inside a tool teams already have open, not to
  rebuild Wazuh's entire feature set.
- It is not limited to Wazuh-only dashboards. The whole point is that Wazuh data
  sits alongside other datasources in Grafana, so mixed dashboards are the
  expected use case, not the exception.


## Success Criteria

The plugin is successful when:

1. A new user can go from zero to a working Wazuh datasource in Grafana in under
   five minutes, with no prior knowledge of OpenSearch.
2. The pre-built dashboards provide immediate value without any customization.
3. Security events from Wazuh can be shown in the same Grafana dashboard as
   Prometheus infrastructure metrics, linked by agent name or Kubernetes node.
4. The plugin works with self-hosted Wazuh deployments (the most common scenario)
   and handles authentication securely — credentials are never exposed to the browser.
5. Engineers who already use Grafana daily start checking security posture in
   Grafana rather than switching to the Wazuh UI for routine monitoring.
