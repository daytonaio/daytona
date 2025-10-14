# TypeScript Disk Management Example

This example demonstrates how to use the Daytona TypeScript SDK to manage disks.

## Prerequisites

1. Install dependencies:

   ```bash
   npm install
   ```

2. Set environment variables:

   ```bash
   export DAYTONA_API_KEY="your-api-key"
   export DAYTONA_API_URL="https://app.daytona.io/api"
   export DAYTONA_TARGET="us"
   ```

## Running the Example

```bash
npm start
```

Or directly with tsx:

```bash
npx tsx index.ts
```

## What This Example Does

1. Lists all existing disks
2. Creates a new 20GB disk
3. Retrieves the disk details by ID
4. Lists disks again to show the new one
5. Deletes the created disk
6. Confirms deletion with a final list

The example includes proper error handling and cleanup.
