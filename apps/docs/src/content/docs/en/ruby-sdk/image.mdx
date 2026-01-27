---
title: "Image"
hideTitleOnPage: true
---

## Image

Represents an image definition for a Daytona sandbox.
Do not construct this class directly. Instead use one of its static factory methods,
such as `Image.base()`, `Image.debian_slim()`, or `Image.from_dockerfile()`.

### Constructors

#### new Image()

```ruby
def initialize(dockerfile:, context_list:)

```

**Parameters**:

- `dockerfile` _String, nil_ - The Dockerfile content
- `context_list` _Array\<Context\>_ - List of context files

**Returns**:

- `Image` - a new instance of Image

### Methods

#### dockerfile()

```ruby
def dockerfile()

```

**Returns**:

- `String, nil` - The generated Dockerfile for the image

#### context_list()

```ruby
def context_list()

```

**Returns**:

- `Array\<Context\>` - List of context files for the image

#### pip_install()

```ruby
def pip_install(*packages, find_links:, index_url:, extra_index_urls:, pre:, extra_options:)

```

Adds commands to install packages using pip

**Parameters**:

- `packages` _Array\<String\>_ - The packages to install
- `find_links` _Array\<String\>, nil_ - The find-links to use
- `index_url` _String, nil_ - The index URL to use
- `extra_index_urls` _Array\<String\>, nil_ - The extra index URLs to use
- `pre` _Boolean_ - Whether to install pre-release packages
- `extra_options` _String_ - Additional options to pass to pip

**Returns**:

- `Image` - The image with the pip install commands added

**Examples:**

```ruby
image = Image.debian_slim("3.12").pip_install("requests", "pandas")

```

#### pip_install_from_requirements()

```ruby
def pip_install_from_requirements(requirements_txt, find_links:, index_url:, extra_index_urls:, pre:, extra_options:)

```

Installs dependencies from a requirements.txt file

**Parameters**:

- `requirements_txt` _String_ - The path to the requirements.txt file
- `find_links` _Array\<String\>, nil_ - The find-links to use
- `index_url` _String, nil_ - The index URL to use
- `extra_index_urls` _Array\<String\>, nil_ - The extra index URLs to use
- `pre` _Boolean_ - Whether to install pre-release packages
- `extra_options` _String_ - Additional options to pass to pip

**Returns**:

- `Image` - The image with the pip install commands added

**Raises**:

- `Sdk:Error` - If the requirements file does not exist

**Examples:**

```ruby
image = Image.debian_slim("3.12").pip_install_from_requirements("requirements.txt")

```

#### pip_install_from_pyproject()

```ruby
def pip_install_from_pyproject(pyproject_toml, optional_dependencies:, find_links:, index_url:, extra_index_url:, pre:, extra_options:)

```

Installs dependencies from a pyproject.toml file

**Parameters**:

- `pyproject_toml` _String_ - The path to the pyproject.toml file
- `optional_dependencies` _Array\<String\>_ - The optional dependencies to install
- `find_links` _String, nil_ - The find-links to use
- `index_url` _String, nil_ - The index URL to use
- `extra_index_url` _String, nil_ - The extra index URL to use
- `pre` _Boolean_ - Whether to install pre-release packages
- `extra_options` _String_ - Additional options to pass to pip

**Returns**:

- `Image` - The image with the pip install commands added

**Raises**:

- `Sdk:Error` - If pyproject.toml parsing is not supported

**Examples:**

```ruby
image = Image.debian_slim("3.12").pip_install_from_pyproject("pyproject.toml", optional_dependencies: ["dev"])

```

#### add_local_file()

```ruby
def add_local_file(local_path, remote_path)

```

Adds a local file to the image

**Parameters**:

- `local_path` _String_ - The path to the local file
- `remote_path` _String_ - The path to the file in the image

**Returns**:

- `Image` - The image with the local file added

**Examples:**

```ruby
image = Image.debian_slim("3.12").add_local_file("package.json", "/home/daytona/package.json")

```

#### add_local_dir()

```ruby
def add_local_dir(local_path, remote_path)

```

Adds a local directory to the image

**Parameters**:

- `local_path` _String_ - The path to the local directory
- `remote_path` _String_ - The path to the directory in the image

**Returns**:

- `Image` - The image with the local directory added

**Examples:**

```ruby
image = Image.debian_slim("3.12").add_local_dir("src", "/home/daytona/src")

```

#### run_commands()

```ruby
def run_commands(*commands)

```

Runs commands in the image

**Parameters**:

- `commands` _Array\<String\>_ - The commands to run

**Returns**:

- `Image` - The image with the commands added

**Examples:**

```ruby
image = Image.debian_slim("3.12").run_commands('echo "Hello, world!"', 'echo "Hello again!"')

```

#### env()

```ruby
def env(env_vars)

```

Sets environment variables in the image

**Parameters**:

- `env_vars` _Hash\<String, String\>_ - The environment variables to set

**Returns**:

- `Image` - The image with the environment variables added

**Raises**:

- `Sdk:Error` -

**Examples:**

```ruby
image = Image.debian_slim("3.12").env({"PROJECT_ROOT" => "/home/daytona"})

```

#### workdir()

```ruby
def workdir(path)

```

Sets the working directory in the image

**Parameters**:

- `path` _String_ - The path to the working directory

**Returns**:

- `Image` - The image with the working directory added

**Examples:**

```ruby
image = Image.debian_slim("3.12").workdir("/home/daytona")

```

#### entrypoint()

```ruby
def entrypoint(entrypoint_commands)

```

Sets the entrypoint for the image

**Parameters**:

- `entrypoint_commands` _Array\<String\>_ - The commands to set as the entrypoint

**Returns**:

- `Image` - The image with the entrypoint added

**Examples:**

```ruby
image = Image.debian_slim("3.12").entrypoint(["/bin/bash"])

```

#### cmd()

```ruby
def cmd(cmd)

```

Sets the default command for the image

**Parameters**:

- `cmd` _Array\<String\>_ - The commands to set as the default command

**Returns**:

- `Image` - The image with the default command added

**Examples:**

```ruby
image = Image.debian_slim("3.12").cmd(["/bin/bash"])

```

#### dockerfile_commands()

```ruby
def dockerfile_commands(dockerfile_commands, context_dir:)

```

Adds arbitrary Dockerfile-like commands to the image

**Parameters**:

- `dockerfile_commands` _Array\<String\>_ - The commands to add to the Dockerfile
- `context_dir` _String, nil_ - The path to the context directory

**Returns**:

- `Image` - The image with the Dockerfile commands added

**Examples:**

```ruby
image = Image.debian_slim("3.12").dockerfile_commands(["RUN echo 'Hello, world!'"])

```
