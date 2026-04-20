from daytona import Daytona, ListSandboxesQuery


def main():
    daytona = Daytona()

    for sandbox in daytona.list(
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
    main()
