## daytona stop

Stop a sandbox

```
daytona stop [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona stop my-sandbox
  daytona stop my-sandbox --wait --timeout 2m
  daytona stop my-sandbox --force --format json
```

### Options

```
  -f, --force              Force stop the sandbox using SIGKILL
      --format string      Output format. Must be one of (yaml, json)
      --timeout duration   Maximum time to wait with --wait (0 waits indefinitely) (default 5m0s)
      --wait               Wait until the sandbox is stopped
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
