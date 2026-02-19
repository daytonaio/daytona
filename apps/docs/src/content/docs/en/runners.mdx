---
title: Customer Managed Compute
---

Runners are machines that power Daytona's compute plane, providing the underlying infrastructure for running sandbox workloads. Each runner is responsible for:

- **Workload execution**: running sandbox workloads
- **Resource management**: allocating and monitoring CPU, memory, and disk resources
- **Health reporting**: continuously reporting metrics and health status to the Daytona control plane
- **Network connectivity**: managing networking, proxy connections, and SSH access for sandboxes

Runners in [shared](/docs/en/regions#shared-regions) and [dedicated](/docs/en/regions#dedicated-regions) regions are fully managed by Daytona — from provisioning and maintenance to monitoring and scaling. For custom regions, you bring your own runner machines and are responsible for their management and operation.

:::caution
Custom runners are currently an experimental feature and may change in future releases.
To request access, please contact [support@daytona.io](mailto:support@daytona.io).
:::

## Custom regions

Custom regions are created and managed by your organization, allowing you to use your own runner machines and scale compute resources independently within each region. This provides maximum control over data locality, compliance, and infrastructure configuration.

Additionally, custom regions have no limits applied for concurrent resource usage, giving you full control over capacity and performance.

### Custom region configuration

**name** (required)

- A unique identifier for your region
- Must contain only letters, numbers, underscores, periods, and hyphens
- Used for targeting this region when creating a sandbox

**proxyUrl** (optional)

- The URL of the proxy service that routes traffic to sandboxes in this region
- Required if the runner machines in this region are deployed in a private network

**sshGatewayUrl** (optional)

- The URL of the SSH gateway that handles SSH connections to sandboxes in this region
- Required if the runner machines in this region are deployed in a private network

**snapshotManagerUrl** (optional)

- The URL of the snapshot manager that handles storage and retrieval of snapshots in this region
- Required if the runner machines in this region are deployed in a private network

### Custom region credentials

When you create a custom region, Daytona will provide credentials for any optional services you configure:

- An API key that should be used by your proxy service to authenticate with Daytona
- An API key that should be used by your SSH gateway service to authenticate with Daytona
- Basic authentication credentials that Daytona uses to access your snapshot manager service

:::note
If needed, these credentials can always be regenerated, but you will need to redeploy the corresponding services with the updated credentials.
:::


## Custom runners

Custom runners are created and managed by your organization, allowing you to use your own runner machines and scale compute resources independently within each custom region.

### Custom runner configuration

**name** (required)

- A unique identifier for the runner
- Must contain only letters, numbers, underscores, periods, and hyphens
- Helps distinguish between multiple runners in the same region

**regionId** (required)

- The ID of the region this runner is assigned to
- Must be a custom region owned by your organization
- All runners in a region share the region's proxy and SSH gateway configuration

### Custom runner token

When you create a custom runner, Daytona will provide you with a **token** that should be used by your runner to authenticate with Daytona.

:::note
Save this token securely. You won't be able to see it again.
:::

### Installing the custom runner

After registering a custom runner and obtaining its secure token, you need to install and configure the Daytona runner application on your infrastructure.

:::note
Detailed installation instructions for the runner application will be provided in a future update. For assistance with runner installation, please contact [support@daytona.io](mailto:support@daytona.io).
:::
