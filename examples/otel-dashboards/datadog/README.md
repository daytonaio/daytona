# Datadog Dashboard for Daytona Sandbox Monitoring

This directory contains a pre-configured Datadog dashboard for monitoring Daytona Sandbox resources (CPU, Memory, Disk) and organization-level quota usage. Metrics are ingested into Datadog via OpenTelemetry (`daytona.sandbox.*`).

## Dashboard Overview

The dashboard is organized into collapsible groups (one per topic):

- **Resource Overview**: High-level view of all sandboxes with aggregate metrics
- **CPU Details**: Detailed CPU utilization, limits, and a usage heatmap
- **Memory Details**: Memory usage patterns and limits
- **Disk Details**: Filesystem usage and space breakdown
- **Organization Quotas**: Organization-level resource usage vs. quota, by region

## Prerequisites

- A Datadog account with permission to create dashboards
- A Datadog API key (for sending telemetry — see "Sending Data to Datadog" below)
- Daytona telemetry flowing into Datadog (see [OpenTelemetry Collection docs](https://www.daytona.io/docs/en/observability/otel-collection/))

## Sending Data to Datadog

Datadog exposes a native OTLP intake endpoint, so you can point Daytona's OTLP destination directly at it — no Datadog Agent required.

1. In the [Daytona Dashboard](https://app.daytona.io), go to **Settings** → **OpenTelemetry** (organization owners only).
2. Configure the destination:
   - **OTLP Endpoint**: your Datadog site's OTLP intake URL
     - US1: `https://otlp.datadoghq.com`
     - EU: `https://otlp.datadoghq.eu`
     - US3: `https://otlp.us3.datadoghq.com`
     - US5: `https://otlp.us5.datadoghq.com`
     - AP1: `https://otlp.ap1.datadoghq.com`
   - **Headers**: add `dd-api-key` = `<YOUR_DATADOG_API_KEY>`

> **Note**
> Datadog's OTLP **metrics** intake requires **delta** temporality for cumulative metric types (counters, histograms). Daytona's sandbox and organization metrics are all **gauges**, so they are accepted as-is. If you also export your own cumulative application metrics from inside sandboxes, configure your SDK to use delta temporality (`OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta`). OTLP **traces** intake is in Preview at Datadog and may require enabling for your account.

### How OTel attributes map to Datadog tags

- `service.name` → the reserved **`service`** tag (used to group/filter sandboxes)
- `region.id` → the **`region.id`** tag (used on the Organization Quotas page)
- `organization.id` → the **`organization.id`** tag

If a tag doesn't resolve in the dashboard, open **Metrics → Summary**, search for `daytona.sandbox.cpu.utilization`, and confirm the exact tag keys Datadog assigned — then adjust the template variables/queries to match.

## Importing the Dashboard

### Via the Web UI

1. Go to **Dashboards** → **New Dashboard**.
2. Give it any name, then open the dashboard's settings (gear icon, top right) → **Import dashboard JSON**.
   - Alternatively, on the **Dashboard List** page, use the **New Dashboard** dropdown → **Import Dashboard JSON file** and upload `dashboard.json`.
3. Paste the contents of `dashboard.json` (or upload the file) and confirm the import.
4. Save. The dashboard will be named **"Daytona Sandbox Resource Monitoring"**.

### Via the API

```bash
curl -X POST "https://api.datadoghq.com/api/v1/dashboard" \
  -H "Content-Type: application/json" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}" \
  -d @dashboard.json
```

Use the API host matching your Datadog site (e.g. `api.datadoghq.eu`, `api.us3.datadoghq.com`).

## Template Variables

The dashboard ships with two template variables for filtering:

- **`$service`** — filter by the `service` tag (individual sandbox). Default: `*` (all).
- **`$region`** — filter by the `region.id` tag (used on the Organization Quotas page). Default: `*` (all).

## Alert Thresholds

The overview tiles and per-resource lists are color-coded with these thresholds:

- **CPU**: Warning at 70%, Critical at 85%
- **Memory**: Warning at 80%, Critical at 90%
- **Disk**: Warning at 75%, Critical at 85%

## Metrics Tracked

Per-sandbox metrics (prefixed with `daytona.sandbox.`):

| Metric | Unit | Description |
| --- | --- | --- |
| `cpu.utilization` | percent | CPU usage percentage (0–100) |
| `cpu.limit` | cores | CPU cores limit |
| `memory.utilization` | percent | Memory usage percentage (0–100) |
| `memory.usage` | bytes | Memory used |
| `memory.limit` | bytes | Memory limit |
| `filesystem.utilization` | percent | Disk usage percentage (0–100) |
| `filesystem.usage` | bytes | Disk space used |
| `filesystem.available` | bytes | Disk space available |
| `filesystem.total` | bytes | Total disk space |

Organization-level quota metrics (also prefixed with `daytona.sandbox.`, tagged by `region.id`):

| Metric | Unit | Description |
| --- | --- | --- |
| `used_cpu` / `total_cpu` | cores | CPU consumed / quota |
| `used_ram` / `total_ram` | GiB | Memory consumed / quota |
| `used_storage` / `total_storage` | GiB | Storage consumed / quota |

> Memory/disk byte metrics are divided by `1073741824` in the dashboard to display GB.

## Troubleshooting

### No data in widgets

1. Confirm telemetry is reaching Datadog: **Metrics → Summary**, search `daytona.sandbox`.
2. Check the OTLP endpoint matches your Datadog **site** and the `dd-api-key` header is set.
3. Verify the `service` / `region.id` tags exist on the metrics (tag keys can differ depending on your collector's attribute mapping). Adjust template variables/queries if needed.
4. Widen the dashboard time range — sandbox metrics are emitted periodically.

### Import fails

1. Validate the JSON: `jq . dashboard.json`.
2. Ensure you're pasting into **Import dashboard JSON** (not the raw widget editor).

## Customization

To modify the dashboard:

1. Import it into Datadog.
2. Edit widgets through the UI.
3. Export via the dashboard settings (gear) → **Export dashboard JSON**.
4. Replace `dashboard.json` with your customized version.

## Additional Resources

- [Datadog OpenTelemetry Documentation](https://docs.datadoghq.com/opentelemetry/)
- [Datadog OTLP Ingestion](https://docs.datadoghq.com/opentelemetry/setup/otlp_ingest/)
- [Datadog Dashboards Documentation](https://docs.datadoghq.com/dashboards/)
- [Daytona OpenTelemetry Collection](https://www.daytona.io/docs/en/observability/otel-collection/)
