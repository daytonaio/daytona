import subprocess

# Simple query to get all registry info
query = "SELECT id, url, username, LENGTH(password) as pwd_len FROM docker_registry;"

result = subprocess.run(
    ['docker', 'exec', 'daytona-db-1', 'psql', '-U', 'user', '-d', 'daytona', '-c', query],
    capture_output=True,
    text=True
)

print(result.stdout)
if result.stderr:
    print("STDERR:", result.stderr)
