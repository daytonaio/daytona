import asyncio

from daytona_sdk import AsyncDaytona, CreateSandboxParams


async def main():
    async with AsyncDaytona() as daytona:
        params = CreateSandboxParams(
            language="python",
        )
        sandbox = await daytona.create(params)

        # Run the code securely inside the sandbox
        response = await sandbox.process.code_run('print("Hello World!")')
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        # Execute an os command in the sandbox
        response = await sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
