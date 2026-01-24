import subprocess
import json

# Query all docker registry credentials
query = """
SELECT 
    id, 
    url, 
    username, 
    LENGTH(password) as password_length,
    "organizationId",
    "createdAt",
    "updatedAt"
FROM docker_registry
ORDER BY "createdAt";
"""

result = subprocess.run(
    ['docker', 'exec', 'daytona-db-1', 'psql', '-U', 'user', '-d', 'daytona', '-t', '-A', '-F', '|', '-c', query],
    capture_output=True,
    text=True
)

print("=" * 100)
print("데이터베이스 등록 Docker Registry Credentials")
print("=" * 100)

if result.returncode == 0:
    lines = result.stdout.strip().split('\n')
    
    print(f"\n총 {len(lines)}개의 레지스트리가 등록되어 있습니다.\n")
    
    for i, line in enumerate(lines, 1):
        if line.strip():
            parts = line.split('|')
            if len(parts) >= 5:
                print(f"\n[레지스트리 #{i}]")
                print(f"  ID: {parts[0]}")
                print(f"  URL: {parts[1]}")
                print(f"  Username: {parts[2]}")
                print(f"  Password Length: {parts[3]} 자")
                print(f"  Organization ID: {parts[4]}")
                if len(parts) > 5:
                    print(f"  Created At: {parts[5]}")
                if len(parts) > 6:
                    print(f"  Updated At: {parts[6]}")
else:
    print(f"Error: {result.stderr}")

print("\n" + "=" * 100)
