## daytona volume create

Create a volume

```
daytona volume create [NAME] [flags]
```

### Examples

```
  daytona volume create my-volume
  # Mount it when creating a sandbox
  daytona create --snapshot my-snapshot:1.0 --volume my-volume:/data
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona volume](daytona_volume.md)  - Manage Daytona volumes
