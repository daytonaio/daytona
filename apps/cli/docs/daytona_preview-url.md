## daytona preview-url

Get signed preview URL for a sandbox port

```
daytona preview-url [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona preview-url my-sandbox --port 3000
  daytona preview-url my-sandbox --port 3000 --expires 7200
  daytona preview-url my-sandbox --port 3000 --format json
```

### Options

```
      --expires int32   URL expiration time in seconds (default 3600)
  -f, --format string   Output format. Must be one of (yaml, json)
  -p, --port int32      Port number to get preview URL for (required)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
