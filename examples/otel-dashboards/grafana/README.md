# Grafana Dashboard for Daytona Sandbox Monitoring

This directory contains a pre-configured Grafana dashboard for monitoring Daytona Sandbox resources including CPU, Memory, and Disk utilization using Prometheus metrics.

## Dashboard Overview

The dashboard provides comprehensive monitoring across multiple pages:

- **Resource Overview**: High-level view of all sandboxes with aggregate metrics
- **CPU Details**: Detailed CPU utilization, limits, and heatmaps
- **Memory Details**: Memory usage patterns and limits
- **Disk Details**: Filesystem usage and space breakdown

## Prerequisites

- Grafana Cloud account (free tier available)
- Daytona account with access to Experimental settings

## Setup

### Step 1: Create a Grafana Cloud Account

1. Go to [grafana.com](https://grafana.com) and click **Create free account**
2. Sign up with email, Google, or GitHub
3. Create a new stack (choose a region close to you)

### Step 2: Set Up OpenTelemetry Connection

1. In Grafana Cloud Portal, go to **Connections** → **Add new connection**
2. Search for **OpenTelemetry (OTLP)** and select it
3. Follow the setup wizard:
   - **Choose instrumentation method**: Select **OpenTelemetry SDK**, then your language
   - **Choose your infrastructure**: Select **Linux**
4. **Create a Grafana Cloud Access token**:
   - Click **Create a Grafana Cloud Access token for your application**
   - Name it something like `daytona-otel-token`
   - Select **All scopes**
   - Click **Create** and **save the token**
5. **Get your configuration values** from the instrumentation instructions:
   - Note the `OTEL_EXPORTER_OTLP_ENDPOINT` value (e.g., `https://otlp-gateway-prod-eu-central-0.grafana.net/otlp`)
   - Note the `OTEL_EXPORTER_OTLP_HEADERS` value (e.g., `Authorization=Basic MTUxNzAz...`)

### Step 3: Configure Daytona

1. Go to the [Daytona Dashboard](https://app.daytona.io)
2. Navigate to **Settings** → **Experimental**
3. Enter the values from Step 2:
   - **OTLP Endpoint**: The endpoint URL from Grafana (e.g., `https://otlp-gateway-prod-eu-central-0.grafana.net/otlp`)
   - **OTLP Headers**: The Authorization header from Grafana (e.g., `Authorization=Basic MTUxNzAz...`)
4. Click **Save**

### Step 4: Verify Metrics Are Flowing

1. Create a sandbox in Daytona and let it run for a few minutes
2. In Grafana Cloud, go to **Observability** → **Application** to see your sandboxes
3. Or go to **Explore**, select your Prometheus data source, and run:

   ```promql
   {__name__=~"daytona_sandbox.*"}
   ```

4. You should see metrics appearing for each sandbox

### Step 5: Import the Dashboard

1. In Grafana Cloud, click **Dashboards** in the left menu
2. Click **New** → **Import**
3. Click **Upload dashboard JSON file** and select `dashboard.json`
4. Select your Prometheus data source from the dropdown (e.g., `grafanacloud-<stack>-prom`)
5. Click **Import**

## Dashboard Variables

The dashboard uses template variables for flexible filtering:

| Variable | Description | Default |
|----------|-------------|---------|
| `$datasource` | Prometheus data source selector | Auto-detected |
| `$service` | Filter by `service_name` label (multi-select) | All services |
| `$interval` | Time aggregation interval (for custom panels) | 1m |

### Interval Options

The `$interval` variable is available for custom panels you may add:

- **1m**: Fine-grained, best for real-time monitoring
- **5m**: Balanced detail and performance
- **10m**: Good for hourly analysis
- **30m**: Overview of trends
- **1h**: Long-term trend analysis

## Widget Descriptions

### Resource Overview Page

| Widget | Type | Description |
|--------|------|-------------|
| Sandbox Count | Stat | Total number of active sandboxes reporting metrics |
| Critical Services | Stat | Count of services exceeding resource thresholds (with color coding) |
| Services Resource Overview | Table | Detailed metrics per service (CPU%, Memory%, Disk%, limits) |
| CPU Utilization by Service | Time Series | CPU usage percentage over time per service |
| Memory Utilization by Service | Time Series | Memory usage percentage over time per service |
| Disk Utilization by Service | Time Series | Disk usage percentage over time per service |
| Top CPU Consumers | Bar Gauge | Services with highest average CPU usage |
| Top Memory Consumers | Bar Gauge | Services with highest average memory usage |
| Top Disk Consumers | Bar Gauge | Services with highest average disk usage |
| Resource Pressure Score | Time Series | Combined weighted score of all resource utilization |

### CPU Details Page

| Widget | Type | Description |
|--------|------|-------------|
| CPU Utilization Timeseries | Time Series | Detailed CPU usage over time per service |
| Current CPU by Service | Stat | Current CPU % with threshold coloring |
| CPU Limit by Service | Table | CPU cores limit, average, and peak usage |
| CPU Usage Heatmap | Heatmap | Distribution of CPU usage values over time |

### Memory Details Page

| Widget | Type | Description |
|--------|------|-------------|
| Memory Utilization Timeseries | Time Series | Memory usage percentage over time |
| Current Memory by Service | Stat | Current memory % with threshold coloring |
| Memory Usage in GB | Time Series (Area) | Absolute memory usage in gigabytes |
| Memory Limits and Usage | Table | Memory used, limit, average, and peak % |

### Disk Details Page

| Widget | Type | Description |
|--------|------|-------------|
| Disk Utilization Timeseries | Time Series | Disk usage percentage over time |
| Current Disk by Service | Stat | Current disk % with threshold coloring |
| Disk Usage in GB | Time Series (Area) | Absolute disk usage in gigabytes |
| Disk Space Breakdown | Table | Used, available, total space, and utilization % |

## Alert Thresholds

The dashboard includes pre-configured color thresholds for visual alerting:

| Resource | Warning (Yellow) | Critical (Red) |
|----------|-----------------|----------------|
| CPU | 70% | 85% |
| Memory | 80% | 90% |
| Disk | 75% | 85% |

These thresholds are configured in stat panels and provide immediate visual feedback when resources are constrained.

## Metrics Reference

All metrics follow the OTEL to Prometheus naming convention (dots become underscores, units are appended as suffixes):

| OTEL Metric | Prometheus Metric | Description | Unit |
|-------------|-------------------|-------------|------|
| `daytona.sandbox.cpu.utilization` | `daytona_sandbox_cpu_utilization_percent` | CPU usage percentage | % (0-100) |
| `daytona.sandbox.cpu.limit` | `daytona_sandbox_cpu_limit_cores` | CPU cores limit | cores |
| `daytona.sandbox.memory.utilization` | `daytona_sandbox_memory_utilization_percent` | Memory usage percentage | % (0-100) |
| `daytona.sandbox.memory.usage` | `daytona_sandbox_memory_usage_bytes` | Memory used | bytes |
| `daytona.sandbox.memory.limit` | `daytona_sandbox_memory_limit_bytes` | Memory limit | bytes |
| `daytona.sandbox.filesystem.utilization` | `daytona_sandbox_filesystem_utilization_percent` | Disk usage percentage | % (0-100) |
| `daytona.sandbox.filesystem.usage` | `daytona_sandbox_filesystem_usage_bytes` | Disk space used | bytes |
| `daytona.sandbox.filesystem.available` | `daytona_sandbox_filesystem_available_bytes` | Available disk space | bytes |
| `daytona.sandbox.filesystem.total` | `daytona_sandbox_filesystem_total_bytes` | Total disk space | bytes |

### Labels

All metrics include the `service_name` label identifying the sandbox.

## Troubleshooting

### No Data Showing

1. **Verify metrics are being received**: Run this PromQL query in Grafana Explore:

   ```promql
   daytona_sandbox_cpu_utilization_percent
   ```

2. **Check data source connection**: Go to **Connections** → **Data Sources** → your Prometheus source → **Test**
3. **Verify time range**: Ensure the dashboard time picker includes when metrics were sent
4. **Check service filter**: Try selecting "All" for the `$service` variable

### High Cardinality Warnings

If you have many sandboxes, consider:

- Reducing the time range
- Using larger aggregation intervals
- Filtering to specific services

### Panel Shows "No Data"

- Verify the metric exists in Grafana Explore using `{__name__=~"daytona.*"}`
- Check label names match: `service_name` (not `service.name`)
- Ensure sandboxes are running and generating metrics

### Dashboard Import Fails

1. Ensure JSON is valid: `jq . dashboard.json`
2. Check that you have dashboard creation permissions in Grafana Cloud

## Customization

### Modifying Panels

1. Import the dashboard to Grafana
2. Enter edit mode (click pencil icon or press `e`)
3. Modify panels as needed
4. Save the dashboard

### Adding New Panels

1. Click **Add** → **Visualization**
2. Select your Prometheus data source
3. Write PromQL queries using the metrics listed above
4. Example for custom metric:

   ```promql
   avg(daytona_sandbox_cpu_utilization_percent{service_name=~"$service"}) by (service_name)
   ```

### Adjusting Thresholds

1. Edit the desired stat panel
2. Go to **Field** → **Thresholds**
3. Modify warning and critical values
4. Save the panel

### Exporting Customized Dashboard

1. Click the share icon in the top navigation bar
2. Select **Export**
3. Enable **Export for sharing externally**
4. Click **Save to file**
5. Replace `dashboard.json` with your customized version

## Additional Resources

- [Grafana Cloud Documentation](https://grafana.com/docs/grafana-cloud/)
- [Grafana Cloud OTLP Documentation](https://grafana.com/docs/grafana-cloud/send-data/otlp/)
- [PromQL Query Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
