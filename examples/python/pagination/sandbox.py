from daytona import Daytona


def main():
    daytona = Daytona()

    result = daytona.list({"my-label": "my-value"}, page=2, limit=10)
    for sandbox in result.items:
        print(f"{sandbox.id}: {sandbox.state}")


if __name__ == "__main__":
    main()
