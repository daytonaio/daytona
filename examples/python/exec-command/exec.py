from daytona import (
    CreateSandboxFromImageParams,
    Daytona,
    DaytonaTimeoutError,
    ExecutionError,
    OutputMessage,
    Resources,
    Sandbox,
)


def main():
    daytona = Daytona()

    params = CreateSandboxFromImageParams(
        image="python:3.9.23-slim",
        language="python",
        resources=Resources(
            cpu=1,
            memory=1,
            disk=3,
        ),
    )
    sandbox = daytona.create(params, timeout=150, on_snapshot_create_logs=print)

    # Run the code securely inside the sandbox
    response = sandbox.process.code_run('print("Hello World!")')
    if response.exit_code != 0:
        print(f"Error: {response.exit_code} {response.result}")
    else:
        print(response.result)

    # Execute an os command in the sandbox
    response = sandbox.process.exec('echo "Hello World from exec!"', timeout=10)
    if response.exit_code != 0:
        print(f"Error: {response.exit_code} {response.result}")
    else:
        print(response.result)

    stateful_code_interpreter(sandbox)

    daytona.delete(sandbox)


def stateful_code_interpreter(sandbox: Sandbox):
    def handle_stdout(message: OutputMessage):
        print(f"[STDOUT] {message.output}")

    def handle_stderr(message: OutputMessage):
        print(f"[STDERR] {message.output}")

    def handle_error(error: ExecutionError):
        print(f"[ERROR] {error.name}: {error.value}\n{error.traceback}")

    print("\n" + "=" * 60)
    print("Stateful Code Interpreter")
    print("=" * 60)

    print("=" * 10 + " Statefulness in the default context " + "=" * 10)
    result = sandbox.code_interpreter.run_code("counter = 1\nprint(f'Initialized counter = {counter}')")
    print(f"[STDOUT] {result.stdout}")

    result = sandbox.code_interpreter.run_code(
        "counter += 1\nprint(f'Counter after second call = {counter}')",
        on_stdout=handle_stdout,
        on_stderr=handle_stderr,
        on_error=handle_error,
    )

    print("=" * 10 + " Context isolation " + "=" * 10)
    ctx = sandbox.code_interpreter.create_context()
    try:
        ctx_result = sandbox.code_interpreter.run_code(
            "value = 'stored in isolated context'\nprint(f'Isolated context value: {value}')",
            context=ctx,
            on_stdout=handle_stdout,
            on_stderr=handle_stderr,
            on_error=handle_error,
        )

        print("-" * 3 + " Print value from same context " + "-" * 3)
        ctx_result = sandbox.code_interpreter.run_code(
            "print(f'Value still available: {value}')",
            context=ctx,
        )
        print(f"[STDOUT] {ctx_result.stdout}")

        print("-" * 3 + " Print value from different context " + "-" * 3)
        _ = sandbox.code_interpreter.run_code(
            "print(value)",
            on_stdout=handle_stdout,
            on_stderr=handle_stderr,
            on_error=handle_error,
        )
    finally:
        sandbox.code_interpreter.delete_context(ctx)

    print("=" * 10 + " Timeout handling " + "=" * 10)
    try:
        code = """
import time
print('Starting long running task...')
time.sleep(5)
print('Finished!')
"""
        _ = sandbox.code_interpreter.run_code(
            code,
            timeout=1,
            on_stdout=handle_stdout,
            on_stderr=handle_stderr,
            on_error=handle_error,
        )
    except DaytonaTimeoutError as exc:
        print(f"Timed out as expected: {exc}")


if __name__ == "__main__":
    main()
