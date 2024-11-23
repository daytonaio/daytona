## daytona prebuild update

Update a prebuild configuration

```
daytona prebuild update [PROJECT_CONFIG] [PREBUILD_ID] [flags]
```

### Options

```
  -b, --branch string           Git branch for the prebuild
  -c, --commit-interval int     Commit interval for running a prebuild - leave blank to ignore push events
  -r, --retention int           Maximum number of resulting builds stored at a time
      --run                     Run the prebuild once after updating it
  -t, --trigger-files strings   Full paths of files whose changes should explicitly trigger a  prebuild
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

* [daytona prebuild](daytona_prebuild.md)	 - Manage prebuilds

