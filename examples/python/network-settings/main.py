from daytona import CreateSandboxFromSnapshotParams, Daytona


def main():
    daytona = Daytona()

    # Default settings
    sandbox1 = daytona.create()
    print("network_block_all:", sandbox1.network_block_all)
    print("network_allow_list:", sandbox1.network_allow_list)

    # Block all network access
    sandbox2 = daytona.create(params=CreateSandboxFromSnapshotParams(network_block_all=True))
    print("network_block_all:", sandbox2.network_block_all)
    print("network_allow_list:", sandbox2.network_allow_list)

    # Explicitly allow list of network addresses
    sandbox3 = daytona.create(params=CreateSandboxFromSnapshotParams(network_allow_list="192.168.1.0/16,10.0.0.0/24"))
    print("network_block_all:", sandbox3.network_block_all)
    print("network_allow_list:", sandbox3.network_allow_list)

    sandbox1.delete()
    sandbox2.delete()
    sandbox3.delete()


if __name__ == "__main__":
    main()
