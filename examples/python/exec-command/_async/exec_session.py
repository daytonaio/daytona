import asyncio

from daytona import AsyncDaytona, SessionExecuteRequest


async def main():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        exec_session_id = "exec-session-1"
        await sandbox.process.create_session(exec_session_id)

        # Get the session details any time
        session = await sandbox.process.get_session(exec_session_id)
        print(session)

        # Execute a first command in the session
        exec_command1 = await sandbox.process.execute_session_command(
            exec_session_id, SessionExecuteRequest(command="export FOO=BAR")
        )
        if exec_command1.exit_code != 0:
            print(f"Error: {exec_command1.exit_code} {exec_command1.output}")

        # Get the session details again to see the command has been executed
        session = await sandbox.process.get_session(exec_session_id)
        print(session)

        # Get the command details
        session_command = await sandbox.process.get_session_command(exec_session_id, exec_command1.cmd_id)
        print(session_command)

        # Execute a second command in the session and see that the environment variable is set
        exec_command2 = await sandbox.process.execute_session_command(
            exec_session_id, SessionExecuteRequest(command="echo $FOO")
        )
        if exec_command2.exit_code != 0:
            print(f"Error: {exec_command2.exit_code} {exec_command2.output}")
        else:
            print(exec_command2.output)

        print("Now getting logs for the second command")
        logs = await sandbox.process.get_session_command_logs(exec_session_id, exec_command2.cmd_id)
        print(logs)

        # You can also list all active sessions
        sessions = await sandbox.process.list_sessions()
        print(sessions)

        # And of course you can delete the session at any time
        await sandbox.process.delete_session(exec_session_id)

        await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
