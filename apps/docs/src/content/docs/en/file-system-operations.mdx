---
title: File System Operations
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

Daytona provides comprehensive file system operations through the `fs` module in sandboxes.

## Basic operations

Daytona provides methods to interact with the file system in sandboxes. You can perform various operations like listing files, creating directories, reading and writing files, and more.

File operations assume you are operating in the sandbox user's home directory (e.g. `workspace` implies `/home/[username]/workspace`). Use a leading `/` when providing absolute paths.

### List files and directories

Daytona provides methods to list files and directories in a sandbox by providing the path to the directory. If the path is not provided, the method will list the files and directories in the sandbox working directory.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# List files in a directory
files = sandbox.fs.list_files("workspace")

for file in files:
    print(f"Name: {file.name}")
    print(f"Is directory: {file.is_dir}")
    print(f"Size: {file.size}")
    print(f"Modified: {file.mod_time}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// List files in a directory
const files = await sandbox.fs.listFiles('workspace')

files.forEach(file => {
  console.log(`Name: ${file.name}`)
  console.log(`Is directory: ${file.isDir}`)
  console.log(`Size: ${file.size}`)
  console.log(`Modified: ${file.modTime}`)
})
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# List directory contents
files = sandbox.fs.list_files("workspace/data")

# Print files and their sizes
files.each do |file|
  puts "#{file.name}: #{file.size} bytes" unless file.is_dir
end

# List only directories
dirs = files.select(&:is_dir)
puts "Subdirectories: #{dirs.map(&:name).join(', ')}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// List files in a directory
files, err := sandbox.FileSystem.ListFiles(ctx, "workspace")
if err != nil {
	log.Fatal(err)
}

for _, file := range files {
	fmt.Printf("Name: %s\n", file.Name)
	fmt.Printf("Is directory: %t\n", file.IsDirectory)
	fmt.Printf("Size: %d\n", file.Size)
	fmt.Printf("Modified: %s\n", file.ModifiedTime)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox) references:

> [**list_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemlist_files)
>
> [**listFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#listfiles)
>
> [**list_files (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#list_files)
>
> [**ListFiles (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.ListFiles)
>
> [**list files and directories (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/GET/files)

### Get directory or file information

Daytona provides methods to get directory or file information such as group, directory, modified time, mode, name, owner, permissions, and size by providing the path to the directory or file.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Get file metadata
info = sandbox.fs.get_file_info("workspace/data/file.txt")
print(f"Size: {info.size} bytes")
print(f"Modified: {info.mod_time}")
print(f"Mode: {info.mode}")

# Check if path is a directory
info = sandbox.fs.get_file_info("workspace/data")
if info.is_dir:
    print("Path is a directory")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Get file details
const info = await fs.getFileDetails('app/config.json')
console.log(`Size: ${info.size}, Modified: ${info.modTime}`)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Get file metadata
info = sandbox.fs.get_file_info("workspace/data/file.txt")
puts "Size: #{info.size} bytes"
puts "Modified: #{info.mod_time}"
puts "Mode: #{info.mode}"

# Check if path is a directory
info = sandbox.fs.get_file_info("workspace/data")
puts "Path is a directory" if info.is_dir
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Get file metadata
info, err := sandbox.FileSystem.GetFileInfo(ctx, "workspace/data/file.txt")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Size: %d bytes\n", info.Size)
fmt.Printf("Modified: %s\n", info.ModifiedTime)
fmt.Printf("Mode: %s\n", info.Mode)

// Check if path is a directory
info, err = sandbox.FileSystem.GetFileInfo(ctx, "workspace/data")
if err != nil {
	log.Fatal(err)
}
if info.IsDirectory {
	fmt.Println("Path is a directory")
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/info?path='
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**get_file_info (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemget_file_info)
>
> [**getFileDetails (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#getfiledetails)
>
> [**get_file_info (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#get_file_info)
>
> [**GetFileInfo (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.GetFileInfo)
>
> [**get file information (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/GET/files/info)

### Create directories

Daytona provides methods to create directories by providing the path to the directory and the permissions to set on the directory.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Create with specific permissions
sandbox.fs.create_folder("workspace/new-dir", "755")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Create with specific permissions
await sandbox.fs.createFolder('workspace/new-dir', '755')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Create a directory with standard permissions
sandbox.fs.create_folder("workspace/data", "755")

# Create a private directory
sandbox.fs.create_folder("workspace/secrets", "700")
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Create with specific permissions
err := sandbox.FileSystem.CreateFolder(ctx, "workspace/new-dir",
	options.WithMode("755"),
)
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/folder?path=&mode=' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**create_folder (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemcreate_folder)
>
> [**createFolder (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#createfolder)
>
> [**create_folder (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#create_folder)
>
> [**CreateFolder (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.CreateFolder)
>
> [**create folder (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/POST/files/folder)

### Upload files

Daytona provides methods to upload a single or multiple files in sandboxes.

#### Upload a single file

Daytona provides methods to upload a single file in sandboxes by providing the content to upload and the path to the file to upload it to.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Upload a single file
with open("local_file.txt", "rb") as f:
    content = f.read()
sandbox.fs.upload_file(content, "remote_file.txt")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Upload a single file
const fileContent = Buffer.from('Hello, World!')
await sandbox.fs.uploadFile(fileContent, 'data.txt')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Upload a text file from string content
content = "Hello, World!"
sandbox.fs.upload_file(content, "tmp/hello.txt")

# Upload a local file
sandbox.fs.upload_file("local_file.txt", "tmp/file.txt")

# Upload binary data
data = { key: "value" }.to_json
sandbox.fs.upload_file(data, "tmp/config.json")
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Upload from a local file path
err := sandbox.FileSystem.UploadFile(ctx, "local_file.txt", "remote_file.txt")
if err != nil {
	log.Fatal(err)
}

// Or upload from byte content
content := []byte("Hello, World!")
err = sandbox.FileSystem.UploadFile(ctx, content, "hello.txt")
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/upload?path=' \
  --request POST \
  --header 'Content-Type: multipart/form-data' \
  --form 'file='
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**upload_file (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemupload_file)
>
> [**uploadFile (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#uploadfile)
>
> [**upload_file (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#upload_file)
>
> [**UploadFile (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.UploadFile)
>
> [**upload file (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/POST/files/upload)

#### Upload multiple files

Daytona provides methods to upload multiple files in sandboxes by providing the content to upload and their destination paths.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Upload multiple files at once
files_to_upload = []

with open("file1.txt", "rb") as f1:
    files_to_upload.append(FileUpload(
        source=f1.read(),
        destination="data/file1.txt",
    ))

with open("file2.txt", "rb") as f2:
    files_to_upload.append(FileUpload(
        source=f2.read(),
        destination="data/file2.txt",
    ))

with open("settings.json", "rb") as f3:
    files_to_upload.append(FileUpload(
        source=f3.read(),
        destination="config/settings.json",
    ))

sandbox.fs.upload_files(files_to_upload)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Upload multiple files at once
const files = [
  {
    source: Buffer.from('Content of file 1'),
    destination: 'data/file1.txt',
  },
  {
    source: Buffer.from('Content of file 2'),
    destination: 'data/file2.txt',
  },
  {
    source: Buffer.from('{"key": "value"}'),
    destination: 'config/settings.json',
  },
]

await sandbox.fs.uploadFiles(files)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Upload multiple files
files = [
  FileUpload.new("Content of file 1", "/tmp/file1.txt"),
  FileUpload.new("workspace/data/file2.txt", "/tmp/file2.txt"),
  FileUpload.new('{"key": "value"}', "/tmp/config.json")
]

sandbox.fs.upload_files(files)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Upload multiple files by calling UploadFile for each
filesToUpload := []struct {
	source      string
	destination string
}{
	{"file1.txt", "data/file1.txt"},
	{"file2.txt", "data/file2.txt"},
	{"settings.json", "config/settings.json"},
}

for _, f := range filesToUpload {
	err := sandbox.FileSystem.UploadFile(ctx, f.source, f.destination)
	if err != nil {
		log.Fatal(err)
	}
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/bulk-upload' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**upload_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemupload_files)
>
> [**uploadFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#uploadfiles)
>
> [**upload_files (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#upload_files)
>
> [**upload multiple files (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/POST/files/bulk-upload)

### Download files

Daytona provides methods to download files from sandboxes.

#### Download a single file

Daytona provides methods to download a single file from sandboxes by providing the path to the file to download.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
content = sandbox.fs.download_file("file1.txt")

with open("local_file.txt", "wb") as f:
    f.write(content)

print(content.decode('utf-8'))
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const downloadedFile = await sandbox.fs.downloadFile('file1.txt')
console.log('File content:', downloadedFile.toString())
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Download and get file content
content = sandbox.fs.download_file("workspace/data/file.txt")
puts content

# Download and save a file locally
sandbox.fs.download_file("workspace/data/file.txt", "local_copy.txt")
size_mb = File.size("local_copy.txt") / 1024.0 / 1024.0
puts "Size of the downloaded file: #{size_mb} MB"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Download and get contents in memory
content, err := sandbox.FileSystem.DownloadFile(ctx, "file1.txt", nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(content))

// Download and save to a local file
localPath := "local_file.txt"
content, err = sandbox.FileSystem.DownloadFile(ctx, "file1.txt", &localPath)
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/download?path='
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**download_file (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemdownload_file)
>
> [**downloadFile (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#downloadfile)
>
> [**download_file (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#download_file)
>
> [**DownloadFile (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.DownloadFile)
>
> [**download file (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/GET/files/download)

#### Download multiple files

Daytona provides methods to download multiple files from sandboxes by providing the paths to the files to download.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Download multiple files at once
files_to_download = [
    FileDownloadRequest(source="data/file1.txt"), # No destination - download to memory
    FileDownloadRequest(source="data/file2.txt", destination="local_file2.txt"), # Download to local file
]

results = sandbox.fs.download_files(files_to_download)

for result in results:
    if result.error:
        print(f"Error downloading {result.source}: {result.error}")
    elif result.result:
        print(f"Downloaded {result.source} to {result.result}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Download multiple files at once
const files = [
  { source: 'data/file1.txt' }, // No destination - download to memory
  { source: 'data/file2.txt', destination: 'local_file2.txt' }, // Download to local file
]

const results = await sandbox.fs.downloadFiles(files)

results.forEach(result => {
  if (result.error) {
    console.error(`Error downloading ${result.source}: ${result.error}`)
  } else if (result.result) {
    console.log(`Downloaded ${result.source} to ${result.result}`)
  }
})
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Download multiple files by calling download_file for each
files_to_download = [
  { remote: "data/file1.txt", local: nil },              # Download to memory
  { remote: "data/file2.txt", local: "local_file2.txt" } # Download to local file
]

files_to_download.each do |f|
  if f[:local]
    sandbox.fs.download_file(f[:remote], f[:local])
    puts "Downloaded #{f[:remote]} to #{f[:local]}"
  else
    content = sandbox.fs.download_file(f[:remote])
    puts "Downloaded #{f[:remote]} to memory (#{content.size} bytes)"
  end
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Download multiple files by calling DownloadFile for each
filesToDownload := []struct {
	remotePath string
	localPath  *string
}{
	{"data/file1.txt", nil},                           // Download to memory
	{"data/file2.txt", ptrString("local_file2.txt")},  // Download to local file
}

for _, f := range filesToDownload {
	content, err := sandbox.FileSystem.DownloadFile(ctx, f.remotePath, f.localPath)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", f.remotePath, err)
		continue
	}
	if f.localPath == nil {
		fmt.Printf("Downloaded %s to memory (%d bytes)\n", f.remotePath, len(content))
	} else {
		fmt.Printf("Downloaded %s to %s\n", f.remotePath, *f.localPath)
	}
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/bulk-download' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "paths": [
    ""
  ]
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**download_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemdownload_files)
>
> [**downloadFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#downloadfiles)
>
> [**download_file (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#download_file)
>
> [**download multiple files (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/POST/files/bulk-download)

### Delete files

Daytona provides methods to delete files or directories from sandboxes by providing the path to the file or directory to delete.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
sandbox.fs.delete_file("workspace/file.txt")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
await sandbox.fs.deleteFile('workspace/file.txt')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Delete a file
sandbox.fs.delete_file("workspace/data/old_file.txt")

# Delete a directory recursively
sandbox.fs.delete_file("workspace/old_dir", recursive: true)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Delete a file
err := sandbox.FileSystem.DeleteFile(ctx, "workspace/file.txt", false)
if err != nil {
	log.Fatal(err)
}

// Delete a directory recursively
err = sandbox.FileSystem.DeleteFile(ctx, "workspace/old_dir", true)
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files?path=' \
  --request DELETE
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**delete_file (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemdelete_file)
>
> [**deleteFile (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#deletefile)
>
> [**delete_file (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#delete_file)
>
> [**DeleteFile (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.DeleteFile)
>
> [**delete file or directory (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/DELETE/files)

## Advanced operations

Daytona provides advanced file system operations such as file permissions, search and replace, and move files.

### File permissions

Daytona provides methods to set file permissions, ownership, and group for a file or directory by providing the path to the file or directory and the permissions to set.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Set file permissions
sandbox.fs.set_file_permissions("workspace/file.txt", "644")

# Get file permissions
file_info = sandbox.fs.get_file_info("workspace/file.txt")
print(f"Permissions: {file_info.permissions}")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Set file permissions
await sandbox.fs.setFilePermissions('workspace/file.txt', { mode: '644' })

// Get file permissions
const fileInfo = await sandbox.fs.getFileDetails('workspace/file.txt')
console.log(`Permissions: ${fileInfo.permissions}`)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Make a file executable
sandbox.fs.set_file_permissions(
  path: "workspace/scripts/run.sh",
  mode: "755"  # rwxr-xr-x
)

# Change file owner
sandbox.fs.set_file_permissions(
  path: "workspace/data/file.txt",
  owner: "daytona",
  group: "daytona"
)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Set file permissions
err := sandbox.FileSystem.SetFilePermissions(ctx, "workspace/file.txt",
	options.WithPermissionMode("644"),
)
if err != nil {
	log.Fatal(err)
}

// Set owner and group
err = sandbox.FileSystem.SetFilePermissions(ctx, "workspace/file.txt",
	options.WithOwner("daytona"),
	options.WithGroup("daytona"),
)
if err != nil {
	log.Fatal(err)
}

// Get file info to check permissions
fileInfo, err := sandbox.FileSystem.GetFileInfo(ctx, "workspace/file.txt")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Mode: %s\n", fileInfo.Mode)
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/permissions?path=' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**set_file_permissions (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemset_file_permissions)
>
> [**setFilePermissions (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#setfilepermissions)
>
> [**set_file_permissions (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#set_file_permissions)
>
> [**SetFilePermissions (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.SetFilePermissions)
>
> [**set file permissions (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/POST/files/permissions)

### Find and replace text in files

Daytona provides methods to find and replace text in files by providing the path to the directory to search in and the pattern to search for.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Search for text in files by providing the path to the directory to search in and the pattern to search for
results = sandbox.fs.find_files(
    path="workspace/src",
    pattern="text-of-interest"
)
for match in results:
    print(f"Absolute file path: {match.file}")
    print(f"Line number: {match.line}")
    print(f"Line content: {match.content}")
    print("\n")

# Replace text in files
sandbox.fs.replace_in_files(
    files=["workspace/file1.txt", "workspace/file2.txt"],
    pattern="old_text",
    new_value="new_text"
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Search for text in files; if a folder is specified, the search is recursive
const results = await sandbox.fs.findFiles({
    path="workspace/src",
    pattern: "text-of-interest"
})
results.forEach(match => {
    console.log('Absolute file path:', match.file)
    console.log('Line number:', match.line)
    console.log('Line content:', match.content)
})

// Replace text in files
await sandbox.fs.replaceInFiles(
    ["workspace/file1.txt", "workspace/file2.txt"],
    "old_text",
    "new_text"
)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Search for TODOs in Ruby files
matches = sandbox.fs.find_files("workspace/src", "TODO:")
matches.each do |match|
  puts "#{match.file}:#{match.line}: #{match.content.strip}"
end

# Replace in specific files
results = sandbox.fs.replace_in_files(
  files: ["workspace/src/file1.rb", "workspace/src/file2.rb"],
  pattern: "old_function",
  new_value: "new_function"
)

# Print results
results.each do |result|
  if result.success
    puts "#{result.file}: #{result.success}"
  else
    puts "#{result.file}: #{result.error}"
  end
end
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Search for text in files
result, err := sandbox.FileSystem.FindFiles(ctx, "workspace/src", "text-of-interest")
if err != nil {
	log.Fatal(err)
}
matches := result.([]map[string]any)
for _, match := range matches {
	fmt.Printf("Absolute file path: %s\n", match["file"])
	fmt.Printf("Line number: %v\n", match["line"])
	fmt.Printf("Line content: %s\n\n", match["content"])
}

// Replace text in files
_, err = sandbox.FileSystem.ReplaceInFiles(ctx,
	[]string{"workspace/file1.txt", "workspace/file2.txt"},
	"old_text",
	"new_text",
)
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

Find text in files:

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/find?path=&pattern='
```

Replace text in files:

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/replace' \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "files": [
    ""
  ],
  "newValue": "",
  "pattern": ""
}'
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/) references:

> [**find_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemfind_files)
>
> [**replace_in_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemreplace_in_files)
>
> [**findFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#findfiles)
>
> [**replaceInFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#replaceinfiles)
>
> [**find_files (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#find_files)
>
> [**replace_in_files (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#replace_in_files)
>
> [**FindFiles (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.FindFiles)
>
> [**ReplaceInFiles (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.ReplaceInFiles)
>
> [**find text in files (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/GET/files/find)
>
> [**replace text in files (API)**](/docs/en/tools/api#daytona-toolbox/tag/file-system/POST/files/replace)

### Move or rename directory or file

Daytona provides methods to move or rename a directory or file in sandboxes by providing the path to the file or directory (source) and the new path to the file or directory (destination).

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Rename a file
sandbox.fs.move_files(
    "workspace/data/old_name.txt",
    "workspace/data/new_name.txt"
)

# Move a file to a different directory
sandbox.fs.move_files(
    "workspace/data/file.txt",
    "workspace/archive/file.txt"
)

# Move a directory
sandbox.fs.move_files(
    "workspace/old_dir",
    "workspace/new_dir"
)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Move a file to a new location
await fs.moveFiles('app/temp/data.json', 'app/data/data.json')
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
# Rename a file
sandbox.fs.move_files(
  "workspace/data/old_name.txt",
  "workspace/data/new_name.txt"
)

# Move a file to a different directory
sandbox.fs.move_files(
  "workspace/data/file.txt",
  "workspace/archive/file.txt"
)

# Move a directory
sandbox.fs.move_files(
  "workspace/old_dir",
  "workspace/new_dir"
)
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
// Rename a file
err := sandbox.FileSystem.MoveFiles(ctx, "workspace/data/old_name.txt", "workspace/data/new_name.txt")
if err != nil {
	log.Fatal(err)
}

// Move a file to a different directory
err = sandbox.FileSystem.MoveFiles(ctx, "workspace/data/file.txt", "workspace/archive/file.txt")
if err != nil {
	log.Fatal(err)
}

// Move a directory
err = sandbox.FileSystem.MoveFiles(ctx, "workspace/old_dir", "workspace/new_dir")
if err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/files/move?source=&destination=' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), [Ruby SDK](/docs/en/ruby-sdk/), [Go SDK](/docs/en/go-sdk/), and [API](/docs/en/tools/api/#daytona-toolbox) references:

> [**move_files (Python SDK)**](/docs/en/python-sdk/sync/file-system/#filesystemmove_files)
>
> [**moveFiles (TypeScript SDK)**](/docs/en/typescript-sdk/file-system/#movefiles)
>
> [**move_files (Ruby SDK)**](/docs/en/ruby-sdk/file-system/#move_files)
>
> [**MoveFiles (Go SDK)**](/docs/en/go-sdk/daytona/#FileSystemService.MoveFiles)
>
> [**move or rename file or directory (API)**](/docs/en/tools/api/#daytona-toolbox/tag/file-system/POST/files/move)
