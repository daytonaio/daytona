import asyncio

from daytona_sdk import AsyncDaytona


async def main():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        try:
            project_dir = "learn-typescript"

            # Clone the repository
            await sandbox.git.clone(
                "https://github.com/panaverse/learn-typescript",
                project_dir,
                "master",
            )

            await sandbox.git.pull(project_dir)

            # Search for the file we want to work on
            matches = await sandbox.fs.find_files(project_dir, "var obj1 = new Base();")
            print("Matches:", matches)

            # Start the language server
            lsp = sandbox.create_lsp_server("typescript", project_dir)
            await lsp.start()

            # Notify the language server of the document we want to work on
            await lsp.did_open(matches[0].file)

            # Get symbols in the document
            symbols = await lsp.document_symbols(matches[0].file)
            print("Symbols:", symbols)

            # Fix the error in the document
            await sandbox.fs.replace_in_files([matches[0].file], "var obj1 = new Base();", "var obj1 = new E();")

            # Notify the language server of the document change
            await lsp.did_close(matches[0].file)
            await lsp.did_open(matches[0].file)

            # Get completions at a specific position
            completions = await lsp.completions(matches[0].file, {"line": 12, "character": 18})
            print("Completions:", completions)

        except Exception as error:
            print("Error creating sandbox:", error)
        finally:
            # Cleanup
            await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
