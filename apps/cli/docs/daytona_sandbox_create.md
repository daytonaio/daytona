## daytona sandbox create

Create a new sandbox

```
daytona sandbox create [flags]
```

### Options

```
      --auto-stop int32       Auto-stop interval in minutes (0 means disabled)
      --class string          Workspace class type (small, medium, large)
  -c, --context stringArray   Files or directories to include in the build context (can be specified multiple times)
      --cpu int32             CPU cores allocated to the sandbox
      --disk int32            Disk space allocated to the sandbox in GB
  -f, --dockerfile string     Path to Dockerfile for Sandbox image
  -e, --env stringArray       Environment variables (format: KEY=VALUE)
      --gpu int32             GPU units allocated to the sandbox
      --image string          Image to use for the sandbox
  -l, --label stringArray     Labels (format: KEY=VALUE)
      --memory int32          Memory allocated to the sandbox in MB
      --public                Make sandbox publicly accessible
      --target string         Target region (eu, us, asia)
      --user string           User associated with the sandbox
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

- [daytona sandbox](daytona_sandbox.md) - Manage Daytona sandboxes
