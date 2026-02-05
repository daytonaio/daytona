---
title: OpenTelemetry Collection
description: Enable distributed tracing for Daytona SDK operations using OpenTelemetry.
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

OpenTelemetry (OTEL) tracing allows you to monitor and debug your Daytona SDK operations by collecting distributed traces. This is particularly useful for understanding performance bottlenecks, debugging issues, and gaining visibility into your sandbox operations.

:::caution
OpenTelemetry collection is currently an experimental feature and may change in future releases. To request access to this feature, please contact [support@daytona.io](mailto:support@daytona.io).
:::

---

## Sandbox Telemetry Collection

Daytona can collect traces, logs, and metrics directly from your sandboxes. This provides complete observability across your entire Daytona environment, from [SDK calls](#sdk-tracing-configuration) to sandbox runtime behavior.

### Configure Sandbox Collection

To enable telemetry collection from sandboxes:

1. Navigate to the [Daytona Dashboard](https://app.daytona.io)
2. Go to **Settings** → **Experimental**
3. Configure the following fields:
   - **OTLP Endpoint**: Your OpenTelemetry collector endpoint (e.g., `https://otlp.nr-data.net`)
   - **OTLP Headers**: Authentication headers in `key=value` format (e.g., `api-key=YOUR_API_KEY`)

Once configured, all sandboxes will automatically export their telemetry data to your specified OTLP endpoint.

### What Gets Collected from Sandboxes

When sandbox telemetry is enabled, the following data is collected:

**Metrics:**

- `daytona.sandbox.cpu.utilization` - CPU usage percentage (0-100%)
- `daytona.sandbox.cpu.limit` - CPU cores limit
- `daytona.sandbox.memory.utilization` - Memory usage percentage (0-100%)
- `daytona.sandbox.memory.usage` - Memory used in bytes
- `daytona.sandbox.memory.limit` - Memory limit in bytes
- `daytona.sandbox.filesystem.utilization` - Disk usage percentage (0-100%)
- `daytona.sandbox.filesystem.usage` - Disk space used in bytes
- `daytona.sandbox.filesystem.available` - Disk space available in bytes
- `daytona.sandbox.filesystem.total` - Total disk space in bytes

**Traces:**

- HTTP requests and responses
- Custom spans from your application code

**Logs:**

- Application logs (stdout/stderr)
- System logs
- Runtime errors and warnings

### Viewing Telemetry in the Dashboard

Logs, traces, and metrics collected from sandboxes can be viewed directly in the Daytona Dashboard. Open the **Sandbox Details** sheet for any sandbox and use the **Logs**, **Traces**, and **Metrics** tabs to inspect the collected telemetry data.

:::note
Daytona retains sandbox telemetry data for **3 days**. If you need to keep the data for longer, it is recommended that you connect your own OTLP-compatible collector using the [sandbox collection configuration](#configure-sandbox-collection).
:::

:::tip
Sandbox telemetry collection works independently from SDK tracing. You can enable one or both depending on your observability needs:

- **SDK tracing only**: Monitor Daytona API operations and SDK calls
- **Sandbox telemetry only**: Monitor application behavior inside sandboxes
- **Both**: Get complete end-to-end observability across your entire stack
  :::

---

## SDK Tracing Configuration

When enabled, the Daytona SDK automatically instruments all SDK operations including:

- Sandbox creation, starting, stopping, and deletion
- File system operations
- Code execution
- Process management
- HTTP requests to the Daytona API

Traces are exported using the OTLP (OpenTelemetry Protocol) format and can be sent to any OTLP-compatible backend such as New Relic, Jaeger, or Zipkin.

### 1. Enable OTEL in SDK

To enable OpenTelemetry tracing, pass the `otelEnabled` experimental flag when initializing the Daytona client:

Alternatively, you can set the `DAYTONA_EXPERIMENTAL_OTEL_ENABLED` environment variable to `true` instead of passing the configuration option:

```bash
export DAYTONA_EXPERIMENTAL_OTEL_ENABLED=true
```

<Tabs syncKey="language">
  <TabItem label="Python" icon="seti:python">
    ```python
    from daytona import Daytona, DaytonaConfig

    # Using async context manager (recommended)
    async with Daytona(DaytonaConfig(
        _experimental={"otelEnabled": True}
    )) as daytona:
        sandbox = await daytona.create()
        # All operations will be traced
    # OpenTelemetry traces are flushed on close
    ```

    Or without context manager:
    ```python
    daytona = Daytona(DaytonaConfig(
        _experimental={"otelEnabled": True}
    ))
    try:
        sandbox = await daytona.create()
        # All operations will be traced
    finally:
        await daytona.close()  # Flushes traces
    ```

  </TabItem>

  <TabItem label="TypeScript" icon="seti:typescript">
    ```typescript
    import { Daytona } from '@daytonaio/sdk'

    // Using async dispose (recommended)
    await using daytona = new Daytona({
      _experimental: { otelEnabled: true }
    })

    const sandbox = await daytona.create()
    // All operations will be traced
    // Traces are automatically flushed on dispose
    ```

    Or with explicit disposal:
    ```typescript
    const daytona = new Daytona({
      _experimental: { otelEnabled: true }
    })

    try {
      const sandbox = await daytona.create()
      // All operations will be traced
    } finally {
      await daytona[Symbol.asyncDispose]()  // Flushes traces
    }
    ```

  </TabItem>

  <TabItem label="Go" icon="seti:go">
    ```go
    import (
        "context"
        "log"

        "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
        "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
    )

    client, err := daytona.NewClientWithConfig(&types.DaytonaConfig{
        Experimental: &types.ExperimentalConfig{
            OtelEnabled: true,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background()) // Flushes traces

    sandbox, err := client.Create(context.Background(), nil)
    // All operations will be traced
    ```

  </TabItem>

  <TabItem label="Ruby" icon="seti:ruby">
    ```ruby
    require 'daytona'

    config = Daytona::Config.new(
      _experimental: { 'otel_enabled' => true }
    )
    daytona = Daytona::Daytona.new(config)

    sandbox = daytona.create
    # All operations will be traced

    daytona.close # Flushes traces
    ```

    Or with `ensure` block:
    ```ruby
    daytona = Daytona::Daytona.new(
      Daytona::Config.new(_experimental: { 'otel_enabled' => true })
    )
    begin
      sandbox = daytona.create
      # All operations will be traced
    ensure
      daytona.close # Flushes traces
    end
    ```

  </TabItem>
</Tabs>

### 2. Configure OTLP Exporter

The SDK uses standard OpenTelemetry environment variables for configuration. Set these before running your application:

#### Required Environment Variables

```bash
# OTLP endpoint (without the /v1/traces path)
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.nr-data.net:4317

# Authentication headers (format: key1=value1,key2=value2)
OTEL_EXPORTER_OTLP_HEADERS="api-key=your-api-key-here"
```

---

## Provider-Specific Examples

### New Relic

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.nr-data.net:4317
OTEL_EXPORTER_OTLP_HEADERS="api-key=YOUR_NEW_RELIC_LICENSE_KEY"
```

### Jaeger (Local)

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

### Grafana Cloud

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-prod-<region>.grafana.net/otlp
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic <BASE64_ENCODED_CREDENTIALS>"
```

Setup: Go to [Grafana Cloud Portal](https://grafana.com) → **Connections** → **Add new connection** → Search for **OpenTelemetry (OTLP)** → Follow the wizard to create an access token. The endpoint and headers will be provided in the instrumentation instructions. See the [Grafana dashboard example](https://github.com/daytonaio/daytona/tree/main/examples/otel-dashboards/grafana) for detailed setup steps.

---

## Complete Example

Here's a complete example showing how to use OpenTelemetry tracing with the Daytona SDK:

<Tabs syncKey="language">
  <TabItem label="Python" icon="seti:python">
    ```python
    import asyncio
    import os
    from daytona import Daytona, DaytonaConfig

    # Set OTEL configuration
    os.environ["OTEL_EXPORTER_OTLP_ENDPOINT"] = "https://otlp.nr-data.net:4317"
    os.environ["OTEL_EXPORTER_OTLP_HEADERS"] = "api-key=YOUR_API_KEY"

    async def main():
        # Initialize Daytona with OTEL enabled
        async with Daytona(DaytonaConfig(
            _experimental={"otelEnabled": True}
        )) as daytona:

            # Create a sandbox - this operation will be traced
            sandbox = await daytona.create()
            print(f"Created sandbox: {sandbox.id}")

            # Execute code - this operation will be traced
            result = await sandbox.process.code_run(""

import numpy as np
print(f"NumPy version: {np.__version__}")
            "")
            print(f"Execution result: {result.result}")

            # Upload a file - this operation will be traced
            await sandbox.fs.upload_file("local.txt", "/home/daytona/remote.txt")

            # Delete sandbox - this operation will be traced
            await daytona.delete(sandbox)

        # Traces are automatically flushed when exiting the context manager

    if __name__ == "__main__":
        asyncio.run(main())
    ```

  </TabItem>

  <TabItem label="TypeScript" icon="seti:typescript">
    ```typescript
    // Set OTEL configuration
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = "https://otlp.nr-data.net:4317"
    process.env.OTEL_EXPORTER_OTLP_HEADERS = "api-key=YOUR_API_KEY"

    import { Daytona } from '@daytonaio/sdk'

    async function main() {
      // Initialize Daytona with OTEL enabled
      await using daytona = new Daytona({
        _experimental: { otelEnabled: true }
      })

      // Create a sandbox - this operation will be traced
      const sandbox = await daytona.create()
      console.log(`Created sandbox: ${sandbox.id}`)

      // Execute code - this operation will be traced
      const result = await sandbox.process.codeRun(`

import numpy as np
print(f"NumPy version: {np.__version__}")
      `)
      console.log(`Execution result: ${result.result}`)

      // Upload a file - this operation will be traced
      await sandbox.fs.uploadFile('local.txt', '/home/daytona/remote.txt')

      // Delete sandbox - this operation will be traced
      await daytona.delete(sandbox)

      // Traces are automatically flushed when the daytona instance is disposed
    }

    main().catch(console.error)
    ```

  </TabItem>

  <TabItem label="Go" icon="seti:go">
    ```go
    package main

    import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
        "github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
    )

    func main() {
        // Set OTEL configuration
        os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://otlp.nr-data.net:4317")
        os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "api-key=YOUR_API_KEY")

        ctx := context.Background()

        // Initialize Daytona with OTEL enabled
        client, err := daytona.NewClientWithConfig(&types.DaytonaConfig{
          Experimental: &types.ExperimentalConfig{
            OtelEnabled: true,
          },
        })
        if err != nil {
          log.Fatal(err)
        }
        defer client.Close(ctx) // Flushes traces on exit

        // Create a sandbox - this operation will be traced
        sandbox, err := client.Create(ctx, nil)
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Created sandbox: %s\n", sandbox.ID)

        // Execute code - this operation will be traced
        result, err := sandbox.Process.CodeRun(ctx, &types.CodeRunParams{
            Code: `

import numpy as np
print(f"NumPy version: {np.__version__}")
            `,
        })
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Execution result: %s\n", result.Result)

        // Upload a file - this operation will be traced
        err = sandbox.Fs.UploadFile(ctx, "local.txt", "/home/daytona/remote.txt")
        if err != nil {
          log.Fatal(err)
        }

        // Delete sandbox - this operation will be traced
        err = client.Delete(ctx, sandbox, nil)
        if err != nil {
          log.Fatal(err)
        }

        // Traces are flushed when client.Close is called via defer
    }
    ```

  </TabItem>

  <TabItem label="Ruby" icon="seti:ruby">
    ```ruby
    require 'daytona'

    # Set OTEL configuration
    ENV["OTEL_EXPORTER_OTLP_ENDPOINT"] = "https://otlp.nr-data.net:4317"
    ENV["OTEL_EXPORTER_OTLP_HEADERS"] = "api-key=YOUR_API_KEY"

    # Initialize Daytona with OTEL enabled
    config = Daytona::Config.new(
      _experimental: { 'otel_enabled' => true }
    )
    daytona = Daytona::Daytona.new(config)

    begin
      # Create a sandbox - this operation will be traced
      sandbox = daytona.create
      puts "Created sandbox: #{sandbox.id}"

      # Execute code - this operation will be traced
      result = sandbox.process.code_run("

import numpy as np
print(f'NumPy version: {np.__version__}')
      ")
      puts "Execution result: #{result.result}"

      # Upload a file - this operation will be traced
      sandbox.fs.upload_file("local.txt", "/home/daytona/remote.txt")

      # Delete sandbox - this operation will be traced
      daytona.delete(sandbox)
    ensure
      daytona.close # Flushes traces
    end
    ```

  </TabItem>
</Tabs>

---

## What Gets Traced

The Daytona SDK automatically instruments the following operations:

### SDK-Level Operations

- `create()` - Sandbox creation and initialization
- `get()` - Retrieving sandbox instances
- `findOne()` - Finding sandboxes by filters
- `list()` - Listing sandboxes
- `start()` - Starting sandboxes
- `stop()` - Stopping sandboxes
- `delete()` - Deleting sandboxes
- All sandbox, snapshot and volume operations (file system, code execution, process management, etc.)

### HTTP Requests

- All API calls to the Daytona backend
- Request duration and response status codes
- Error information for failed requests

### Trace Attributes

Each trace includes valuable metadata such as:

- Service name and version
- HTTP method, URL, and status code
- Request and response duration
- Error details (if applicable)
- Custom SDK operation metadata

---

## Dashboard Examples

- [New Relic](https://github.com/daytonaio/daytona/tree/main/examples/otel-dashboards/new-relic)
- [Grafana](https://github.com/daytonaio/daytona/tree/main/examples/otel-dashboards/grafana)

## Troubleshooting

### Verify Traces Are Being Sent

1. Check that environment variables are set correctly
2. Verify your OTLP endpoint is reachable
3. Confirm API keys/headers are valid
4. Check your observability platform for incoming traces
5. Look for connection errors in application logs

### Common Issues

**Traces not appearing:**

- Ensure `otelEnabled: true` is set in the configuration
- Verify OTLP endpoint and headers are correct
- Check that you're properly closing/disposing the Daytona instance to flush traces

**Connection refused:**

- Verify the OTLP endpoint URL is correct
- Ensure the endpoint is accessible from your application
- Check firewall rules if running in a restricted environment

**Authentication errors:**

- Verify API key format matches your provider's requirements
- Check that the `OTEL_EXPORTER_OTLP_HEADERS` format is correct (key=value pairs)

---

## Best Practices

1. **Always close the client**: Use `async with` (Python), `await using` (TypeScript), `defer client.Close()` (Go), or `ensure daytona.close` (Ruby) to ensure traces are properly flushed
1. **Monitor trace volume**: Be aware that enabling tracing will increase network traffic and storage in your observability platform
1. **Use in development first**: Test OTEL configuration in development before enabling in production
1. **Configure sampling**: For high-volume applications, consider configuring trace sampling to reduce costs

---

## Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
- [Daytona SDK Documentation](/docs/en/introduction/)
