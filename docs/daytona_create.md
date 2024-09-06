## daytona create

Create a workspace

```
daytona create [REPOSITORY_URL] [flags]
```

### Options

```
      --blank                      Create a blank project without using existing configurations
      --branch string              Specify the Git branch to use in the project
      --builder BuildChoice        Specify the builder (currently auto/devcontainer/none)
  -c, --code                       Open the workspace in the IDE after workspace creation
      --custom-image string        Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well
      --custom-image-user string   Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well
      --devcontainer-path string   Automatically assign the devcontainer builder with the path passed as the flag value
      --env stringArray            Specify environment variables (e.g. --env 'KEY1=VALUE1' --env 'KEY2=VALUE2' ...')
  -i, --ide string                 Specify the IDE (vscode, browser, cursor, ssh, jupyter, clion, goland, intellij, phpstorm, pycharm, rider, rubymine, webstorm)
      --manual                     Manually enter the Git repository
      --multi-project              Workspace with multiple projects/repos
      --name string                Specify the workspace name
  -t, --target string              Specify the target (e.g. 'local')
  -y, --yes                        Automatically confirm any prompts
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona](daytona.md)	 - Daytona is a Dev Environment Manager

