import time

from daytona_sdk import CreateSandboxParams, Daytona, Image


def main():
    daytona = Daytona()

    # Generate unique name for the image to avoid conflicts
    image_name = f"python-example:{int(time.time())}"

    # Create local file with some data and add it to the image
    with open("file_example.txt", "w") as f:
        f.write("Hello, World!")

    # Create a Python image with common data science packages
    image = (
        Image.debian_slim("3.12")
        .pip_install(["numpy", "pandas", "matplotlib", "scipy", "scikit-learn", "jupyter"])
        .run_commands(
            [
                "apt-get update && apt-get install -y git",
                "groupadd -r daytona && useradd -r -g daytona -m daytona",
                "mkdir -p /home/daytona/workspace",
            ]
        )
        .workdir("/home/daytona/workspace")
        .env({"MY_ENV_VAR": "My Environment Variable"})
        .add_local_file("file_example.txt", "/home/daytona/workspace/file_example.txt")
    )

    # Create the image
    print(f"=== Creating Image: {image_name} ===")
    daytona.create_image(image_name, image, on_logs=lambda chunk: print(chunk, end=""))

    # Create first sandbox using the pre-built image
    print("\n=== Creating Sandbox from Pre-built Image ===")
    sandbox1 = daytona.create(
        CreateSandboxParams(image=image_name), on_image_build_logs=lambda chunk: print(chunk, end="")
    )

    try:
        # Verify the first sandbox environment
        print("Verifying sandbox from pre-built image:")
        response = sandbox1.process.exec("python --version && pip list")
        print("Python environment:")
        print(response.result)

        # Verify the file was added to the image
        response = sandbox1.process.exec("cat workspace/file_example.txt")
        print("File content:")
        print(response.result)
    finally:
        # Clean up first sandbox
        daytona.remove(sandbox1)

    # Create second sandbox with a new dynamic image
    print("=== Creating Sandbox with Dynamic Image ===")

    # Define a new dynamic image for the second sandbox
    dynamic_image = (
        Image.debian_slim("3.11")
        .pip_install(["pytest", "pytest-cov", "black", "isort", "mypy", "ruff"])
        .run_commands(["apt-get update && apt-get install -y git", "mkdir -p /home/daytona/project"])
        .workdir("/home/daytona/project")
        .env({"ENV_VAR": "My Environment Variable"})
    )

    # Create sandbox with the dynamic image
    sandbox2 = daytona.create(
        CreateSandboxParams(
            image=dynamic_image,
        ),
        timeout=0,
        on_image_build_logs=print,
    )

    try:
        # Verify the second sandbox environment
        print("Verifying sandbox with dynamic image:")
        response = sandbox2.process.exec("pip list | grep -E 'pytest|black|isort|mypy|ruff'")
        print("Development tools:")
        print(response.result)
    finally:
        # Clean up second sandbox
        daytona.remove(sandbox2)


if __name__ == "__main__":
    main()
