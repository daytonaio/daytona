## daytona prebuild update

Update a prebuild configuration

```
daytona prebuild update [PROJECT_CONFIG] [PREBUILD_ID] [flags]
```

### Options

```
  -b, --branch string           Git branch for the prebuild
  -c, --commit-interval int     Commit interval for the prebuild
  -r, --retention int           Retention period for the prebuild
      --run                     Run the prebuild once after updating it
  -t, --trigger-files strings   Files that trigger the prebuild
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona prebuild](daytona_prebuild.md)	 - Manage prebuilds

