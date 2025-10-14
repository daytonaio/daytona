# Python Disk Management Examples

This directory contains both synchronous and asynchronous examples for disk management using the Daytona Python SDK.

## Prerequisites

1. Install dependencies:

   ```bash
   pip install -r requirements.txt
   ```

2. Set environment variables:

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"
   export DAYTONA_TARGET="us"
   ```

## Running the Examples

### Synchronous Example

```bash
python disk.py
```

### Asynchronous Example

```bash
python _async/disk.py
```

## What These Examples Do

Both examples demonstrate the same disk operations:

1. Lists all existing disks
2. Creates a new 20GB disk
3. Retrieves the disk details by ID
4. Lists disks again to show the new one
5. Deletes the created disk
6. Confirms deletion with a final list

The examples include proper error handling and cleanup.
