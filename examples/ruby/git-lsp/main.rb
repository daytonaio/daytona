# frozen_string_literal: true

require 'daytona'

def section(title)
  puts "\n=== #{title} ==="
end

daytona = Daytona::Daytona.new
# Custom image with a TypeScript language server (for the LSP showcase) and git.
sandbox = daytona.create(
  Daytona::CreateSandboxFromImageParams.new(
    image: Daytona::Image.base('ubuntu:25.10').run_commands(
      'apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils git',
      'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
      'apt-get install -y nodejs',
      'npm install -g ts-node typescript typescript-language-server'
    ),
    timeout: 200
  ),
  on_snapshot_create_logs: method(:puts)
)
begin
  git = sandbox.git
  repo = 'demo-repo'

  cat = ->(file) { sandbox.process.exec(command: "cat #{file}", cwd: repo).result.strip }
  write = ->(path, content) { sandbox.fs.upload_file(content, path) }

  # ----------------------------- Git operations -----------------------------
  puts "git version: #{sandbox.process.exec(command: 'git --version').result.strip}"

  section('init')
  git.init(repo, initial_branch: 'main')
  puts "initialized repo at #{repo}"

  section('configure_user + get_config (local scope)')
  git.configure_user('Ada Lovelace', 'ada@example.com', scope: 'local', path: repo)
  puts "user.name  = #{git.get_config('user.name', scope: 'local', path: repo)}"
  puts "user.email = #{git.get_config('user.email', scope: 'local', path: repo)}"

  section('set_config / get_config (local) + unset key')
  git.set_config('core.editor', 'nano', scope: 'local', path: repo)
  puts "core.editor     = #{git.get_config('core.editor', scope: 'local', path: repo)}"
  puts "user.signingkey = #{git.get_config('user.signingkey', scope: 'local', path: repo).inspect} (unset -> nil)"

  section('remote_add / remotes / remote_get')
  git.remote_add(repo, 'origin', 'https://github.com/panaverse/learn-typescript.git')
  puts "remotes       = #{git.remotes(repo).remotes.map { |r| [r.name, r.url] }}"
  puts "remote_get    = #{git.remote_get(repo, 'origin')}"
  puts "remote_get(?) = #{git.remote_get(repo, 'upstream').inspect} (missing -> nil)"

  section('add / commit')
  write.call("#{repo}/a.txt", "line1\n")
  git.add(repo, ['a.txt'])
  commit = git.commit(path: repo, message: 'first commit', author: 'Ada Lovelace', email: 'ada@example.com')
  puts "commit sha = #{commit.sha}"

  section('branches (current marker)')
  b = git.branches(repo)
  puts "branches = #{b.branches} | current = #{b.current}"

  section('status (detached / upstream / current)')
  s = git.status(repo)
  puts "current=#{s.current_branch} detached=#{s.detached} upstream=#{s.upstream.inspect} ahead=#{s.ahead} behind=#{s.behind}"

  section('create_branch + delete_branch')
  git.create_branch(repo, 'feature')
  git.checkout_branch(repo, 'main')
  git.delete_branch(repo, 'feature')
  puts "deleted branch 'feature'"

  section('reset (mixed) -> unstage')
  write.call("#{repo}/b.txt", "staged\n")
  git.add(repo, ['b.txt'])
  puts "staged before reset: #{git.status(repo).file_status.map { |f| [f.name, f.staging] }}"
  git.reset(repo)
  puts "staged after reset : #{git.status(repo).file_status.map { |f| [f.name, f.staging] }}"

  section('restore (worktree) -> discard local changes')
  write.call("#{repo}/a.txt", "corrupted\n")
  puts "a.txt before restore: #{cat.call('a.txt')}"
  git.restore(repo, ['a.txt'])
  puts "a.txt after restore : #{cat.call('a.txt')}"

  section('reset (keep)')
  write.call("#{repo}/a.txt", "v2\n")
  git.add(repo, ['a.txt'])
  git.commit(path: repo, message: 'second commit', author: 'Ada Lovelace', email: 'ada@example.com')
  git.reset(repo, mode: 'keep', target: 'HEAD~1')
  puts "a.txt after keep reset to HEAD~1: #{cat.call('a.txt')}"

  section('clone (shallow, depth=1)')
  git.clone(url: 'https://github.com/panaverse/learn-typescript', path: 'shallow', branch: 'master', depth: 1)
  puts "shallow clone commit count (expect 1) = #{sandbox.process.exec(command: 'git rev-list --count HEAD',
                                                                       cwd: 'shallow').result.strip}"

  section('pull (remote + branch)')
  git.pull(path: 'shallow', remote: 'origin', branch: 'master')
  puts 'pulled origin/master into shallow clone (already up to date)'

  section('dangerously_authenticate')
  git.dangerously_authenticate('ci-bot', 'ghp_faketoken', host: 'example.com')
  puts "credential.helper (global) = #{git.get_config('credential.helper', scope: 'global')}"

  puts "\nAll new git operations exercised successfully ✅"

  # --------------------------------- LSP -----------------------------------
  project_dir = 'learn-typescript'

  section('clone project for LSP')
  # Clone the repository
  git.clone(
    url: 'https://github.com/panaverse/learn-typescript',
    path: project_dir,
    branch: 'master'
  )

  git.pull(path: project_dir)

  # Search for the file we want to work on
  matches = sandbox.fs.find_files(project_dir, 'var obj1 = new Base();')
  puts "Matches: #{matches}"

  section('LSP: document symbols + completions')
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
ensure
  daytona.delete(sandbox)
end
