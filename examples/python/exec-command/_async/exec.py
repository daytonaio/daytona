import asyncio

from daytona import AsyncDaytona, CreateSandboxFromImageParams, Resources


async def main():
    async with AsyncDaytona() as daytona:
        params = CreateSandboxFromImageParams(
            image="python:3.9.23-slim",
            language="python",
            resources=Resources(
                cpu=1,
                memory=1,
                disk=3,
            ),
        )
        sandbox = await daytona.create(params, timeout=150, on_snapshot_create_logs=print)

        # Run the code securely inside the sandbox
        response = await sandbox.process.code_run('print("Hello World!")')
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        # Execute an os command in the sandbox
        response = await sandbox.process.exec('echo "Hello World from exec!"', timeout=10)
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
