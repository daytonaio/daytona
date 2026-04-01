// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk.Models;
using GitAddRequestModel = Daytona.ToolboxApiClient.Model.GitAddRequest;
using GitApi = Daytona.ToolboxApiClient.Api.GitApi;
using GitCloneRequestModel = Daytona.ToolboxApiClient.Model.GitCloneRequest;
using GitCommitRequestModel = Daytona.ToolboxApiClient.Model.GitCommitRequest;
using GitRepoRequestModel = Daytona.ToolboxApiClient.Model.GitRepoRequest;

namespace Daytona.Sdk;

public class SandboxGit
{
    private readonly GitApi _api;

    internal SandboxGit(GitApi api)
    {
        _api = api;
    }

    public Task CloneAsync(
        string url,
        string path,
        string? branch = null,
        string? commitId = null,
        string? username = null,
        string? password = null,
        CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.CloneRepositoryAsync(new GitCloneRequestModel(branch!, commitId!, password!, path, url, username!), ct),
            ct
        );

    public async Task<BranchesResponse> BranchesAsync(string path, CancellationToken ct = default)
    {
        var branches = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.ListBranchesAsync(path, ct),
            ct
        );
        return GeneratedClientSupport.ToSdkBranchesResponse(branches);
    }

    public Task AddAsync(string path, List<string> files, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.AddFilesAsync(new GitAddRequestModel(files, path), ct),
            ct
        );

    public async Task<GitCommitResponse> CommitAsync(string path, string message, string author, string email, CancellationToken ct = default)
    {
        var commit = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.CommitChangesAsync(new GitCommitRequestModel(false, author, email, message, path), ct),
            ct
        );
        return GeneratedClientSupport.ToSdkGitCommitResponse(commit);
    }

    public async Task<GitStatus> StatusAsync(string path, CancellationToken ct = default)
    {
        var status = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.GetStatusAsync(path, ct),
            ct
        );
        return GeneratedClientSupport.ToSdkGitStatus(status);
    }

    public Task PushAsync(string path, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.PushChangesAsync(new GitRepoRequestModel(path: path), ct),
            ct
        );

    public Task PullAsync(string path, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.PullChangesAsync(new GitRepoRequestModel(path: path), ct),
            ct
        );
}