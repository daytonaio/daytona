import { Daytona, SandboxListSortDirection, SandboxListSortField, SandboxState } from '@daytona/sdk'

async function main() {
  const daytona = new Daytona()

  for await (const sandbox of daytona.list({
    limit: 10,
    labels: { env: 'dev' },
    states: [SandboxState.STARTED],
    sort: SandboxListSortField.CREATED_AT,
    order: SandboxListSortDirection.DESC,
  })) {
    console.log(sandbox.id)
  }
}

main().catch(console.error)
