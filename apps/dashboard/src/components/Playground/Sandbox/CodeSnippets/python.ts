/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeSnippetGenerator } from './types'
import { joinGroupedSections } from './utils'

export const PythonSnippetGenerator: CodeSnippetGenerator = {
  getImports(p) {
    return (
      [
        'from daytona import Daytona',
        p.actions.useConfigObject ? 'DaytonaConfig' : '',
        p.config.useSandboxCreateParams
          ? p.config.createSandboxFromSnapshot
            ? 'CreateSandboxFromSnapshotParams'
            : 'CreateSandboxFromImageParams'
          : '',
        p.config.useResources ? 'Resources' : '',
        p.config.createSandboxFromImage ? 'Image' : '',
      ]
        .filter(Boolean)
        .join(', ') + '\n'
    )
  },

  getConfig(p) {
    if (!p.actions.useConfigObject) return ''
    return ['\n# Define the configuration', 'config = DaytonaConfig()'].filter(Boolean).join('\n') + '\n'
  },

  getClientInit(p) {
    return ['# Initialize the Daytona client', `daytona = Daytona(${p.actions.useConfigObject ? 'config' : ''})`]
      .filter(Boolean)
      .join('\n')
  },

  getResources(p) {
    if (!p.config.useResources) return ''
    const ind = '\t'
    return [
      '\n\n# Create a Sandbox with custom resources\nresources = Resources(',
      p.config.useResourcesCPU
        ? `${ind}cpu=${p.state['resources']['cpu']}, # ${p.state['resources']['cpu']} CPU cores`
        : '',
      p.config.useResourcesMemory
        ? `${ind}memory=${p.state['resources']['memory']}, # ${p.state['resources']['memory']}GB RAM`
        : '',
      p.config.useResourcesDisk
        ? `${ind}disk=${p.state['resources']['disk']}, # ${p.state['resources']['disk']}GB disk space`
        : '',
      ')',
    ]
      .filter(Boolean)
      .join('\n')
  },

  getSandboxParams(p) {
    if (!p.config.useSandboxCreateParams) return ''
    const ind = '\t'
    return [
      `\n\nparams = ${p.config.createSandboxFromSnapshot ? 'CreateSandboxFromSnapshotParams' : 'CreateSandboxFromImageParams'}(`,
      p.config.useCustomSandboxSnapshotName ? `${ind}snapshot="${p.state['snapshotName']}",` : '',
      p.config.createSandboxFromImage ? `${ind}image=Image.debian_slim("3.13"),` : '',
      p.config.useResources ? `${ind}resources=resources,` : '',
      p.config.useLanguageParam ? `${ind}language="${p.state['language']}",` : '',
      ...(p.config.createSandboxParamsExist
        ? [
            p.config.useAutoStopInterval
              ? `${ind}auto_stop_interval=${p.state['createSandboxBaseParams']['autoStopInterval']}, # ${p.state['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${p.state['createSandboxBaseParams']['autoStopInterval']} minute${(p.state['createSandboxBaseParams']['autoStopInterval'] as number) > 1 ? 's' : ''}`}`
              : '',
            p.config.useAutoArchiveInterval
              ? `${ind}auto_archive_interval=${p.state['createSandboxBaseParams']['autoArchiveInterval']}, # Auto-archive after a Sandbox has been stopped for ${p.state['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${p.state['createSandboxBaseParams']['autoArchiveInterval']} minutes`}`
              : '',
            p.config.useAutoDeleteInterval
              ? `${ind}auto_delete_interval=${p.state['createSandboxBaseParams']['autoDeleteInterval']}, # ${p.state['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : p.state['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${p.state['createSandboxBaseParams']['autoDeleteInterval']} minutes`}`
              : '',
          ]
        : []),
      ')',
    ]
      .filter(Boolean)
      .join('\n')
  },

  getSandboxCreate(p) {
    return [
      '\n# Create the Sandbox instance',
      `sandbox = daytona.create(${p.config.useSandboxCreateParams ? 'params' : ''})`,
      'print(f"Sandbox created:{sandbox.id}")',
    ].join('\n')
  },

  getCodeRun(p) {
    if (!p.actions.codeToRunExists) return ''
    const ind = '\t'
    return [
      '\n\n# Run code securely inside the Sandbox',
      'codeRunResponse = sandbox.process.code_run(',
      `'''${p.state['codeRunParams'].languageCode}'''`,
      ')',
      'if codeRunResponse.exit_code != 0:',
      `${ind}print(f"Error: {codeRunResponse.exit_code} {codeRunResponse.result}")`,
      'else:',
      `${ind}print(codeRunResponse.result)`,
    ].join('\n')
  },

  getShellRun(p) {
    if (!p.actions.shellCommandExists) return ''
    return [
      '\n\n# Execute shell commands',
      `shellRunResponse = sandbox.process.exec("${p.state['shellCommandRunParams'].shellCommand}")`,
      'print(shellRunResponse.result)',
    ].join('\n')
  },

  getFileSystemOps(p) {
    const sections: string[] = []
    const ind = '\t'

    if (p.actions.fileSystemCreateFolderParamsSet) {
      sections.push(
        [
          '# Create folder with specific permissions',
          `sandbox.fs.create_folder("${p.state['createFolderParams'].folderDestinationPath}", "${p.state['createFolderParams'].permissions}")`,
        ].join('\n'),
      )
    }

    if (p.actions.fileSystemListFilesLocationSet) {
      sections.push(
        [
          '# List files in a directory',
          `files = sandbox.fs.list_files("${p.state['listFilesParams'].directoryPath}")`,
          'for file in files:',
          `${ind}print(f"Name: {file.name}")`,
          `${ind}print(f"Is directory: {file.is_dir}")`,
          `${ind}print(f"Size: {file.size}")`,
          `${ind}print(f"Modified: {file.mod_time}")`,
        ].join('\n'),
      )
    }

    if (p.actions.fileSystemDeleteFileRequiredParamsSet) {
      sections.push(
        [
          `# Delete ${p.actions.useFileSystemDeleteFileRecursive ? 'directory' : 'file'}`,
          `sandbox.fs.delete_file("${p.state['deleteFileParams'].filePath}"${p.actions.useFileSystemDeleteFileRecursive ? ', True' : ''})`,
        ].join('\n'),
      )
    }

    return joinGroupedSections(sections)
  },

  getGitOps(p) {
    const sections: string[] = []
    const ind = '\t'

    if (p.actions.gitCloneOperationRequiredParamsSet) {
      sections.push(
        [
          '# Clone git repository',
          'sandbox.git.clone(',
          `${ind}url="${p.state['gitCloneParams'].repositoryURL}",`,
          `${ind}path="${p.state['gitCloneParams'].cloneDestinationPath}",`,
          p.actions.useGitCloneBranch ? `${ind}branch="${p.state['gitCloneParams'].branchToClone}",` : '',
          p.actions.useGitCloneCommitId ? `${ind}commit_id="${p.state['gitCloneParams'].commitToClone}",` : '',
          p.actions.useGitCloneUsername ? `${ind}username="${p.state['gitCloneParams'].authUsername}",` : '',
          p.actions.useGitClonePassword ? `${ind}password="${p.state['gitCloneParams'].authPassword}"` : '',
          ')',
        ]
          .filter(Boolean)
          .join('\n'),
      )
    }

    if (p.actions.gitStatusOperationLocationSet) {
      sections.push(
        [
          '# Get repository status',
          `status = sandbox.git.status("${p.state['gitStatusParams'].repositoryPath}")`,
          'print(f"Current branch: {status.current_branch}")',
          'print(f"Commits ahead: {status.ahead}")',
          'print(f"Commits behind: {status.behind}")',
          'for file_status in status.file_status:',
          '\tprint(f"File: {file_status.name}")',
        ].join('\n'),
      )
    }

    if (p.actions.gitBranchesOperationLocationSet) {
      sections.push(
        [
          '# List branches',
          `branchesResponse = sandbox.git.branches("${p.state['gitBranchesParams'].repositoryPath}")`,
          'for branch in branchesResponse.branches:',
          '\tprint(f"Branch: {branch}")',
        ].join('\n'),
      )
    }

    return joinGroupedSections(sections)
  },

  buildFullSnippet(p) {
    const imports = this.getImports(p)
    const config = this.getConfig(p)
    const client = this.getClientInit(p)
    const resources = this.getResources(p)
    const params = this.getSandboxParams(p)
    const create = this.getSandboxCreate(p)
    const codeRun = this.getCodeRun(p)
    const shell = this.getShellRun(p)
    const fsOps = this.getFileSystemOps(p)
    const gitOps = this.getGitOps(p)

    return `${imports}${config}\n${client}${resources}${params}\n${create}${codeRun}${shell}${fsOps}${gitOps}`
  },
}
