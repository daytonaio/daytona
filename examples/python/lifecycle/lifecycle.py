from pprint import pprint

from daytona_sdk import Daytona

daytona = Daytona()

print("Creating sandbox")
sandbox = daytona.create()
print("Sandbox created")

sandbox.set_labels(
    {
        "public": True,
    }
)

print("Stopping sandbox")
daytona.stop(sandbox)
print("Sandbox stopped")

print("Starting sandbox")
daytona.start(sandbox)
print("Sandbox started")

print("Getting existing sandbox")
existing_sandbox = daytona.get_current_sandbox(sandbox.id)
print("Get existing sandbox")

response = existing_sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
if response.exit_code != 0:
    print(f"Error: {response.exit_code} {response.result}")
else:
    print(response.result)

sandboxes = daytona.list()
print("Total sandboxes count:", len(sandboxes))
# This will show all attributes of the first sandbox
pprint(vars(sandboxes[0].info()))

print("Removing sandbox")
daytona.delete(sandbox)
print("Sandbox removed")
