import asyncio

from daytona import AsyncDaytona, ListSandboxesParams, SandboxState


async def main():
    daytona = AsyncDaytona()

    limit = 2
    states = [SandboxState.STARTED, SandboxState.STOPPED]

    page1 = await daytona.list(ListSandboxesParams(limit=limit, states=states))
    for sandbox in page1.items:
        print(f"{sandbox.id}: {sandbox.state}")

    if page1.next_cursor:
        page2 = await daytona.list(ListSandboxesParams(cursor=page1.next_cursor, limit=limit, states=states))
        for sandbox in page2.items:
            print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    asyncio.run(main())
