## daytona api

Make an authenticated request to the Daytona API

### Synopsis

Make an authenticated HTTP request to the Daytona API and print the raw response body to stdout.

PATH is resolved against the active profile's API URL and the request is authenticated with the active profile's credentials. Responses with status 400 or above still print the body, then exit non-zero.

```
daytona api PATH [flags]
```

### Examples

```
  daytona api /sandbox
  daytona api /sandbox/my-sandbox -X DELETE
  daytona api /snapshots -X POST --input snapshot.json
  cat body.json | daytona api /sandbox -X POST --input -
```

### Options

```
      --input string    Request body source: a file path or '-' for stdin (POST, PUT, and PATCH only)
  -X, --method string   HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD) (default "GET")
```

### Options inherited from parent commands

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
```

### SEE ALSO

* [daytona](daytona.md)  - Daytona CLI
