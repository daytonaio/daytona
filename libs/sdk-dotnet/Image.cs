// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text;
using System.Text.Json;

namespace Daytona.Sdk;

public class Image
{
    private readonly StringBuilder _dockerfileBuilder = new();

    public static Image Base(string baseImage)
    {
        var image = new Image();
        image._dockerfileBuilder.AppendLine($"FROM {baseImage}");
        return image;
    }

    public static Image DebianSlim(string pythonVersion)
    {
        var safeVersion = string.IsNullOrWhiteSpace(pythonVersion) ? "3.12" : pythonVersion;
        return Base($"python:{safeVersion}-slim");
    }

    public Image PipInstall(params string[] packages)
    {
        if (packages.Length == 0)
        {
            return this;
        }

        _dockerfileBuilder.Append("RUN python -m pip install");
        foreach (var pkg in packages)
        {
            _dockerfileBuilder.Append(' ').Append(pkg);
        }

        _dockerfileBuilder.AppendLine();
        return this;
    }

    public Image RunCommands(params string[] commands)
    {
        foreach (var command in commands)
        {
            _dockerfileBuilder.AppendLine($"RUN {command}");
        }

        return this;
    }

    public Image Env(Dictionary<string, string> envVars)
    {
        foreach (var (key, value) in envVars)
        {
            _dockerfileBuilder.AppendLine($"ENV {key}={EscapeEnvValue(value)}");
        }

        return this;
    }

    public Image Workdir(string path)
    {
        _dockerfileBuilder.AppendLine($"WORKDIR {path}");
        return this;
    }

    public Image Entrypoint(params string[] commands)
    {
        _dockerfileBuilder.AppendLine($"ENTRYPOINT {ToJsonArray(commands)}");
        return this;
    }

    public Image Cmd(params string[] commands)
    {
        _dockerfileBuilder.AppendLine($"CMD {ToJsonArray(commands)}");
        return this;
    }

    public string Dockerfile => _dockerfileBuilder.ToString();

    private static string EscapeEnvValue(string value) => value.Replace("\n", "\\n").Replace("\"", "\\\"");

    private static string ToJsonArray(string[] values) => JsonSerializer.Serialize(values);
}