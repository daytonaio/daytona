# frozen_string_literal: true

require 'timeout'
require 'securerandom'
require 'spec_helper'

WebMock.allow_net_connect!

RSpec.describe 'Daytona SDK E2E', :e2e do
  before(:all) do
    skip 'DAYTONA_API_KEY not set' unless ENV['DAYTONA_API_KEY']

    @daytona = Daytona::Daytona.new
    params = Daytona::CreateSandboxFromSnapshotParams.new(language: Daytona::CodeLanguage::PYTHON)
    @sandbox = @daytona.create(params)

    @session_id = "e2e-sess-#{SecureRandom.hex(4)}"
    @volume_name = "e2e-vol-#{SecureRandom.hex(4)}"
    @git_repo_path = '/tmp/e2e-hello-world'
    @fs_dir = '/tmp/e2e-fs-tests'
    @shared = {}
  end

  after(:all) do
    begin; @sandbox&.process&.delete_session(@session_id); rescue StandardError; nil; end
    if @shared&.dig(:interpreter_context)
      begin; @sandbox&.code_interpreter&.delete_context(@shared[:interpreter_context]); rescue StandardError; nil; end
    end
    if @shared&.dig(:volume)
      begin; @daytona&.volume&.delete(@shared[:volume]); rescue StandardError; nil; end
    end
    begin
      @daytona&.delete(@sandbox) if @sandbox
    rescue StandardError => e
      puts "Cleanup error: #{e.message}"
    end
  end

  around(:each) do |example|
    Timeout.timeout(60) { example.run }
  end

  context 'Sandbox Lifecycle', order: :defined do
    it 'has a valid non-empty id' do
      expect(@sandbox.id).to be_a(String)
      expect(@sandbox.id).not_to be_empty
    end

    it 'has state started' do
      expect(@sandbox.state).to eq('started')
    end

    it 'has resource properties (cpu, memory, disk) > 0' do
      expect(@sandbox.cpu).to be > 0
      expect(@sandbox.memory).to be > 0
      expect(@sandbox.disk).to be > 0
    end

    it 'has created_at timestamp' do
      expect(@sandbox.created_at).to be_a(String)
      expect(@sandbox.created_at).not_to be_empty
    end

    it 'returns home directory from get_user_home_dir' do
      home_dir = @sandbox.get_user_home_dir
      expect(home_dir).to be_a(String)
      expect(home_dir).to include('/')
    end

    it 'returns working directory from get_work_dir' do
      work_dir = @sandbox.get_work_dir
      expect(work_dir).to be_a(String)
      expect(work_dir).to include('/')
    end

    it 'sets labels via labels=' do
      @sandbox.labels = { 'test' => 'e2e', 'env' => 'ci' }
      expect(@sandbox.labels).to include('test' => 'e2e')
      expect(@sandbox.labels).to include('env' => 'ci')
    end

    it 'updates auto_stop_interval' do
      @sandbox.auto_stop_interval = 30
      expect(@sandbox.auto_stop_interval).to eq(30)
      # Reset
      @sandbox.auto_stop_interval = 15
    end

    it 'updates auto_archive_interval' do
      @sandbox.auto_archive_interval = 20_000
      expect(@sandbox.auto_archive_interval).to eq(20_000)
      # Reset
      @sandbox.auto_archive_interval = 10_080
    end

    it 'updates auto_delete_interval and disables it' do
      @sandbox.auto_delete_interval = 1440
      expect(@sandbox.auto_delete_interval).to eq(1440)
      # Disable
      @sandbox.auto_delete_interval = -1
      expect(@sandbox.auto_delete_interval).to eq(-1)
    end

    it 'refreshes sandbox data' do
      @sandbox.refresh
      expect(@sandbox.id).not_to be_empty
      expect(@sandbox.state).to eq('started')
    end

    it 'refreshes sandbox activity' do
      expect { @sandbox.refresh_activity }.not_to raise_error
    end

    it 'stops and starts the sandbox' do
      @sandbox.stop
      @sandbox.refresh
      expect(@sandbox.state).to eq('stopped')

      @sandbox.start
      @sandbox.refresh
      expect(@sandbox.state).to eq('started')
    end
  end

  context 'File System', order: :defined do
    it 'creates a folder' do
      expect { @sandbox.fs.create_folder(@fs_dir, '755') }.not_to raise_error
    end

    it 'uploads a file from string content' do
      @sandbox.fs.upload_file('hello world', "#{@fs_dir}/hello.txt")
      info = @sandbox.fs.get_file_info("#{@fs_dir}/hello.txt")
      expect(info.name).to eq('hello.txt')
    end

    it 'uploads multiple files via upload_files' do
      files = [
        Daytona::FileUpload.new('content A', "#{@fs_dir}/multi_a.txt"),
        Daytona::FileUpload.new('content B', "#{@fs_dir}/multi_b.txt")
      ]
      @sandbox.fs.upload_files(files)

      info_a = @sandbox.fs.get_file_info("#{@fs_dir}/multi_a.txt")
      info_b = @sandbox.fs.get_file_info("#{@fs_dir}/multi_b.txt")
      expect(info_a.name).to eq('multi_a.txt')
      expect(info_b.name).to eq('multi_b.txt')
    end

    it 'lists files in directory' do
      files = @sandbox.fs.list_files(@fs_dir)
      names = files.map(&:name)
      expect(names).to include('hello.txt')
      expect(names).to include('multi_a.txt')
      expect(names).to include('multi_b.txt')
    end

    it 'gets file info with correct name and size' do
      info = @sandbox.fs.get_file_info("#{@fs_dir}/hello.txt")
      expect(info.name).to eq('hello.txt')
      expect(info.size).to be >= 11
    end

    it 'downloads file with matching content' do
      downloaded = @sandbox.fs.download_file("#{@fs_dir}/hello.txt")
      content = downloaded.open.read
      expect(content).to include('hello world')
    end

    it 'finds text content in files' do
      matches = @sandbox.fs.find_files(@fs_dir, 'hello')
      expect(matches).not_to be_nil
      expect(matches.length).to be > 0
    end

    it 'searches files by glob pattern' do
      result = @sandbox.fs.search_files(@fs_dir, '*.txt')
      expect(result.files).to be_a(Array)
      expect(result.files.length).to be >= 1
    end

    it 'replaces text in files' do
      result = @sandbox.fs.replace_in_files(
        files: ["#{@fs_dir}/hello.txt"],
        pattern: 'hello',
        new_value: 'goodbye'
      )
      expect(result).not_to be_nil
    end

    it 'verifies replacement by downloading again' do
      content = @sandbox.fs.download_file("#{@fs_dir}/hello.txt").open.read
      expect(content).to include('goodbye world')
    end

    it 'sets file permissions' do
      expect {
        @sandbox.fs.set_file_permissions(path: "#{@fs_dir}/multi_a.txt", mode: '644')
      }.not_to raise_error
    end

    it 'moves files to new location' do
      @sandbox.fs.upload_file('moveable content', "#{@fs_dir}/to_move.txt")
      @sandbox.fs.move_files("#{@fs_dir}/to_move.txt", "#{@fs_dir}/moved.txt")
      info = @sandbox.fs.get_file_info("#{@fs_dir}/moved.txt")
      expect(info.name).to eq('moved.txt')
    end

    it 'deletes a file' do
      @sandbox.fs.upload_file('delete me', "#{@fs_dir}/delete_me.txt")
      @sandbox.fs.delete_file("#{@fs_dir}/delete_me.txt")
      remaining = @sandbox.fs.list_files(@fs_dir).map(&:name)
      expect(remaining).not_to include('delete_me.txt')
    end

    it 'handles nested directory operations' do
      @sandbox.fs.create_folder("#{@fs_dir}/nested/deep", '755')
      @sandbox.fs.upload_file('deep content', "#{@fs_dir}/nested/deep/file.txt")
      info = @sandbox.fs.get_file_info("#{@fs_dir}/nested/deep/file.txt")
      expect(info.name).to eq('file.txt')

      content = @sandbox.fs.download_file("#{@fs_dir}/nested/deep/file.txt").open.read
      expect(content).to eq('deep content')
    end
  end

  context 'Process Execution', order: :defined do
    it 'executes basic echo command' do
      response = @sandbox.process.exec(command: 'echo hello')
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('hello')
    end

    it 'executes command with cwd option' do
      response = @sandbox.process.exec(command: 'pwd', cwd: '/tmp')
      expect(response.exit_code).to eq(0)
      expect(response.result.strip).to eq('/tmp')
    end

    it 'executes command with env vars' do
      response = @sandbox.process.exec(command: 'echo $MY_E2E_VAR', env: { 'MY_E2E_VAR' => 'e2e_value' })
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('e2e_value')
    end

    it 'returns non-zero exit code on failure' do
      response = @sandbox.process.exec(command: 'exit 42')
      expect(response.exit_code).to eq(42)
    end

    it 'captures stderr output' do
      response = @sandbox.process.exec(command: 'echo err >&2 && echo ok')
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('ok')
    end

    it 'runs Python code via code_run' do
      response = @sandbox.process.code_run(code: 'print("hello from python")')
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('hello from python')
    end

    it 'runs multi-line Python code' do
      code = <<~PYTHON
        x = 10
        y = 20
        print(f"sum={x + y}")
      PYTHON
      response = @sandbox.process.code_run(code: code)
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('sum=30')
    end

    it 'handles code_run with syntax error' do
      response = @sandbox.process.code_run(code: 'def bad(')
      expect(response.exit_code).not_to eq(0)
    end
  end

  context 'Sessions', order: :defined do
    it 'creates a session' do
      expect { @sandbox.process.create_session(@session_id) }.not_to raise_error
    end

    it 'gets session details' do
      session = @sandbox.process.get_session(@session_id)
      expect(session).not_to be_nil
      expect(session.session_id).to eq(@session_id)
    end

    it 'executes command in session' do
      response = @sandbox.process.execute_session_command(
        session_id: @session_id,
        req: Daytona::SessionExecuteRequest.new(command: 'echo session_test')
      )
      expect(response.exit_code).to eq(0)
      expect(response.stdout).to include('session_test')
    end

    it 'maintains state across session commands' do
      @sandbox.process.execute_session_command(
        session_id: @session_id,
        req: Daytona::SessionExecuteRequest.new(command: 'export SESSION_VAR=persistent')
      )
      response = @sandbox.process.execute_session_command(
        session_id: @session_id,
        req: Daytona::SessionExecuteRequest.new(command: 'echo $SESSION_VAR')
      )
      expect(response.exit_code).to eq(0)
      expect(response.stdout).to include('persistent')
    end

    it 'lists sessions including ours' do
      sessions = @sandbox.process.list_sessions
      session_ids = sessions.map(&:session_id)
      expect(session_ids).to include(@session_id)
    end

    it 'gets session command logs' do
      cmd_response = @sandbox.process.execute_session_command(
        session_id: @session_id,
        req: Daytona::SessionExecuteRequest.new(command: 'echo logs_test')
      )
      logs = @sandbox.process.get_session_command_logs(
        session_id: @session_id,
        command_id: cmd_response.cmd_id
      )
      expect(logs).not_to be_nil
      expect(logs.stdout).to include('logs_test')
    end

    it 'deletes a session' do
      @sandbox.process.delete_session(@session_id)
      sessions = @sandbox.process.list_sessions.map(&:session_id)
      expect(sessions).not_to include(@session_id)
      @session_id = "e2e-deleted-#{SecureRandom.hex(4)}"
    end
  end

  context 'Git Operations', order: :defined do
    it 'clones a public repo' do
      expect {
        @sandbox.git.clone(url: 'https://github.com/octocat/Hello-World.git', path: @git_repo_path)
      }.not_to raise_error
    end

    it 'gets git status with current branch' do
      status = @sandbox.git.status(@git_repo_path)
      expect(status).not_to be_nil
      expect(status.current_branch).to be_a(String)
      expect(status.current_branch).not_to be_empty
    end

    it 'lists branches' do
      branches = @sandbox.git.branches(@git_repo_path)
      expect(branches.branches).to be_a(Array)
      expect(branches.branches.length).to be > 0
    end

    it 'adds files to staging' do
      @sandbox.fs.upload_file('e2e git test', "#{@git_repo_path}/e2e_file.txt")
      expect {
        @sandbox.git.add(@git_repo_path, ['e2e_file.txt'])
      }.not_to raise_error
    end

    it 'commits with message via exec' do
      @sandbox.process.exec(command: "git config --global user.email 'e2e@test.com'")
      @sandbox.process.exec(command: "git config --global user.name 'E2E Test'")
      response = @sandbox.process.exec(
        command: "cd #{@git_repo_path} && git commit -m 'E2E test commit'"
      )
      expect(response.exit_code).to eq(0)
      expect(response.result).to include('E2E test commit')
    end

    it 'creates a new branch' do
      expect {
        @sandbox.git.create_branch(@git_repo_path, 'e2e-test-branch')
      }.not_to raise_error
    end

    it 'checks out a branch' do
      @sandbox.git.checkout_branch(@git_repo_path, 'e2e-test-branch')
      status = @sandbox.git.status(@git_repo_path)
      expect(status.current_branch).to eq('e2e-test-branch')
    end

    it 'deletes a branch via exec' do
      @sandbox.git.checkout_branch(@git_repo_path, 'master')
      response = @sandbox.process.exec(command: "cd #{@git_repo_path} && git branch -d e2e-test-branch")
      expect(response.exit_code).to eq(0)
    end
  end

  context 'Code Interpreter', order: :defined do
    it 'runs simple Python code' do
      result = @sandbox.code_interpreter.run_code('print("interpreter hello")')
      expect(result).to be_a(Daytona::ExecutionResult)
      expect(result.stdout).to include('interpreter hello')
    end

    it 'maintains state across runs in default context' do
      @sandbox.code_interpreter.run_code('ci_var = 42')
      result = @sandbox.code_interpreter.run_code('print(ci_var)')
      expect(result.stdout.strip).to include('42')
    end

    it 'creates an isolated context' do
      @shared[:interpreter_context] = @sandbox.code_interpreter.create_context
      expect(@shared[:interpreter_context]).not_to be_nil
      expect(@shared[:interpreter_context].id).to be_a(String)
    end

    it 'runs code in isolated context' do
      ctx = @shared[:interpreter_context]
      @sandbox.code_interpreter.run_code('isolated_val = 99', context: ctx)
      result = @sandbox.code_interpreter.run_code('print(isolated_val)', context: ctx)
      expect(result.stdout.strip).to include('99')
    end

    it 'lists contexts including created one' do
      contexts = @sandbox.code_interpreter.list_contexts
      expect(contexts).to be_a(Array)
      ids = contexts.map(&:id)
      expect(ids).to include(@shared[:interpreter_context].id)
    end

    it 'deletes a context' do
      ctx_to_delete = @sandbox.code_interpreter.create_context
      expect {
        @sandbox.code_interpreter.delete_context(ctx_to_delete)
      }.not_to raise_error
    end
  end

  context 'Volume Management', order: :defined do
    it 'creates a volume' do
      vol = @daytona.volume.create(@volume_name)
      @shared[:volume] = vol
      expect(vol).not_to be_nil
      expect(vol.name).to eq(@volume_name)
      expect(vol.id).to be_a(String)
    end

    it 'lists volumes including created one' do
      volumes = @daytona.volume.list
      expect(volumes).to be_a(Array)
      names = volumes.map(&:name)
      expect(names).to include(@volume_name)
    end

    it 'gets volume by name' do
      vol = @daytona.volume.get(@volume_name)
      expect(vol.name).to eq(@volume_name)
      expect(vol.id).to be_a(String)
    end

    it 'deletes a volume' do
      vol = @daytona.volume.get(@volume_name)
      attempts = 0
      while vol.state != 'ready' && attempts < 10
        sleep 1
        vol = @daytona.volume.get(@volume_name)
        attempts += 1
      end
      expect { @daytona.volume.delete(vol) }.not_to raise_error
      @shared[:volume] = nil
    end
  end

  context 'Snapshot Operations', order: :defined do
    it 'lists snapshots' do
      result = @daytona.snapshot.list
      expect(result).to be_a(Daytona::PaginatedResource)
      expect(result.total).to be >= 0
      expect(result.items).to be_a(Array)
    end

    it 'lists with pagination' do
      result = @daytona.snapshot.list(page: 1, limit: 2)
      expect(result.page).to eq(1)
      expect(result.items.length).to be <= 2
    end

    it 'gets snapshot by name' do
      # Use the first available snapshot from list
      list_result = @daytona.snapshot.list(page: 1, limit: 1)
      skip 'No snapshots available' if list_result.items.empty?

      snapshot_name = list_result.items.first.name
      snapshot = @daytona.snapshot.get(snapshot_name)
      expect(snapshot).to be_a(Daytona::Snapshot)
      expect(snapshot.name).to eq(snapshot_name)
    end
  end

  context 'Client Operations', order: :defined do
    it 'lists sandboxes' do
      result = @daytona.list
      expect(result).to be_a(Daytona::PaginatedResource)
      expect(result.total).to be > 0
      expect(result.items).to be_a(Array)
    end

    it 'lists with pagination' do
      result = @daytona.list({}, page: 1, limit: 2)
      expect(result.page).to eq(1)
      expect(result.items.length).to be <= 2
    end

    it 'gets sandbox by id' do
      fetched = @daytona.get(@sandbox.id)
      expect(fetched.id).to eq(@sandbox.id)
      expect(fetched.state).to eq('started')
    end

    it 'lists sandboxes filtered by labels' do
      @sandbox.labels = { 'test' => 'e2e' } unless @sandbox.labels&.dig('test') == 'e2e'
      result = @daytona.list({ 'test' => 'e2e' })
      expect(result).to be_a(Daytona::PaginatedResource)
      ids = result.items.map(&:id)
      expect(ids).to include(@sandbox.id)
    end
  end
end
