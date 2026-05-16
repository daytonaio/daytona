// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { createApiResponse } from './helpers'
import { Git } from '../Git'

jest.mock('@daytona/toolbox-api-client', () => ({}), { virtual: true })

describe('Git', () => {
  const apiClient = {
    addFiles: jest.fn(),
    listBranches: jest.fn(),
    createBranch: jest.fn(),
    deleteBranch: jest.fn(),
    checkoutBranch: jest.fn(),
    cloneRepository: jest.fn(),
    commitChanges: jest.fn(),
    pushChanges: jest.fn(),
    pullChanges: jest.fn(),
    getStatus: jest.fn(),
  }

  const git = new Git(apiClient as unknown as never)

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('delegates add/branches/branch operations', async () => {
    apiClient.listBranches.mockResolvedValue(createApiResponse({ branches: ['main'], current: 'main' }))

    await git.add('/repo', ['a.ts'])
    await expect(git.branches('/repo')).resolves.toEqual({ branches: ['main'], current: 'main' })
    await git.createBranch('/repo', 'feature/a')
    await git.deleteBranch('/repo', 'feature/a')
    await git.checkoutBranch('/repo', 'main')

    expect(apiClient.addFiles).toHaveBeenCalledWith({ path: '/repo', files: ['a.ts'] })
  })

  it('delegates clone/commit/push/pull/status', async () => {
    apiClient.commitChanges.mockResolvedValue(createApiResponse({ hash: 'abc123' }))
    apiClient.getStatus.mockResolvedValue(createApiResponse({ currentBranch: 'main', ahead: 0, behind: 0 }))

    await git.clone('https://github.com/org/repo.git', '/repo', 'main', 'c1', 'u', 'p')
    await expect(git.commit('/repo', 'msg', 'author', 'author@example.com', true)).resolves.toEqual({ sha: 'abc123' })
    await git.push('/repo', 'u', 'p')
    await git.pull('/repo', 'u', 'p')
    await expect(git.status('/repo')).resolves.toEqual({ currentBranch: 'main', ahead: 0, behind: 0 })
  })

  it('passes the full clone payload including auth and commit id', async () => {
    await git.clone('https://github.com/org/repo.git', '/repo', 'develop', 'deadbeef', 'alice', 'secret')

    expect(apiClient.cloneRepository).toHaveBeenCalledWith({
      url: 'https://github.com/org/repo.git',
      branch: 'develop',
      path: '/repo',
      username: 'alice',
      password: 'secret',
      commit_id: 'deadbeef',
      depth: undefined,
      single_branch: undefined,
      shallow_since: undefined,
      no_tags: undefined,
      filter: undefined,
      sparse: undefined,
      sparse_paths: undefined,
      reference_path: undefined,
      dissociate: undefined,
      recurse_submodules: undefined,
      shallow_submodules: undefined,
      filter_submodules: undefined,
    })
  })

  it('passes advanced clone options', async () => {
    await git.clone('https://github.com/org/repo.git', '/repo', {
      branch: 'main',
      commitId: 'abc123',
      username: 'alice',
      password: 'secret',
      depth: 1,
      singleBranch: false,
      shallowSince: '2025-01-01',
      noTags: true,
      filter: 'blob:none',
      sparse: true,
      sparsePaths: ['src', 'README.md'],
      referencePath: '/cache/repo.git',
      dissociate: true,
      recurseSubmodules: true,
      shallowSubmodules: true,
      filterSubmodules: true,
    })

    expect(apiClient.cloneRepository).toHaveBeenCalledWith({
      url: 'https://github.com/org/repo.git',
      branch: 'main',
      path: '/repo',
      username: 'alice',
      password: 'secret',
      commit_id: 'abc123',
      depth: 1,
      single_branch: false,
      shallow_since: '2025-01-01',
      no_tags: true,
      filter: 'blob:none',
      sparse: true,
      sparse_paths: ['src', 'README.md'],
      reference_path: '/cache/repo.git',
      dissociate: true,
      recurse_submodules: true,
      shallow_submodules: true,
      filter_submodules: true,
    })
  })

  it('omits optional clone params when they are not provided', async () => {
    await git.clone('https://github.com/org/repo.git', '/repo')

    expect(apiClient.cloneRepository).toHaveBeenCalledWith({
      url: 'https://github.com/org/repo.git',
      branch: undefined,
      path: '/repo',
      username: undefined,
      password: undefined,
      commit_id: undefined,
      depth: undefined,
      single_branch: undefined,
      shallow_since: undefined,
      no_tags: undefined,
      filter: undefined,
      sparse: undefined,
      sparse_paths: undefined,
      reference_path: undefined,
      dissociate: undefined,
      recurse_submodules: undefined,
      shallow_submodules: undefined,
      filter_submodules: undefined,
    })
  })

  it('maps commit hashes to sha responses', async () => {
    apiClient.commitChanges.mockResolvedValue(createApiResponse({ hash: 'feedface' }))

    await expect(git.commit('/repo', 'msg', 'Alice', 'alice@example.com')).resolves.toEqual({ sha: 'feedface' })
    expect(apiClient.commitChanges).toHaveBeenCalledWith({
      path: '/repo',
      message: 'msg',
      author: 'Alice',
      email: 'alice@example.com',
      allow_empty: undefined,
    })
  })

  it('forwards push and pull credentials', async () => {
    await git.push('/repo', 'alice', 'secret')
    await git.pull('/repo', 'alice', 'secret')

    expect(apiClient.pushChanges).toHaveBeenCalledWith({ path: '/repo', username: 'alice', password: 'secret' })
    expect(apiClient.pullChanges).toHaveBeenCalledWith({ path: '/repo', username: 'alice', password: 'secret' })
  })

  it('propagates api client failures', async () => {
    const error = new Error('git failed')
    apiClient.getStatus.mockRejectedValue(error)

    await expect(git.status('/repo')).rejects.toBe(error)
  })
})
