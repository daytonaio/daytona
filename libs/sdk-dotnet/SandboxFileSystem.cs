// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk.Models;
using Daytona.ToolboxApiClient.Client;
using FileParameterModel = Daytona.ToolboxApiClient.Client.FileParameter;
using FileSystemApi = Daytona.ToolboxApiClient.Api.FileSystemApi;
using ReplaceRequestModel = Daytona.ToolboxApiClient.Model.ReplaceRequest;

namespace Daytona.Sdk;

public class SandboxFileSystem
{
    private readonly FileSystemApi _api;

    internal SandboxFileSystem(FileSystemApi api)
    {
        _api = api;
    }

    public Task CreateFolderAsync(string path, string mode = "755", CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.CreateFolderAsync(path, mode, ct),
            ct
        );

    public Task DeleteFileAsync(string path, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.DeleteFileAsync(path, cancellationToken: ct),
            ct
        );

    public async Task<byte[]> DownloadFileAsync(string remotePath, CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.DownloadFileAsync(remotePath, ct),
            ct
        );

        await using var stream = response.Content;
        using var ms = new MemoryStream();
        await stream.CopyToAsync(ms, ct);
        return ms.ToArray();
    }

    public Task UploadFileAsync(byte[] content, string remotePath, CancellationToken ct = default)
    {
        var fileName = Path.GetFileName(remotePath);
        var fileParam = new FileParameterModel(fileName, new MemoryStream(content, writable: false));

        return GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.UploadFileAsync(remotePath, fileParam, ct),
            ct
        );
    }

    public async Task<List<FileInfoModel>> ListFilesAsync(string path, CancellationToken ct = default)
    {
        var files = await GeneratedClientSupport.ExecuteToolboxAsync(async () =>
        {
            var requestOptions = new RequestOptions();
            requestOptions.HeaderParameters.Add("Accept", "application/json");
            requestOptions.QueryParameters.Add(ClientUtils.ParameterToMultiMap("", "path", path));

            var response = await _api.AsynchronousClient.GetAsync<List<Daytona.ToolboxApiClient.Model.FileInfo>>(
                "/files/",
                requestOptions,
                _api.Configuration,
                ct
            );
            return response.Data;
        }, ct);

        return files.Select(GeneratedClientSupport.ToSdkFileInfo).ToList();
    }

    public async Task<List<MatchResult>> FindFilesAsync(string path, string pattern, CancellationToken ct = default)
    {
        var matches = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.FindInFilesAsync(path, pattern, ct),
            ct
        );

        return matches.Select(GeneratedClientSupport.ToSdkMatch).ToList();
    }

    public async Task<SearchResult> SearchFilesAsync(string path, string pattern, CancellationToken ct = default)
    {
        var search = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.SearchFilesAsync(path, pattern, ct),
            ct
        );

        return GeneratedClientSupport.ToSdkSearchResult(search);
    }

    public Task ReplaceInFilesAsync(List<string> files, string pattern, string newValue, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.ReplaceInFilesAsync(new ReplaceRequestModel(files, newValue, pattern), ct),
            ct
        );

    public Task MoveFilesAsync(string source, string destination, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.MoveFileAsync(source, destination, ct),
            ct
        );
}