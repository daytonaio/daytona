from daytona_sdk import CreateSandboxParams, Daytona


def main():
    daytona = Daytona()

    params = CreateSandboxParams(
        language="python",
    )
    sandbox = daytona.create(params)

    # Run the code securely inside the sandbox
    response = sandbox.process.code_run('print("Hello World!")')
    if response.exit_code != 0:
        print(f"Error: {response.exit_code} {response.result}")
    else:
        print(response.result)

    # Execute an os command in the sandbox
    response = sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
    if response.exit_code != 0:
        print(f"Error: {response.exit_code} {response.result}")
    else:
        print(response.result)

    daytona.delete(sandbox)


if __name__ == "__main__":
    main()
