import asyncio

from daytona import AsyncDaytona, CreateSandboxFromSnapshotParams


async def main():
    async with AsyncDaytona() as daytona:
        # Default settings
        sandbox1 = await daytona.create()
        print("network_block_all:", sandbox1.network_block_all)
        print("network_allow_list:", sandbox1.network_allow_list)

        # Block all network access
        sandbox2 = await daytona.create(params=CreateSandboxFromSnapshotParams(network_block_all=True))
        print("network_block_all:", sandbox2.network_block_all)
        print("network_allow_list:", sandbox2.network_allow_list)

        # Explicitly allow list of network addresses
        sandbox3 = await daytona.create(
            params=CreateSandboxFromSnapshotParams(network_allow_list="192.168.1.0/16,10.0.0.0/24")
        )
        print("network_block_all:", sandbox3.network_block_all)
        print("network_allow_list:", sandbox3.network_allow_list)

        await sandbox1.delete()
        await sandbox2.delete()
        await sandbox3.delete()


if __name__ == "__main__":
    asyncio.run(main())
