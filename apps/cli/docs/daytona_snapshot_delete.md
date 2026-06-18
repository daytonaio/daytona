## daytona snapshot delete

Delete a snapshot

```
daytona snapshot delete [SNAPSHOT_ID | SNAPSHOT_NAME] [flags]
```

### Examples

```
  daytona snapshot delete my-snapshot:1.0
  daytona snapshot delete --all --dry-run
  daytona snapshot delete --all --yes --format json
```

### Options

```
  -a, --all                Delete all snapshots
      --dry-run            Show what would be deleted without deleting anything
  -f, --format string      Output format. Must be one of (yaml, json)
      --ignore-not-found   Treat a missing snapshot as a successful delete
  -y, --yes                Skip the confirmation prompt for bulk deletes
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
