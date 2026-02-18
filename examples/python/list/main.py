from daytona import Daytona, ListSandboxesParams, SandboxState


def main():
    daytona = Daytona()

    limit = 2
    states = [SandboxState.STARTED, SandboxState.STOPPED]

    page1 = daytona.list(ListSandboxesParams(limit=limit, states=states))
    for sandbox in page1.items:
        print(f"{sandbox.id}: {sandbox.state}")

    if page1.next_cursor:
        page2 = daytona.list(ListSandboxesParams(cursor=page1.next_cursor, limit=limit, states=states))
        for sandbox in page2.items:
            print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    main()
