## daytona snapshot create

Create a snapshot

```
daytona snapshot create [SNAPSHOT] [flags]
```

### Options

```
  -c, --context stringArray   Files or directories to include in the build context (can be specified multiple times). If not provided, context will be automatically determined from COPY/ADD commands in the Dockerfile
      --cpu int32             CPU cores that will be allocated to the underlying sandboxes (default: 1)
      --disk int32            Disk space that will be allocated to the underlying sandboxes in GB (default: 3)
  -f, --dockerfile string     Path to Dockerfile to build
  -e, --entrypoint string     The entrypoint command for the snapshot
  -i, --image string          The image name for the snapshot
      --memory int32          Memory that will be allocated to the underlying sandboxes in GB (default: 1)
      --region string         ID of the region where the snapshot will be available (defaults to organization default region)
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
