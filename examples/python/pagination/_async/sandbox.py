import asyncio

from daytona import AsyncDaytona, ListSandboxesQuery, SandboxListSortDirection, SandboxListSortField, SandboxState


async def main():
    async with AsyncDaytona() as daytona:
        async for sandbox in daytona.list(
            ListSandboxesQuery(
                limit=10,
                labels={"env": "dev"},
                states=[SandboxState.STARTED],
                sort=SandboxListSortField.CREATEDAT,
                order=SandboxListSortDirection.DESC,
            )
        ):
            print(sandbox.id)


if __name__ == "__main__":
    asyncio.run(main())
