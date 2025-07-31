## daytona snapshot push

Push local snapshot

### Synopsis

Push a local Docker image to Daytona. To securely build it on our infrastructure, use 'daytona snapshot build'

```
daytona snapshot push [SNAPSHOT] [flags]
```

### Options

```
      --cpu int32           CPU cores that will be allocated to the underlying sandboxes (default: 1)
      --disk int32          Disk space that will be allocated to the underlying sandboxes in GB (default: 3)
  -e, --entrypoint string   The entrypoint command for the image
      --memory int32        Memory that will be allocated to the underlying sandboxes in GB (default: 1)
  -n, --name string         Specify the Snapshot name
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

- [daytona snapshot](daytona_snapshot.md) - Manage Daytona snapshots
