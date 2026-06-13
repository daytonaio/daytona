## daytona create

Create a new sandbox

```
daytona create [flags]
```

### Examples

```
  daytona create --snapshot my-snapshot:1.0 --name my-sandbox
  daytona create --name my-sandbox --env-file .env --if-exists reuse
  daytona create --snapshot my-snapshot:1.0 --name my-sandbox --format json
```

### Options

```
      --auto-archive int32          Auto-archive interval in minutes (0 means the maximum interval will be used) (default 10080)
      --auto-delete int32           Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) (default -1)
      --auto-stop int32             Auto-stop interval in minutes (0 means disabled) (default 15)
  -c, --context stringArray         Files or directories to include in the build context (can be specified multiple times)
      --cpu int32                   CPU cores allocated to the sandbox
      --disk int32                  Disk space allocated to the sandbox in GB
  -f, --dockerfile string           Path to Dockerfile for Sandbox snapshot
  -e, --env stringArray             Environment variables (format: KEY=VALUE)
      --env-file string             Read environment variables from a dotenv-style file (entries from --env override file values)
      --format string               Output format. Must be one of (yaml, json)
      --gpu int32                   GPU units allocated to the sandbox
      --if-exists string            Behavior when a sandbox with the same name already exists (error, reuse; reuse requires --name) (default "error")
  -l, --label stringArray           Labels (format: KEY=VALUE)
      --label-file string           Read labels from a dotenv-style file (entries from --label override file values)
      --memory int32                Memory allocated to the sandbox in MB
      --name string                 Name of the sandbox
      --network-allow-list string   Comma-separated list of allowed CIDR network addresses for the sandbox
      --network-block-all           Whether to block all network access for the sandbox
      --public                      Make sandbox publicly accessible
      --snapshot string             Snapshot to use for the sandbox
      --target string               Target region (eu, us)
      --timeout duration            Maximum time to wait for the sandbox to start (0 means wait indefinitely)
      --user string                 User associated with the sandbox
  -v, --volume stringArray          Volumes to mount (format: VOLUME_ID_OR_NAME:MOUNT_PATH)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
