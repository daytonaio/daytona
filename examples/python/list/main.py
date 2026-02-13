from daytona import Daytona, SandboxState


def main():
    daytona = Daytona()

    states_filter = [SandboxState.STARTED, SandboxState.STOPPED]

    page1 = daytona.list_v2(limit=2, states=states_filter)
    for sandbox in page1.items:
        print(f"{sandbox.id}: {sandbox.state}")

    if page1.next_cursor:
        page2 = daytona.list_v2(cursor=page1.next_cursor, limit=2, states=states_filter)
        for sandbox in page2.items:
            print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    main()
