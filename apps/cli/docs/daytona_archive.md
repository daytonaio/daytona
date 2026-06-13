## daytona archive

Archive a sandbox

```
daytona archive [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona archive my-sandbox
  daytona archive my-sandbox --wait --timeout 10m
  daytona archive my-sandbox --wait --format json
```

### Options

```
  -f, --format string      Output format. Must be one of (yaml, json)
      --timeout duration   Maximum time to wait with --wait (0 waits indefinitely) (default 5m0s)
      --wait               Wait until the sandbox is archived
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
