## daytona project-config add

Add a project config

```
daytona project-config add [flags]
```

### Options

```
      --builder BuildChoice             Specify the builder (currently auto/devcontainer/none)
      --custom-image string             Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well
      --custom-image-user string        Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well
      --devcontainer-path string        Automatically assign the devcontainer builder with the path passed as the flag value
      --env stringArray                 Specify environment variables (e.g. --env 'KEY1=VALUE1' --env 'KEY2=VALUE2' ...')
      --git-provider-config-id string   Specify the Git Provider Configuration Id
      --manual                          Manually enter the Git repository
      --name string                     Specify the project config name
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona project-config](daytona_project-config.md)	 - Manage project configs

