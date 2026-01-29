import type { APIRoute } from 'astro'
import fs from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

function findWorkspaceRoot(startPath: string): string {
  let current = resolve(startPath)
  const root = resolve(current, '/')

  while (current !== root) {
    const nxJson = join(current, 'nx.json')

    if (fs.existsSync(nxJson)) {
      return current
    }

    const parent = resolve(current, '..')
    if (parent === current) break
    current = parent
  }

  return process.cwd()
}

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const workspaceRoot = findWorkspaceRoot(__dirname)
const toolboxApiPath = join(
  workspaceRoot,
  'apps/daemon/pkg/toolbox/docs/swagger.json'
)
const toolboxApiSpec = JSON.parse(fs.readFileSync(toolboxApiPath, 'utf-8'))

export const GET: APIRoute = () => {
  return new Response(JSON.stringify(toolboxApiSpec, null, 2), {
    headers: {
      'Content-Type': 'application/json',
      'Content-Disposition': 'attachment; filename="toolbox-openapi.json"',
    },
  })
}
