## daytona start

Start a sandbox

```
daytona start [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona start my-sandbox
  daytona start my-sandbox --wait --timeout 2m
  daytona start my-sandbox --wait --format json
```

### Options

```
  -f, --format string      Output format. Must be one of (yaml, json)
      --timeout duration   Maximum time to wait with --wait (0 waits indefinitely) (default 5m0s)
      --wait               Wait until the sandbox is started
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
