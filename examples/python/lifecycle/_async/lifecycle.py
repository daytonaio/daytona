import asyncio

from daytona import AsyncDaytona, Resources


async def main():
    async with AsyncDaytona() as daytona:
        print("Creating sandbox")
        sandbox = await daytona.create()
        print("Sandbox created")

        _ = await sandbox.set_labels(
            {
                "public": "true",
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

        result = await daytona.list()
        print("Total sandboxes count:", result.total)

        print(f"Printing first sandbox -> id: {result.items[0].id} state: {result.items[0].state}")

        # Hot resize: increase CPU and memory on a running sandbox
        print("Resizing sandbox (hot resize)...")
        await sandbox.resize(Resources(cpu=2, memory=2), hot=True)
        print(f"Hot resize complete: CPU={sandbox.cpu}, Memory={sandbox.memory}GB, Disk={sandbox.disk}GB")

        # Cold resize: stop sandbox first, then resize (can also change disk)
        print("Stopping sandbox for cold resize...")
        await daytona.stop(sandbox)
        print("Resizing sandbox (cold resize)...")
        await sandbox.resize(Resources(cpu=4, memory=4, disk=20), hot=False)
        print(f"Cold resize complete: CPU={sandbox.cpu}, Memory={sandbox.memory}GB, Disk={sandbox.disk}GB")
        await daytona.start(sandbox)
        print("Sandbox restarted with new resources")

        print("Removing sandbox")
        await daytona.delete(sandbox)
        print("Sandbox removed")


if __name__ == "__main__":
    asyncio.run(main())
