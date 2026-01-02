---
title: Python SDK Reference
description: Interact with Daytona Sandboxes using the Python SDK
next: /docs/python-sdk/daytona
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

The Daytona Python SDK provides a robust interface for programmatically interacting with Daytona Sandboxes.

## Installation

Install the Daytona Python SDK using pip:

```bash
pip install daytona
```

Or using poetry:

```bash
poetry add daytona
```

## Getting Started

### Create a Sandbox

Create a Daytona Sandbox to run your code securely in an isolated environment. The following snippet is an example "Hello World" program that runs securely inside a Daytona Sandbox.

<Tabs>
  <TabItem label="Sync" icon="seti:python">
    ```python
    from daytona import Daytona

    def main():
        # Initialize the SDK (uses environment variables by default)
        daytona = Daytona()

        # Create a new sandbox
        sandbox = daytona.create()

        # Execute a command
        response = sandbox.process.exec("echo 'Hello, World!'")
        print(response.result)

    if __name__ == "__main__":
        main()
    ```
  </TabItem>

  <TabItem label="Async" icon="seti:python">
    ```python
    import asyncio
    from daytona import AsyncDaytona

    async def main():
        # Initialize the SDK (uses environment variables by default)
        async with AsyncDaytona() as daytona:
            # Create a new sandbox
            sandbox = await daytona.create()

            # Execute a command
            response = await sandbox.process.exec("echo 'Hello, World!'")
            print(response.result)

    if __name__ == "__main__":
        asyncio.run(main())
    ```
  </TabItem>
</Tabs>

## Configuration

The Daytona SDK can be configured using environment variables or by passing options to the constructor:

<Tabs>
  <TabItem label="Sync" icon="seti:python">
    ```python
    from daytona import Daytona, DaytonaConfig

    # Using environment variables (DAYTONA_API_KEY, DAYTONA_API_URL, DAYTONA_TARGET)
    daytona = Daytona()

    # Using explicit configuration
    config = DaytonaConfig(
        api_key="YOUR_API_KEY",
        api_url="https://app.daytona.io/api",
        target="us"
    )
    daytona = Daytona(config)
    ```
  </TabItem>

  <TabItem label="Async" icon="seti:python">
    ```python
    import asyncio
    from daytona import AsyncDaytona, DaytonaConfig

    async def main():
        try:
            # Using environment variables (DAYTONA_API_KEY, DAYTONA_API_URL, DAYTONA_TARGET)
            daytona = AsyncDaytona()
            # Your async code here
            pass
        finally:
            await daytona.close()

        # Using explicit configuration
        config = DaytonaConfig(
            api_key="YOUR_API_KEY",
            api_url="https://app.daytona.io/api",
            target="us"
        )
        async with AsyncDaytona(config) as daytona:
            # Your code here
            pass

    if __name__ == "__main__":
        asyncio.run(main())
    ```
  </TabItem>
</Tabs>

For more information on configuring the Daytona SDK, see [configuration](/docs/en/configuration).
