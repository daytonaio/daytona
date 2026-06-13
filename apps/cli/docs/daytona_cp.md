## daytona cp

Copy files between the local machine and a sandbox

### Synopsis

Copy files or directories between the local filesystem and a sandbox.

Exactly one of SOURCE or DESTINATION must reference a sandbox path using the
<sandbox>:<path> form, where <sandbox> is a sandbox ID or name. Directories
are copied recursively. Copying into an existing directory places the source
basename inside it, and missing parent directories are created.

```
daytona cp SOURCE DESTINATION [flags]
```

### Examples

```
  daytona cp ./config.yaml my-sandbox:/workspace/config.yaml
  daytona cp my-sandbox:/workspace/output ./output
  daytona cp ./src my-sandbox:/workspace/src --format json
```

### Options

```
  -f, --format string   Output format. Must be one of (yaml, json)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
