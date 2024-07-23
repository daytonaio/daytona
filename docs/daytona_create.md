## daytona create

Create a workspace

```
daytona create [REPOSITORY_URL] [flags]
```

### Options

```
      --blank                      Create a blank project without using existing configurations
      --builder BuildChoice        Specify the builder (currently auto/devcontainer/none)
  -c, --code                       Open the workspace in the IDE after workspace creation
      --custom-image string        Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well
      --custom-image-user string   Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well
      --devcontainer-path string   Automatically assign the devcontainer builder with the path passed as the flag value
  -i, --ide string                 Specify the IDE ('vscode' or 'browser')
      --manual                     Manually enter the git repositories
      --multi-project              Workspace with multiple projects/repos
      --name string                Specify the workspace name
      --provider string            Specify the provider (e.g. 'docker-provider')
  -t, --target string              Specify the target (e.g. 'local')
```

### Options inherited from parent commands

```
      --help            help for daytona
  -o, --output string   Output format. Must be one of (yaml, json)
```

### SEE ALSO

* [daytona](daytona.md)	 - Daytona is a Dev Environment Manager

