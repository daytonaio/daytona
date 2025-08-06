export const daytona_python = `from daytona_sdk import Daytona, DaytonaConfig
 
# Initialize the Daytona client
daytona = Daytona(DaytonaConfig())

# Create the sandbox instance
sandbox = daytona.create()

# Run the code securely inside the sandbox
response = sandbox.process.code_run('print("Hello World!")')
print(response.result)

# Execute an os command in the sandbox
response = sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
print(response.result)

# Add a new file to the workspace
file_content = b"Hello, World!"
sandbox.fs.upload_file("/home/daytona/data.txt", file_content)

# delete the sandbox
daytona.remove(sandbox)
`
