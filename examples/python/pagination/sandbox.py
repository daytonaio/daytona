from daytona import Daytona, ListSandboxesQuery, SandboxListSortDirection, SandboxListSortField, SandboxState


def main():
    daytona = Daytona()

    for sandbox in daytona.list(
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
    main()
