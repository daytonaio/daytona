import asyncio

from daytona import AsyncDaytona


async def main():
    async with AsyncDaytona() as daytona:
        result = await daytona.list(labels={"my-label": "my-value"}, page=2, limit=10)
        for sandbox in result.items:
            print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    asyncio.run(main())
