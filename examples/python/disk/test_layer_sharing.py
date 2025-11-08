#!/usr/bin/env python3
"""Test layer sharing between source and forked disks"""
import json
import subprocess
import time

from daytona import CreateSandboxFromImageParams, Daytona

daytona = Daytona()

print("1. Creating source disk...")
source_disk = daytona.disk.create(f"source-{int(time.time())}", 10)
print(f"   Source disk ID: {source_disk.id}")

print("\n2. Creating sandbox with source disk...")
sandbox = daytona.create(
    CreateSandboxFromImageParams(
        image="ubuntu:22.04",
        sandbox_name=f"source-sandbox-{int(time.time())}",
        disk_id=source_disk.id,
        language="python",
    ),
    timeout=120,
)

print("\n3. Writing test file...")
with open("/tmp/test.txt", "w") as f:
    f.write("Test content")
sandbox.fs.upload_file("/tmp/test.txt", "/workspace/test.txt")
time.sleep(2)

print("\n4. Stopping sandbox...")
sandbox.stop()
time.sleep(5)

print("\n5. Getting source disk info...")
result = subprocess.run(
    ["daytona", "disk", "info", source_disk.id, "--json"], capture_output=True, text=True, check=False
)
source_info = json.loads(result.stdout)
print(f"   Source disk layers: {len(source_info['layers'])} layer(s)")
for layer in source_info["layers"]:
    print(f"     - {layer['id']}: {layer['size']} bytes")

print("\n6. Forking disk...")
forked_disk = daytona.disk.fork(source_disk, f"forked-{int(time.time())}")
time.sleep(5)

print("\n7. Getting forked disk info...")
result = subprocess.run(
    ["daytona", "disk", "info", forked_disk.id, "--json"], capture_output=True, text=True, check=False
)
forked_info = json.loads(result.stdout)
print(f"   Forked disk layers: {len(forked_info['layers'])} layer(s)")
for layer in forked_info["layers"]:
    print(f"     - {layer['id']}: {layer['size']} bytes")

print("\n8. Comparing layer IDs...")
source_layer_ids = {l["id"] for l in source_info["layers"]}
forked_layer_ids = {l["id"] for l in forked_info["layers"]}
shared_layers = source_layer_ids & forked_layer_ids

print(f"\n✅ Shared layers: {len(shared_layers)} out of {len(source_layer_ids)}")
if shared_layers:
    print("   Shared layer IDs:")
    for layer_id in shared_layers:
        print(f"     - {layer_id}")
else:
    print("   ⚠️ No layers are being shared!")

print("\n📝 Cleanup:")
print(f"   daytona disk delete {source_disk.id}")
print(f"   daytona disk delete {forked_disk.id}")
