import asyncio
from pprint import pprint

from daytona_sdk import AsyncDaytona


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
        existing_sandbox = await daytona.get_current_sandbox(sandbox.id)
        print("Get existing sandbox")

        response = await existing_sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            print(response.result)

        sandboxes = await daytona.list()
        print("Total sandboxes count:", len(sandboxes))
        # This will show all attributes of the first sandbox
        pprint(vars(await sandboxes[0].info()))

        print("Removing sandbox")
        await daytona.delete(sandbox)
        print("Sandbox removed")


if __name__ == "__main__":
    asyncio.run(main())
