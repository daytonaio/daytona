import asyncio

from daytona import AsyncDaytona, ListSandboxesQuery


async def main():
    async with AsyncDaytona() as daytona:
        async for sandbox in daytona.list(
            ListSandboxesQuery(
                limit=10,
                labels={"env": "dev"},
                states=["started"],
                sort="createdAt",
                order="desc",
            )
        ):
            print(sandbox.id)


if __name__ == "__main__":
    asyncio.run(main())
