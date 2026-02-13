import asyncio

from daytona import AsyncDaytona, SandboxState


async def main():
    daytona = AsyncDaytona()

    states_filter = [SandboxState.STARTED, SandboxState.STOPPED]

    page1 = await daytona.list_v2(limit=2, states=states_filter)
    for sandbox in page1.items:
        print(f"{sandbox.id}: {sandbox.state}")

    if page1.next_cursor:
        page2 = await daytona.list_v2(cursor=page1.next_cursor, limit=2, states=states_filter)
        for sandbox in page2.items:
            print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    asyncio.run(main())
