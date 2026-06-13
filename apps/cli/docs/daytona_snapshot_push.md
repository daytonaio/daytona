## daytona snapshot push

Push local snapshot

### Synopsis

Push a local Docker image to Daytona. To securely build it on our infrastructure, use 'daytona snapshot build'

```
daytona snapshot push [SNAPSHOT] [flags]
```

### Examples

```
  daytona snapshot push my-image:latest --name my-snapshot:1.0
  daytona snapshot push my-image:latest --name my-snapshot:1.0 --cpu 2 --memory 4
```

### Options

```
      --cpu int32           CPU cores that will be allocated to the underlying sandboxes (default: 1)
      --disk int32          Disk space that will be allocated to the underlying sandboxes in GB (default: 3)
  -e, --entrypoint string   The entrypoint command for the image
      --memory int32        Memory that will be allocated to the underlying sandboxes in GB (default: 1)
  -n, --name string         Specify the Snapshot name
      --region string       ID of the region where the snapshot will be available (defaults to organization default region)
      --timeout duration    Maximum time to wait for the snapshot to be validated (0 means wait indefinitely)
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
