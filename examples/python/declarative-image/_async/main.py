import asyncio
import time

from daytona import (
    AsyncDaytona,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    CreateSnapshotParams,
    Image,
    Resources,
)


async def main():
    daytona = AsyncDaytona()

    try:
        # Generate unique name for the snapshot to avoid conflicts
        snapshot_name = f"python-example:{int(time.time())}"

        # Create local file with some data and add it to the image
        with open("file_example.txt", "w") as f:
            f.write("Hello, World!")

        # Create a Python image with common data science packages
        image = (
            Image.debian_slim("3.12")
            .pip_install(["numpy", "pandas", "matplotlib", "scipy", "scikit-learn", "jupyter"])
            .run_commands(
                "apt-get update && apt-get install -y git",
                "groupadd -r daytona && useradd -r -g daytona -m daytona",
                "mkdir -p /home/daytona/workspace",
            )
            .dockerfile_commands(["USER daytona"])
            .workdir("/home/daytona/workspace")
            .env({"MY_ENV_VAR": "My Environment Variable"})
            .add_local_file("file_example.txt", "/home/daytona/workspace/file_example.txt")
        )

        # Create the image
        print(f"=== Creating Snapshot: {snapshot_name} ===")
        await daytona.snapshot.create(
            CreateSnapshotParams(
                name=snapshot_name,
                image=image,
                resources=Resources(
                    cpu=1,
                    memory=1,
                    disk=3,
                ),
            ),
            on_logs=print,
        )

        # Create first sandbox using the pre-built image
        print("\n=== Creating Sandbox from Pre-built Image ===")
        sandbox1 = await daytona.create(CreateSandboxFromSnapshotParams(snapshot=snapshot_name))

        try:
            # Verify the first sandbox environment
            print("Verifying sandbox from pre-built image:")
            response = await sandbox1.process.exec("python --version && pip list")
            print("Python environment:")
            print(response.result)

            # Verify the file was added to the image
            response = await sandbox1.process.exec("cat workspace/file_example.txt")
            print("File content:")
            print(response.result)
        finally:
            # Clean up first sandbox
            await daytona.delete(sandbox1)

        # Create second sandbox with a new dynamic image
        print("=== Creating Sandbox with Dynamic Image ===")

        # Define a new dynamic image for the second sandbox
        dynamic_image = (
            Image.debian_slim("3.11")
            .pip_install(["pytest", "pytest-cov", "black", "isort", "mypy", "ruff"])
            .run_commands("apt-get update && apt-get install -y git", "mkdir -p /home/daytona/project")
            .workdir("/home/daytona/project")
            .env({"ENV_VAR": "My Environment Variable"})
        )

        # Create sandbox with the dynamic image
        sandbox2 = await daytona.create(
            CreateSandboxFromImageParams(
                image=dynamic_image,
            ),
            timeout=0,
            on_snapshot_create_logs=print,
        )

        try:
            # Verify the second sandbox environment
            print("Verifying sandbox with dynamic image:")
            response = await sandbox2.process.exec("pip list | grep -E 'pytest|black|isort|mypy|ruff'")
            print("Development tools:")
            print(response.result)
        finally:
            # Clean up second sandbox
            await daytona.delete(sandbox2)
    finally:
        await daytona.close()


if __name__ == "__main__":
    asyncio.run(main())
