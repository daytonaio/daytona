# frozen_string_literal: true

require 'json'

daytona = Daytona::Daytona.new
params = Daytona::CreateSandboxFromSnapshotParams.new(language: Daytona::CodeLanguage::PYTHON)

# Create a Sandbox
sandbox = daytona.create(params)
puts "Created sandbox ##{sandbox.id}"

# List files in the Sandbox
files = sandbox.fs.list_files('.')
puts "Initial files: #{files}"

# Create a new directory directory in the Sandbox
project_files = 'project-files'
sandbox.fs.create_folder(project_files, '755')

# Create local file for demonstration
local_file_path = 'local-example.txt'
File.write(local_file_path, 'This is a local file created for demonstration purposes')

# Create a configuration file with JSON dadta
config_data = JSON.dump(name: 'project-config', version: '1.0.0', settings: { debug: true, max_connections: 10 })

# Upload multiple files at once - both from local path and from bytes
script = <<~BASH
  #!/bin/bash
  echo "Hello from script!"
  exit 0
BASH
sandbox.fs.upload_files(
  [Daytona::FileUpload.new(local_file_path, File.join(project_files, 'example.txt')),
   Daytona::FileUpload.new(config_data, File.join(project_files, 'config.json')),
   Daytona::FileUpload.new(script, File.join(project_files, 'script.sh'))]
)

# Execute commands on the sandbox to verify files and make them executable
ls_cmd = sandbox.process.exec(command: "ls -la #{project_files}")
puts ls_cmd.result

# Make the script executable
sandbox.process.exec(command: "chmod +x #{File.join(project_files, 'script.sh')}")

# Run the script
run_cmd = sandbox.process.exec(command: "./#{File.join(project_files, 'script.sh')}")
puts run_cmd.result

# Search for files in the project
matches = sandbox.fs.search_files(project_files, '*.json')
puts "JSON files found: #{matches}"

# Download from remote and save it locally
sandbox.fs.download_file(File.join(project_files, 'config.json'), 'local-config.json')
file = File.new('local-config.json')
puts "Content of local-config.json: #{file.read}"
puts "Size of the downloaded file: #{file.size} bytes"

# Download from remote and get the reference to temporary file
file = sandbox.fs.download_file(File.join(project_files, 'example.txt'))
puts "Content of example.txt: #{file.open.read}"
puts "Size of the downloaded file: #{file.size} bytes"

# Cleanup
File.delete('local-config.json') if File.exist?('local-config.json')
File.delete('example.txt') if File.exist?('example.txt')
daytona.delete(sandbox)
