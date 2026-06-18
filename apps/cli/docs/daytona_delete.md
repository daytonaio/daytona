## daytona delete

Delete a sandbox

```
daytona delete [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona delete my-sandbox
  daytona delete my-sandbox --wait --format json
  daytona delete --all --dry-run
  daytona delete --all --yes
```

### Options

```
  -a, --all                Delete all sandboxes
      --dry-run            Show what would be deleted without deleting anything
  -f, --format string      Output format. Must be one of (yaml, json)
      --ignore-not-found   Treat a missing sandbox as a successful delete
      --timeout duration   Maximum time to wait with --wait (0 waits indefinitely) (default 5m0s)
      --wait               Wait until the sandbox is fully deleted
  -y, --yes                Skip the confirmation prompt for bulk deletes
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
