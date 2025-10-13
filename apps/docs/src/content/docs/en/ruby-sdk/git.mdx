---
title: "Git"
hideTitleOnPage: true
---

## Git

Main class for a new Git handler instance.

### Constructors

#### new Git()

```ruby
def initialize(sandbox_id:, toolbox_api:, otel_state:)

```

Initializes a new Git handler instance.

**Parameters**:

- `sandbox_id` _String_ - The Sandbox ID.
- `toolbox_api` _DaytonaToolboxApiClient:GitApi_ - API client for Sandbox operations.
- `otel_state` _Daytona:OtelState, nil_ -

**Returns**:

- `Git` - a new instance of Git

### Methods

#### sandbox_id()

```ruby
def sandbox_id()

```

**Returns**:

- `String` - The Sandbox ID

#### toolbox_api()

```ruby
def toolbox_api()

```

**Returns**:

- `DaytonaToolboxApiClient:GitApi` - API client for Sandbox operations

#### add()

```ruby
def add(path, files)

```

Stages the specified files for the next commit, similar to
running 'git add' on the command line.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `files` _Array\<String\>_ - List of file paths or directories to stage, relative to the repository root.

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if adding files fails

**Examples:**

```ruby
# Stage a single file
sandbox.git.add("workspace/repo", ["file.txt"])

# Stage multiple files
sandbox.git.add("workspace/repo", [
  "src/main.rb",
  "spec/main_spec.rb",
  "README.md"
])

```

#### branches()

```ruby
def branches(path)

```

Lists branches in the repository.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.

**Returns**:

- `DaytonaApiClient:ListBranchResponse` - List of branches in the repository.

**Raises**:

- `Daytona:Sdk:Error` - if listing branches fails

**Examples:**

```ruby
response = sandbox.git.branches("workspace/repo")
puts "Branches: #{response.branches}"

```

#### clone()

```ruby
def clone(url:, path:, branch:, commit_id:, username:, password:)

```

Clones a Git repository into the specified path. It supports
cloning specific branches or commits, and can authenticate with the remote
repository if credentials are provided.

**Parameters**:

- `url` _String_ - Repository URL to clone from.
- `path` _String_ - Path where the repository should be cloned. Relative paths are resolved
based on the sandbox working directory.
- `branch` _String, nil_ - Specific branch to clone. If not specified,
clones the default branch.
- `commit_id` _String, nil_ - Specific commit to clone. If specified,
the repository will be left in a detached HEAD state at this commit.
- `username` _String, nil_ - Git username for authentication.
- `password` _String, nil_ - Git password or token for authentication.

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if cloning repository fails

**Examples:**

```ruby
# Clone the default branch
sandbox.git.clone(
  url: "https://github.com/user/repo.git",
  path: "workspace/repo"
)

# Clone a specific branch with authentication
sandbox.git.clone(
  url: "https://github.com/user/private-repo.git",
  path: "workspace/private",
  branch: "develop",
  username: "user",
  password: "token"
)

# Clone a specific commit
sandbox.git.clone(
  url: "https://github.com/user/repo.git",
  path: "workspace/repo-old",
  commit_id: "abc123"
)

```

#### commit()

```ruby
def commit(path:, message:, author:, email:, allow_empty:)

```

Creates a new commit with the staged changes. Make sure to stage
changes using the add() method before committing.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `message` _String_ - Commit message describing the changes.
- `author` _String_ - Name of the commit author.
- `email` _String_ - Email address of the commit author.
- `allow_empty` _Boolean_ - Allow creating an empty commit when no changes are staged. Defaults to false.

**Returns**:

- `GitCommitResponse` - Response containing the commit SHA.

**Raises**:

- `Daytona:Sdk:Error` - if committing changes fails

**Examples:**

```ruby
# Stage and commit changes
sandbox.git.add("workspace/repo", ["README.md"])
commit_response = sandbox.git.commit(
  path: "workspace/repo",
  message: "Update documentation",
  author: "John Doe",
  email: "john@example.com",
  allow_empty: true
)
puts "Commit SHA: #{commit_response.sha}"

```

#### push()

```ruby
def push(path:, username:, password:)

```

Pushes all local commits on the current branch to the remote
repository. If the remote repository requires authentication, provide
username and password/token.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `username` _String, nil_ - Git username for authentication.
- `password` _String, nil_ - Git password or token for authentication.

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if pushing changes fails

**Examples:**

```ruby
# Push without authentication (for public repos or SSH)
sandbox.git.push("workspace/repo")

# Push with authentication
sandbox.git.push(
  path: "workspace/repo",
  username: "user",
  password: "github_token"
)

```

#### pull()

```ruby
def pull(path:, username:, password:)

```

Pulls changes from the remote repository. If the remote repository requires authentication,
provide username and password/token.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `username` _String, nil_ - Git username for authentication.
- `password` _String, nil_ - Git password or token for authentication.

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if pulling changes fails

**Examples:**

```ruby
# Pull without authentication
sandbox.git.pull("workspace/repo")

# Pull with authentication
sandbox.git.pull(
  path: "workspace/repo",
  username: "user",
  password: "github_token"
)

```

#### status()

```ruby
def status(path)

```

Gets the current Git repository status.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.

**Returns**:

- `DaytonaToolboxApiClient:GitStatus` - Repository status information including:

**Raises**:

- `Daytona:Sdk:Error` - if getting status fails

**Examples:**

```ruby
status = sandbox.git.status("workspace/repo")
puts "On branch: #{status.current_branch}"
puts "Commits ahead: #{status.ahead}"
puts "Commits behind: #{status.behind}"

```

#### checkout_branch()

```ruby
def checkout_branch(path, branch)

```

Checkout branch in the repository.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `branch` _String_ - Name of the branch to checkout

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if checking out branch fails

**Examples:**

```ruby
# Checkout a branch
sandbox.git.checkout_branch("workspace/repo", "feature-branch")

```

#### create_branch()

```ruby
def create_branch(path, name)

```

Create branch in the repository.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `name` _String_ - Name of the new branch to create

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if creating branch fails

**Examples:**

```ruby
# Create a new branch
sandbox.git.create_branch("workspace/repo", "new-feature")

```

#### delete_branch()

```ruby
def delete_branch(path, name)

```

Delete branch in the repository.

**Parameters**:

- `path` _String_ - Path to the Git repository root. Relative paths are resolved based on
the sandbox working directory.
- `name` _String_ - Name of the branch to delete

**Returns**:

- `void`

**Raises**:

- `Daytona:Sdk:Error` - if deleting branch fails

**Examples:**

```ruby
# Delete a branch
sandbox.git.delete_branch("workspace/repo", "old-feature")

```
