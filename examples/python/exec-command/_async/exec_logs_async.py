import asyncio

from daytona import AsyncDaytona, SessionExecuteRequest


async def main():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        try:
            session_id = "exec-session-1"
            await sandbox.process.create_session(session_id)

            command = await sandbox.process.execute_session_command(
                session_id,
                SessionExecuteRequest(
                    command=(
                        "counter=1; "
                        "while (( counter <= 3 )); do "
                        'echo "Count: $counter"; '
                        "((counter++)); "
                        "sleep 2; "
                        "done"
                    ),
                    run_async=True,
                ),
            )

            logs_task = asyncio.create_task(
                sandbox.process.get_session_command_logs_async(
                    session_id, command.cmd_id, lambda chunk: print(f"Log chunk: {chunk}")
                )
            )

            print("Continuing execution while logs are streaming...")
            await asyncio.sleep(1)
            print("Other operations completed!")

            print("Now waiting for logs to complete...")
            await logs_task
        except Exception as e:
            print(f"Error: {e}")
        finally:
            print("Cleaning up sandbox...")
            await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
