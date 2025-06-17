import asyncio

from daytona import AsyncDaytona


async def main():
    async with AsyncDaytona() as daytona:
        print("Creating sandbox")
        sandbox = await daytona.create()
        print("Sandbox created")

        await sandbox.set_labels(
            {
                "public": True,
            }
        )

        print("Stopping sandbox")
        await daytona.stop(sandbox)
        print("Sandbox stopped")

        print("Starting sandbox")
        await daytona.start(sandbox)
        print("Sandbox started")

        print("Getting existing sandbox")
        existing_sandbox = await daytona.get(sandbox.id)
        print("Get existing sandbox")

        response = await existing_sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        sandboxes = await daytona.list()
        print("Total sandboxes count:", len(sandboxes))

        print(f"Printing sandboxes[0] -> id: {sandboxes[0].id} state: {sandboxes[0].state}")

        print("Removing sandbox")
        await daytona.delete(sandbox)
        print("Sandbox removed")


if __name__ == "__main__":
    asyncio.run(main())
