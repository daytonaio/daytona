## daytona snapshot create

Create a snapshot

```
daytona snapshot create [SNAPSHOT] [flags]
```

### Examples

```
  daytona snapshot create my-snapshot:1.0 --image ubuntu:22.04 --entrypoint "sleep infinity"
  daytona snapshot create my-snapshot:1.0 --dockerfile ./Dockerfile --context ./app
  daytona snapshot create my-snapshot:1.0 --image ubuntu:22.04 --cpu 2 --memory 4 --disk 10
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
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
