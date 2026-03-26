import * as _fs from 'fs'
import { join } from 'path'
import { parseArgs } from 'util'
import * as yaml from 'yaml'

const fs = _fs.promises

const __dirname = import.meta.dirname

// Default local path relative to this script (apps/docs/tools -> apps/cli/hack/docs)
const DEFAULT_LOCAL_PATH = join(__dirname, '../../cli/hack/docs')

// content to appear above the commands outline
const prepend = `---
title: CLI
description: A reference of supported operations using the Daytona CLI.
sidebar:
  label: Daytona CLI Reference
---
import Aside from "@components/Aside.astro";
import Label from "@components/Label.astro";

The \`daytona\` command-line tool provides access to Daytona's core features including managing Snapshots and the lifecycle of Daytona Sandboxes. View the installation instructions by clicking [here](/docs/getting-started#cli).

This reference lists all commands supported by the \`daytona\` command-line tool complete with a description of their behaviour, and any supported flags.
You can access this documentation on a per-command basis by appending the \`--help\`/\`-h\` flag when invoking \`daytona\`.
`

// content to appear below the commands outline
const append = ``

const notes = {
  'daytona autocomplete': `\n<Aside type="note">
If using bash shell environment, make sure you have bash-completion installed in order to get full autocompletion functionality.
Linux Installation: \`\`\`sudo apt-get install bash-completion\`\`\`
macOS Installation: \`\`\`brew install bash-completion\`\`\`
</Aside>`,
}

async function fetchRawDocs(ref) {
  const url =
    'https://api.github.com/repos/daytonaio/daytona/contents/apps/cli/hack/docs'
  const request = await fetch(`${url}?ref=${ref}`)
  const response = await request.json()

  const files = []

  for (const file of response) {
    const { download_url } = file

    if (!download_url) continue

    const contentsReq = await fetch(download_url)
    let contents = await contentsReq.text()

    contents = yaml.parse(contents)

    files.push(contents)
  }

  return files
}

async function readLocalDocs(localPath) {
  const dirEntries = await fs.readdir(localPath)
  const yamlFiles = dirEntries.filter(f => f.endsWith('.yaml'))

  const files = []

  for (const fileName of yamlFiles) {
    const filePath = join(localPath, fileName)
    const contents = await fs.readFile(filePath, 'utf-8')
    files.push(yaml.parse(contents))
  }

  return files
}

function flagToRow(flag) {
  let { name, shorthand, usage } = flag

  name = `\`--${name}\``
  shorthand = shorthand ? `\`-${shorthand}\`` : ''
  usage = usage ? usage : ''
  if (usage.endsWith('\n')) {
    usage = usage.slice(0, -1)
  }

  return `| ${name} | ${shorthand} | ${usage} |\n`
}

function yamlToMarkdown(files) {
  return files.map(rawDoc => {
    let output = ''
    output += `## ${rawDoc.name}\n`
    output += `${rawDoc.synopsis}\n\n`

    if (!rawDoc.usage) {
      rawDoc.usage = `${rawDoc.name} [flags]`
    }

    output += '```shell\n'
    output += `${rawDoc.usage}\n`
    output += '```\n\n'

    output += '__Flags__\n'
    output += '| Long | Short | Description |\n'
    output += '| :--- | :---- | :---------- |\n'

    if (rawDoc.options) {
      for (const flag of rawDoc.options) {
        const row = flagToRow(flag)
        output += row
      }
    }

    if (rawDoc.inherited_options) {
      for (const flag of rawDoc.inherited_options) {
        const row = flagToRow(flag)
        output += row
      }
    }

    if (notes[rawDoc.name]) {
      output += notes[rawDoc.name]
    }

    output += '\n'

    return output
  })
}

async function process(args) {
  const { output, ref, local, path } = args.values

  let files
  if (local || path) {
    const localPath = path || DEFAULT_LOCAL_PATH
    console.log(`reading local docs from ${localPath}...`)
    files = await readLocalDocs(localPath)
  } else {
    console.log(`fetching docs for ${ref} from GitHub...`)
    files = await fetchRawDocs(ref)
  }

  const transformed = yamlToMarkdown(files)

  const singleMarkdown = transformed.join('\n')
  console.log(`writing to '${output}'...`)
  await fs.writeFile(output, `${prepend}\n${singleMarkdown}\n${append}`)
  console.log('done')
}

const commandOpts = {
  ref: {
    type: 'string',
    default: 'main',
  },
  local: {
    type: 'boolean',
    short: 'l',
    default: false,
  },
  path: {
    type: 'string',
    short: 'p',
  },
  output: {
    type: 'string',
    short: 'o',
    default: `${__dirname}/../src/content/docs/en/tools/cli.mdx`,
  },
}

const args = parseArgs({ options: commandOpts })
process(args)
