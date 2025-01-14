## daytona create

Create a workspace

```
daytona create [REPOSITORY_URL | WORKSPACE_CONFIG_NAME]... [flags]
```

### Options

```
      --blank                        Create a blank workspace without using existing templates
      --branch strings               Specify the Git branches to use in the workspaces
      --builder BuildChoice          Specify the builder (currently auto/devcontainer/none)
      --custom-image string          Create the workspace with the custom image passed as the flag value; Requires setting --custom-image-user flag as well
      --custom-image-user string     Create the workspace with the custom image user passed as the flag value; Requires setting --custom-image flag as well
      --devcontainer-path string     Automatically assign the devcontainer builder with the path passed as the flag value
      --env stringArray              Specify environment variables (e.g. --env 'KEY1=VALUE1' --env 'KEY2=VALUE2' ...')
      --git-provider-config string   Specify the Git provider configuration ID or alias
  -i, --ide string                   Specify the IDE (vscode, code-insiders, browser, cursor, codium, codium-insiders, ssh, jupyter, fleet, positron, zed, windsurf, clion, goland, intellij, phpstorm, pycharm, rider, rubymine, webstorm)
      --manual                       Manually enter the Git repository
      --multi-workspace              Target with multiple workspaces/repos
  -n, --no-ide                       Do not open the target in the IDE after target creation
  -t, --target string                Specify the target (e.g. 'local')
  -y, --yes                          Automatically confirm any prompts
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona](daytona.md)	 - Daytona is a Dev Environment Manager

