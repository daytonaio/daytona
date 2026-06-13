## daytona

Daytona CLI

### Synopsis

Command line interface for Daytona Sandboxes

Exit codes: 0 success; 1 runtime failure; 2 invalid flags or arguments (where validated); 124 wait timeout. 'daytona exec' exits with the remote command's exit code; 255 indicates a CLI-side failure.

```
daytona [flags]
```

### Options

```
      --help       help for daytona
      --no-input   Never prompt for input; fail instead when input would be required
  -v, --version    Display the version of Daytona
```

### SEE ALSO

* [daytona api](daytona_api.md)  - Make an authenticated request to the Daytona API
* [daytona archive](daytona_archive.md)  - Archive a sandbox
* [daytona autocomplete](daytona_autocomplete.md)  - Adds a completion script for your shell environment
* [daytona cp](daytona_cp.md)  - Copy files between the local machine and a sandbox
* [daytona create](daytona_create.md)  - Create a new sandbox
* [daytona delete](daytona_delete.md)  - Delete a sandbox
* [daytona docs](daytona_docs.md)  - Opens the Daytona documentation in your default browser.
* [daytona exec](daytona_exec.md)  - Execute a command in a sandbox
* [daytona info](daytona_info.md)  - Get sandbox info
* [daytona list](daytona_list.md)  - List sandboxes
* [daytona login](daytona_login.md)  - Log in to Daytona
* [daytona logout](daytona_logout.md)  - Logout from Daytona
* [daytona logs](daytona_logs.md)  - View the build logs of a sandbox
* [daytona mcp](daytona_mcp.md)  - Manage Daytona MCP Server
* [daytona organization](daytona_organization.md)  - Manage Daytona organizations
* [daytona preview-url](daytona_preview-url.md)  - Get signed preview URL for a sandbox port
* [daytona snapshot](daytona_snapshot.md)  - Manage Daytona snapshots
* [daytona ssh](daytona_ssh.md)  - SSH into a sandbox
* [daytona start](daytona_start.md)  - Start a sandbox
* [daytona stop](daytona_stop.md)  - Stop a sandbox
* [daytona version](daytona_version.md)  - Print the version number
* [daytona volume](daytona_volume.md)  - Manage Daytona volumes
