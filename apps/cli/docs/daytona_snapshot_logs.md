## daytona snapshot logs

View the build logs of a snapshot

### Synopsis

View the build logs of a snapshot. With --follow the logs are streamed until the snapshot build reaches a terminal state.

```
daytona snapshot logs [SNAPSHOT_ID | SNAPSHOT_NAME] [flags]
```

### Examples

```
  daytona snapshot logs my-snapshot:1.0
  daytona snapshot logs my-snapshot:1.0 --follow
```

### Options

```
  -f, --follow   Follow the logs until the snapshot build reaches a terminal state
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
