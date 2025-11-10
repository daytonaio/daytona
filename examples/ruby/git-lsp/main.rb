# frozen_string_literal: true

daytona = Daytona::Daytona.new
sandbox = daytona.create(
  Daytona::CreateSandboxFromImageParams.new(
    image: Daytona::Image.base('ubuntu:25.10').run_commands(
      'apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils',
      'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
      'apt-get install -y nodejs',
      'npm install -g ts-node typescript typescript-language-server'
    ),
    timeout: 200
  )
)

project_dir = 'learn-typescript'

# Clone the repository
sandbox.git.clone(
  url: 'https://github.com/panaverse/learn-typescript',
  path: project_dir,
  branch: 'master'
)

sandbox.git.pull(path: project_dir)

# Search for the file we want to work on
matches = sandbox.fs.find_files(project_dir, 'var obj1 = new Base();')
puts "Matches: #{matches}"

# Start the language server
lsp = sandbox.create_lsp_server(language_id: Daytona::LspServer::Language::TYPESCRIPT, path_to_project: project_dir)
lsp.start

# Notify the language server of the document we want to work on
lsp.did_open(matches.first.file)

# Get symbols in the document
symbols = lsp.document_symbols(matches.first.file)
puts "Symbols: #{symbols}"

# Fix the error in the document
sandbox.fs.replace_in_files(
  files: [matches.first.file],
  pattern: 'var obj1 = new Base();',
  new_value: 'var obj1 = new E();'
)

# Notify the language server of the document change
lsp.did_close(matches.first.file)
lsp.did_open(matches.first.file)

# Get completions at a specific position
completions = lsp.completions(
  path: matches.first.file,
  position: Daytona::LspServer::Position.new(line: 12, character: 18)
)
print("Completions: #{completions}")

daytona.delete(sandbox)
