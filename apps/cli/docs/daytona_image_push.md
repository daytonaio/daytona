## daytona image push

Push local image

### Synopsis

Push local image or build and push from Dockerfile. If building locally, the image will be built with an AMD architecture.

```
daytona image push [IMAGE] [flags]
```

### Options

```
  -c, --context string      Build context directory (defaults to Dockerfile directory)
  -f, --dockerfile string   Path to Dockerfile to build before pushing
  -e, --entrypoint string   The entrypoint command for the image
```

### Options inherited from parent commands

```
      --help   help for daytona
```

### SEE ALSO

- [daytona image](daytona_image.md) - Manage Daytona images
