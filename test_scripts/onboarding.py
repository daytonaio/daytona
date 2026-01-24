from daytona import Daytona, DaytonaConfig
  
# Define the configuration
config = DaytonaConfig(
    api_key="dtn_07318c7f3eff5b0227f85bffdc29929168ad8faa417deaedac886e8c1e9b8f48",
    target="http://localhost:3000/api"
)

# Initialize the Daytona client
daytona = Daytona(config)

# Create the Sandbox instance
sandbox = daytona.create()

# Run the code securely inside the Sandbox
response = sandbox.process.code_run('print("Hello World from code!")')
if response.exit_code != 0:
  print(f"Error: {response.exit_code} {response.result}")
else:
    print(response.result)
  