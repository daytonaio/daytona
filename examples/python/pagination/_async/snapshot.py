import asyncio

from daytona import AsyncDaytona


async def main():
    async with AsyncDaytona() as daytona:
        result = await daytona.snapshot.list(page=2, limit=10)
        for snapshot in result.items:
            print(f"{snapshot.name} ({snapshot.image_name})")


if __name__ == "__main__":
    asyncio.run(main())
