# Disk Management Examples

This directory contains examples demonstrating how to use the Daytona SDK to manage disks.

## Examples

### TypeScript

- **`typescript/disk/index.ts`** - Basic disk operations (create, list, get, delete)

### Python

- **`python/disk/disk.py`** - Synchronous disk operations
- **`python/disk/_async/disk.py`** - Asynchronous disk operations

## Prerequisites

1. **Daytona SDK installed**:
   - TypeScript: `npm install @daytonaio/sdk`
   - Python: `pip install daytona`

2. **Environment variables set**:

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"
   export DAYTONA_TARGET="us"
   ```

   Or for JWT authentication:

   ```bash
   export DAYTONA_JWT_TOKEN="your-jwt-token"
   export DAYTONA_ORGANIZATION_ID="your-org-id"
   export DAYTONA_API_URL="https://app.daytona.io/api"
   export DAYTONA_TARGET="us"
   ```

## Running the Examples

### TypeScript

```bash
cd examples/typescript/disk
npx tsx index.ts
```

### Python (Synchronous)

```bash
cd examples/python/disk
python disk.py
```

### Python (Asynchronous)

```bash
cd examples/python/disk/_async
python disk.py
```

## What the Examples Do

Each example demonstrates the following disk operations:

1. **List Disks** - Shows all existing disks in your organization
2. **Create Disk** - Creates a new 20GB disk with a unique name
3. **Get Disk** - Retrieves the disk details by ID
4. **List Again** - Shows the updated list including the new disk
5. **Delete Disk** - Removes the created disk
6. **Final List** - Confirms the disk was deleted

## Expected Output

The examples will show output similar to:

```
🚀 Starting Disk Management Example
=====================================

📋 Listing all disks...
Found 0 existing disks:

💾 Creating a new disk...
✅ Created disk: example-disk-1703123456789 (disk-uuid) - 20GB - State: fresh

🔍 Getting disk details...
✅ Retrieved disk: example-disk-1703123456789 - 20GB - State: fresh

📋 Listing disks after creation...
Found 1 disks:
  - example-disk-1703123456789 (disk-uuid) - 20GB - State: fresh

⏳ Waiting 2 seconds before cleanup...

🗑️  Deleting the disk...
✅ Deleted disk: example-disk-1703123456789

📋 Final disk list...
Found 0 disks after cleanup

🎉 Disk management example completed successfully!
```

## Disk States

Disks can be in the following states:

- **`fresh`** - Newly created disk
- **`pulling`** - Disk is being prepared
- **`ready`** - Disk is ready for use
- **`attached`** - Disk is attached to a sandbox
- **`uploading`** - Disk data is being uploaded
- **`stored`** - Disk is stored/deleted

## Notes

- Disks are created with a unique name using timestamp to avoid conflicts
- The examples clean up after themselves by deleting the created disk
- All operations include proper error handling
- The examples use 20GB as the disk size, but you can modify this as needed
