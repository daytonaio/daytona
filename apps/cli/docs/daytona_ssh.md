## daytona ssh

SSH into a sandbox

### Synopsis

Establish an SSH connection to a running sandbox

```
daytona ssh [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona ssh my-sandbox
  daytona ssh my-sandbox --expires 60
```

### Options

```
      --expires int   SSH access token expiration time in minutes (defaults to 24 hours) (default 1440)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
