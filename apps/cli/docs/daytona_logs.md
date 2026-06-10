## daytona logs

View the build logs of a sandbox

### Synopsis

View the build logs of a sandbox. With --follow the logs are streamed until the sandbox build reaches a terminal state.

```
daytona logs [SANDBOX_ID | SANDBOX_NAME] [flags]
```

### Examples

```
  daytona logs my-sandbox
  daytona logs my-sandbox --follow
```

### Options

```
  -f, --follow   Follow the logs until the sandbox build reaches a terminal state
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
