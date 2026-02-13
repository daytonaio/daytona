/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeSnippetGenerator } from './types'
import { joinGroupedSections } from './utils'

export const TypeScriptSnippetGenerator: CodeSnippetGenerator = {
  getImports(p) {
    return (
      [
        'import { Daytona',
        p.actions.useConfigObject ? 'DaytonaConfig' : '',
        p.config.createSandboxFromImage ? 'Image' : '',
      ]
        .filter(Boolean)
        .join(', ') + " } from '@daytonaio/sdk'\n"
    )
  },

  getConfig(p) {
    if (!p.actions.useConfigObject) return ''
    return ['\n// Define the configuration', 'const config: DaytonaConfig = { }'].filter(Boolean).join('\n') + '\n'
  },

  getClientInit(p) {
    return [
      '\t// Initialize the Daytona client',
      `\tconst daytona = new Daytona(${p.actions.useConfigObject ? 'config' : ''})`,
    ]
      .filter(Boolean)
      .join('\n')
  },

  getResources(p) {
    if (!p.config.useResources) return ''
    const ind = '\t\t\t\t'
    return [
      `${ind.slice(0, -1)}resources: {`,
      p.config.useResourcesCPU
        ? `${ind}cpu: ${p.state['resources']['cpu']}, // ${p.state['resources']['cpu']} CPU cores`
        : '',
      p.config.useResourcesMemory
        ? `${ind}memory: ${p.state['resources']['memory']}, // ${p.state['resources']['memory']}GB RAM`
        : '',
      p.config.useResourcesDisk
        ? `${ind}disk: ${p.state['resources']['disk']}, // ${p.state['resources']['disk']}GB disk space`
        : '',
      `${ind.slice(0, -1)}}`,
    ]
      .filter(Boolean)
      .join('\n')
  },

  getSandboxParams(p) {
    if (!p.config.useSandboxCreateParams) return ''
    const ind = '\t\t\t'
    return [
      `{`,
      p.config.useCustomSandboxSnapshotName ? `${ind}snapshot: '${p.state['snapshotName']}',` : '',
      p.config.createSandboxFromImage ? `${ind}image: Image.debianSlim("3.13"),` : '',
      this.getResources(p),
      p.config.useLanguageParam ? `${ind}language: '${p.state['language']}',` : '',
      ...(p.config.createSandboxParamsExist
        ? [
            p.config.useAutoStopInterval
              ? `${ind}autoStopInterval: ${p.state['createSandboxBaseParams']['autoStopInterval']}, // ${p.state['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${p.state['createSandboxBaseParams']['autoStopInterval']} minute${(p.state['createSandboxBaseParams']['autoStopInterval'] as number) > 1 ? 's' : ''}`}`
              : '',
            p.config.useAutoArchiveInterval
              ? `${ind}autoArchiveInterval: ${p.state['createSandboxBaseParams']['autoArchiveInterval']}, // Auto-archive after a Sandbox has been stopped for ${p.state['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${p.state['createSandboxBaseParams']['autoArchiveInterval']} minutes`}`
              : '',
            p.config.useAutoDeleteInterval
              ? `${ind}autoDeleteInterval: ${p.state['createSandboxBaseParams']['autoDeleteInterval']}, // ${p.state['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : p.state['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${p.state['createSandboxBaseParams']['autoDeleteInterval']} minutes`}`
              : '',
          ]
        : []),
      `${ind.slice(0, -1)}}`,
    ]
      .filter(Boolean)
      .join('\n')
  },

  getSandboxCreate(p) {
    return [
      '\t\t// Create the Sandbox instance',
      `\t\tconst sandbox = await daytona.create(${p.config.useSandboxCreateParams ? this.getSandboxParams(p) : ''})`,
    ].join('\n')
  },

  getCodeRun(p) {
    if (!p.actions.codeToRunExists) return ''
    const ind = '\t\t'
    return [
      `\n\n${ind}// Run code securely inside the Sandbox`,
      `${ind}const codeRunResponse = await sandbox.process.codeRun(\``,
      `${(p.state['codeRunParams'].languageCode ?? '').replace(/`/g, '\\`').replace(/\$\{/g, '\\${')}`, // Escape backticks and ${ to prevent breaking the template literal
      `${ind}\`)`,
      `${ind}if (codeRunResponse.exitCode !== 0) {`,
      `${ind + '\t'}console.error("Error running code:", codeRunResponse.exitCode, codeRunResponse.result)`,
      `${ind}} else {`,
      `${ind + '\t'}console.log(codeRunResponse.result)`,
      `${ind}}`,
    ].join('\n')
  },

  getShellRun(p) {
    if (!p.actions.shellCommandExists) return ''
    const ind = '\t\t'
    return [
      `\n\n${ind}// Execute shell commands`,
      `${ind}const shellRunResponse = await sandbox.process.executeCommand('${p.state['shellCommandRunParams'].shellCommand}')`,
      `${ind}console.log(shellRunResponse.result)`,
    ].join('\n')
  },

  getFileSystemOps(p) {
    const sections: string[] = []
    const ind = '\t\t\t'
    const base = ind.slice(0, -1)

    if (p.actions.fileSystemCreateFolderParamsSet) {
      sections.push(
        [
          `${base}// Create folder with specific permissions`,
          `${base}await sandbox.fs.createFolder("${p.state['createFolderParams'].folderDestinationPath}", "${p.state['createFolderParams'].permissions}")`,
        ].join('\n'),
      )
    }

    if (p.actions.fileSystemListFilesLocationSet) {
      sections.push(
        [
          `${base}// List files in a directory`,
          `${base}const files = await sandbox.fs.listFiles("${p.state['listFilesParams'].directoryPath}")`,
          `${base}files.forEach(file => {`,
          `${ind}console.log(\`Name: \${file.name}\`)`,
          `${ind}console.log(\`Is directory: \${file.isDir}\`)`,
          `${ind}console.log(\`Size: \${file.size}\`)`,
          `${ind}console.log(\`Modified: \${file.modTime}\`)`,
          `${base}})`,
        ].join('\n'),
      )
    }

    if (p.actions.fileSystemDeleteFileRequiredParamsSet) {
      sections.push(
        [
          `${base}// Delete ${p.actions.useFileSystemDeleteFileRecursive ? 'directory' : 'file'}`,
          `${base}await sandbox.fs.deleteFile("${p.state['deleteFileParams'].filePath}"${p.actions.useFileSystemDeleteFileRecursive ? ', true' : ''})`,
        ].join('\n'),
      )
    }

    return joinGroupedSections(sections)
  },

  getGitOps(p) {
    const sections: string[] = []
    const ind = '\t\t\t'
    const base = ind.slice(0, -1)

    if (p.actions.gitCloneOperationRequiredParamsSet) {
      sections.push(
        [
          `${base}// Clone git repository`,
          `${base}await sandbox.git.clone(`,
          `${ind}"${p.state['gitCloneParams'].repositoryURL}",`,
          `${ind}"${p.state['gitCloneParams'].cloneDestinationPath}",`,
          p.actions.useGitCloneBranch ? `${ind}"${p.state['gitCloneParams'].branchToClone}",` : '',
          p.actions.useGitCloneCommitId ? `${ind}"${p.state['gitCloneParams'].commitToClone}",` : '',
          p.actions.useGitCloneUsername ? `${ind}"${p.state['gitCloneParams'].authUsername}",` : '',
          p.actions.useGitClonePassword ? `${ind}"${p.state['gitCloneParams'].authPassword}"` : '',
          `${base})`,
        ]
          .filter(Boolean)
          .join('\n'),
      )
    }

    if (p.actions.gitStatusOperationLocationSet) {
      sections.push(
        [
          `${base}// Get repository status`,
          `${base}const status = await sandbox.git.status("${p.state['gitStatusParams'].repositoryPath}")`,
          `${base}console.log(\`Current branch: \${status.currentBranch}\`)`,
          `${base}console.log(\`Commits ahead: \${status.ahead}\`)`,
          `${base}console.log(\`Commits behind: \${status.behind}\`)`,
          `${base}status.fileStatus.forEach(file => {`,
          `${ind}console.log(\`File: \${file.name}\`)`,
          `${base}})`,
        ].join('\n'),
      )
    }

    if (p.actions.gitBranchesOperationLocationSet) {
      sections.push(
        [
          `${base}// List branches`,
          `${base}const branchesResponse = await sandbox.git.branches("${p.state['gitBranchesParams'].repositoryPath}")`,
          `${base}branchesResponse.branches.forEach(branch => {`,
          `${ind}console.log(\`Branch: \${branch}\`)`,
          `${base}})`,
        ].join('\n'),
      )
    }

    return joinGroupedSections(sections)
  },

  buildFullSnippet(p) {
    const imports = this.getImports(p)
    const config = this.getConfig(p)
    const client = this.getClientInit(p)
    const create = this.getSandboxCreate(p)
    const codeRun = this.getCodeRun(p)
    const shell = this.getShellRun(p)
    const fsOps = this.getFileSystemOps(p)
    const gitOps = this.getGitOps(p)

    return `${imports}${config}
async function main() {
${client}
\ttry {
${create}${codeRun}${shell}${fsOps}${gitOps}
\t} catch (error) {
\t\tconsole.error("Sandbox flow error:", error)
\t}
}
main().catch(console.error)`
  },
}
