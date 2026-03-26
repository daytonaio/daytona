---
title: "Image"
hideTitleOnPage: true
---


## Image

Represents an image definition for a Daytona sandbox.
Do not construct this class directly. Instead use one of its static factory methods,
such as `Image.base()`, `Image.debianSlim()` or `Image.fromDockerfile()`.

### Accessors

#### contextList

##### Get Signature

```ts
get contextList(): Context[]
```

###### Returns

`Context`[]

The list of context files to be added to the image.

***

#### dockerfile

##### Get Signature

```ts
get dockerfile(): string
```

**Returns**:

- `string` - The Dockerfile content.

### Methods

#### base()

```ts
static base(image: string): Image
```

Creates an Image from an existing base image.

**Parameters**:

- `image` _string_ - The base image to use.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.base('python:3.12-slim-bookworm')
```

***

#### debianSlim()

```ts
static debianSlim(pythonVersion?: "3.9" | "3.10" | "3.11" | "3.12" | "3.13"): Image
```

Creates a Debian slim image based on the official Python Docker image.

**Parameters**:

- `pythonVersion?` _The Python version to use._ - `"3.9"` | `"3.10"` | `"3.11"` | `"3.12"` | `"3.13"`


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.debianSlim('3.12')
```

***

#### fromDockerfile()

```ts
static fromDockerfile(path: string): Image
```

Creates an Image from an existing Dockerfile.

**Parameters**:

- `path` _string_ - The path to the Dockerfile.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.fromDockerfile('Dockerfile')
```

***

#### addLocalDir()

```ts
addLocalDir(localPath: string, remotePath: string): Image
```

Adds a local directory to the image.

**Parameters**:

- `localPath` _string_ - The path to the local directory.
- `remotePath` _string_ - The path of the directory in the image.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .addLocalDir('src', '/home/daytona/src')
```

***

#### addLocalFile()

```ts
addLocalFile(localPath: string, remotePath: string): Image
```

Adds a local file to the image.

**Parameters**:

- `localPath` _string_ - The path to the local file.
- `remotePath` _string_ - The path of the file in the image.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .addLocalFile('requirements.txt', '/home/daytona/requirements.txt')
```

***

#### cmd()

```ts
cmd(cmd: string[]): Image
```

Sets the default command for the image.

**Parameters**:

- `cmd` _string\[\]_ - The command to set as the default command.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .cmd(['/bin/bash'])
```

***

#### dockerfileCommands()

```ts
dockerfileCommands(dockerfileCommands: string[], contextDir?: string): Image
```

Extends an image with arbitrary Dockerfile-like commands.

**Parameters**:

- `dockerfileCommands` _string\[\]_ - The commands to add to the Dockerfile.
- `contextDir?` _string_ - The path to the context directory.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .dockerfileCommands(['RUN echo "Hello, world!"'])
```

***

#### entrypoint()

```ts
entrypoint(entrypointCommands: string[]): Image
```

Sets the entrypoint for the image.

**Parameters**:

- `entrypointCommands` _string\[\]_ - The commands to set as the entrypoint.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .entrypoint(['/bin/bash'])
```

***

#### env()

```ts
env(envVars: Record<string, string>): Image
```

Sets environment variables in the image.

**Parameters**:

- `envVars` _Record\<string, string\>_ - The environment variables to set.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .env({ FOO: 'bar' })
```

***

#### pipInstall()

```ts
pipInstall(packages: string | string[], options?: PipInstallOptions): Image
```

Adds commands to install packages using pip.

**Parameters**:

- `packages` _The packages to install._ - `string` | `string`[]
- `options?` _PipInstallOptions_ - The options for the pip install command.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.debianSlim('3.12').pipInstall('numpy', { findLinks: ['https://pypi.org/simple'] })
```

***

#### pipInstallFromPyproject()

```ts
pipInstallFromPyproject(pyprojectToml: string, options?: PyprojectOptions): Image
```

Installs dependencies from a pyproject.toml file.

**Parameters**:

- `pyprojectToml` _string_ - The path to the pyproject.toml file.
- `options?` _PyprojectOptions_ - The options for the pip install command.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.debianSlim('3.12')
image.pipInstallFromPyproject('pyproject.toml', { optionalDependencies: ['dev'] })
```

***

#### pipInstallFromRequirements()

```ts
pipInstallFromRequirements(requirementsTxt: string, options?: PipInstallOptions): Image
```

Installs dependencies from a requirements.txt file.

**Parameters**:

- `requirementsTxt` _string_ - The path to the requirements.txt file.
- `options?` _PipInstallOptions_ - The options for the pip install command.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image.debianSlim('3.12')
image.pipInstallFromRequirements('requirements.txt', { findLinks: ['https://pypi.org/simple'] })
```

***

#### runCommands()

```ts
runCommands(...commands: (string | string[])[]): Image
```

Runs commands in the image.

**Parameters**:

- `commands` _...\(string \| string\[\]\)\[\]_ - The commands to run.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .runCommands(
   'echo "Hello, world!"',
   ['bash', '-c', 'echo Hello, world, again!']
 )
```

***

#### workdir()

```ts
workdir(dirPath: string): Image
```

Sets the working directory in the image.

**Parameters**:

- `dirPath` _string_ - The path to the working directory.


**Returns**:

- `Image` - The Image instance.

**Example:**

```ts
const image = Image
 .debianSlim('3.12')
 .workdir('/home/daytona')
```

***


## Context

Represents a context file to be added to the image.

**Properties**:

- `archivePath` _string_ - The path inside the archive file in object storage.
- `sourcePath` _string_ - The path to the source file or directory.
## PipInstallOptions

Options for the pip install command.

**Properties**:

- `extraIndexUrls?` _string\[\]_ - The extra index URLs to use for the pip install command.
- `extraOptions?` _string_ - The extra options to use for the pip install command. Given string is passed directly to the pip install command.
- `findLinks?` _string\[\]_ - The find-links to use for the pip install command.
- `indexUrl?` _string_ - The index URL to use for the pip install command.
- `pre?` _boolean_ - Whether to install pre-release versions.
    




### Extended by

- `PyprojectOptions`
## PyprojectOptions

Options for the pip install command from a pyproject.toml file.

**Properties**:

- `extraIndexUrls?` _string\[\]_ - The extra index URLs to use for the pip install command.
    - _Inherited from_: `PipInstallOptions.extraIndexUrls`
- `extraOptions?` _string_ - The extra options to use for the pip install command. Given string is passed directly to the pip install command.
    - _Inherited from_: `PipInstallOptions.extraOptions`
- `findLinks?` _string\[\]_ - The find-links to use for the pip install command.
    - _Inherited from_: `PipInstallOptions.findLinks`
- `indexUrl?` _string_ - The index URL to use for the pip install command.
    - _Inherited from_: `PipInstallOptions.indexUrl`
- `optionalDependencies?` _string\[\]_ - The optional dependencies to install.
- `pre?` _boolean_ - Whether to install pre-release versions.
    - _Inherited from_: `PipInstallOptions.pre`



**Extends:**

- `PipInstallOptions`