## daytona exec

Execute a command in a sandbox

### Synopsis

Execute a command in a running sandbox.

Exits with the remote command's exit code; exit code 255 indicates a CLI-side failure (for example the sandbox was not found or is not running).

```
daytona exec [SANDBOX_ID | SANDBOX_NAME] -- [COMMAND] [ARGS...] [flags]
```

### Examples

```
  daytona exec my-sandbox -- ls -la /workspace
  daytona exec my-sandbox -- python -c "print('hi')"
  daytona exec my-sandbox --cwd /workspace --format json -- npm test
```

### Options

```
      --cwd string      Working directory for command execution
  -f, --format string   Output format. Must be one of (yaml, json)
      --timeout int     Command timeout in seconds (0 for no timeout)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
