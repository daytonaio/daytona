## dtn ssh

SSH into a project using the terminal or edit its SSH config


```
dtn ssh [WORKSPACE] [PROJECT] [CMD...] [flags]
```

### Options

```
  -o, --option stringArray Specify SSH options in KEY=VALUE format.
  -y, --yes Automatically confirm any prompts
  -edit Edit SSH config for the specified project
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona](daytona.md)	 - Daytona is a Dev Environment Manager


### Key Changes

- **Command Name**: Updated from `daytona ssh` to `dtn ssh`.
- **New Option**: Added `-edit` flag for editing SSH configurations directly from the CLI.