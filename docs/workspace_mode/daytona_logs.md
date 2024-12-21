## daytona logs

View logs for a workspace/project

### Synopsis

Stream logs from a workspace or project with automatic reconnection support.
Examples:
  # Stream workspace logs
  daytona logs my-workspace

  # Stream project logs with auto-reconnect
  daytona logs my-workspace my-project --retry

  # Stream logs from a specific time
  daytona logs my-workspace --from="2024-12-18T22:00:00Z"

  # Follow logs with reconnection enabled
  daytona logs my-workspace --follow --retry

```
daytona logs [WORKSPACE] [PROJECT_NAME] [flags]
```

### Options

```
  -f, --follow            Follow logs
      --from string       Show logs from this time (RFC3339 format)
      --max-retries int   Maximum number of reconnection attempts (default 5)
      --retry             Enable automatic reconnection (default true)
  -w, --workspace         View workspace logs
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona](daytona.md)	 - Use the Daytona CLI to manage your workspace

