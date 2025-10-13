# New Relic Dashboard for Daytona Sandbox Monitoring

This directory contains a pre-configured New Relic dashboard for monitoring Daytona Sandbox resources including CPU, Memory, and Disk utilization.

## Dashboard Overview

The dashboard provides comprehensive monitoring across multiple pages:

- **Resource Overview**: High-level view of all sandboxes with aggregate metrics
- **CPU Details**: Detailed CPU utilization, limits, and heatmaps
- **Memory Details**: Memory usage patterns and limits
- **Disk Details**: Filesystem usage and space breakdown

## Prerequisites

- New Relic account with access to create dashboards
- Your New Relic Account ID
- `jq` installed (for the automated setup script)
- New Relic CLI (`newrelic`) installed (optional, for CLI import)

## Preparing the Dashboard

Before importing, you need to add your Account ID into the `dashboard.json`.

### Quick Setup (One-Liner)

```bash
jq --arg account_id "YOUR_ACCOUNT_ID" 'walk(if type == "object" and has("accountIds") then .accountIds = [($account_id | tonumber)] else . end)' dashboard.json > dashboard-configured.json
```

Replace `YOUR_ACCOUNT_ID` with your actual numeric Account ID (e.g., `1234567`).

**Example:**

```bash
jq --arg account_id "1234567" 'walk(if type == "object" and has("accountIds") then .accountIds = [($account_id | tonumber)] else . end)' dashboard.json > dashboard-configured.json
```

This will create `dashboard-configured.json` ready for import.

### Using New Relic CLI (Optional)

If you have the New Relic CLI installed:

```bash
# First, prepare the dashboard with your Account ID (using the one-liner above)
jq --arg account_id "YOUR_ACCOUNT_ID" 'walk(if type == "object" and has("accountIds") then .accountIds = [($account_id | tonumber)] else . end)' dashboard.json > dashboard-configured.json

# Then import using the CLI
newrelic entity dashboard create --dashboard dashboard-configured.json
```

## Import via Web UI

1. **Find Your Account ID**:
   - Log in to [New Relic](https://one.newrelic.com)
   - Click on your account name in the bottom left
   - Your Account ID is displayed in the account dropdown

2. **Prepare the Dashboard**:

   ```bash
   jq --arg account_id "YOUR_ACCOUNT_ID" 'walk(if type == "object" and has("accountIds") then .accountIds = [($account_id | tonumber)] else . end)' dashboard.json > dashboard-configured.json
   ```

3. **Import the Dashboard**:
   - Go to [New Relic Dashboards](https://one.newrelic.com/dashboards)
   - Click **"Import dashboard"** in the top right
   - Upload `dashboard-configured.json`
   - Click **"Import dashboard"**

4. **Verify**:
   - The dashboard should now appear in your dashboards list
   - Named: "Daytona Sandbox Resource Monitoring - Multi-Service"

## Dashboard Widgets

### Resource Overview Page

- **Sandbox Count**: Total number of active sandboxes
- **Critical Services**: Count of services exceeding resource thresholds
- **Services Resource Overview**: Detailed table of all metrics per service
- **CPU/Memory/Disk Utilization**: Time-series graphs per service
- **Top Consumers**: Bar charts showing highest resource usage
- **Resource Pressure Score**: Combined metric showing overall resource strain

### Detailed Pages

Each resource type (CPU, Memory, Disk) has a dedicated page with:

- Time-series utilization graphs
- Current values with threshold alerts
- Usage in absolute units (cores, GB)
- Limits and capacity information
- Historical averages and peak values

## Alert Thresholds

The dashboard includes pre-configured alert severities:

- **CPU**: Warning at 70%, Critical at 85%
- **Memory**: Warning at 80%, Critical at 90%
- **Disk**: Warning at 75%, Critical at 85%

## Metrics Tracked

All metrics are prefixed with `daytona.sandbox.`:

- `cpu.utilization`: CPU usage percentage
- `cpu.limit`: CPU cores limit
- `memory.utilization`: Memory usage percentage
- `memory.usage`: Memory used in bytes
- `memory.limit`: Memory limit in bytes
- `filesystem.utilization`: Disk usage percentage
- `filesystem.usage`: Disk space used in bytes
- `filesystem.available`: Available disk space in bytes
- `filesystem.total`: Total disk space in bytes

## Troubleshooting

### Dashboard Import Fails

1. Ensure JSON is valid: `jq . dashboard-configured.json`
2. Verify Account ID is numeric (not a string)
3. Check that you have dashboard creation permissions

### jq Not Installed

If the command fails:

- Install `jq`: `brew install jq` (macOS) or `apt-get install jq` (Linux)
- Verify installation: `jq --version`

## Customization

To modify the dashboard:

1. Import it to New Relic
2. Edit widgets through the UI
3. Export the modified dashboard
4. Replace `dashboard.json` with your customized version

## Additional Resources

- [New Relic Dashboard Documentation](https://docs.newrelic.com/docs/query-your-data/explore-query-data/dashboards/introduction-dashboards/)
- [NRQL Query Language](https://docs.newrelic.com/docs/query-your-data/nrql-new-relic-query-language/get-started/introduction-nrql-new-relics-query-language/)
- [New Relic CLI](https://github.com/newrelic/newrelic-cli)
