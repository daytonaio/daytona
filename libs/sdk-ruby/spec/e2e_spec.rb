# frozen_string_literal: true

require 'timeout'
require 'spec_helper'

WebMock.allow_net_connect!

RSpec.describe 'Daytona SDK E2E', :e2e do
  SESSION_ID = 'test-session'

  before(:all) do
    skip 'DAYTONA_API_KEY not set' unless ENV['DAYTONA_API_KEY']

    @daytona = Daytona::Daytona.new
    params = Daytona::CreateSandboxFromSnapshotParams.new(language: Daytona::CodeLanguage::PYTHON)
    @sandbox = @daytona.create(params)
  end

  after(:all) do
    @sandbox&.process&.delete_session(SESSION_ID)
  rescue StandardError
    nil
  ensure
    begin
      @daytona&.delete(@sandbox) if @sandbox
    rescue StandardError => e
      puts "Cleanup error: #{e.message}"
    end
  end

  around(:each) do |example|
    Timeout.timeout(45) { example.run }
  end

  it 'validates sandbox lifecycle details and updates' do
    expect(@sandbox.id).to be_a(String)
    expect(@sandbox.id).not_to be_empty
    expect(@sandbox.state).to eq('started')

    user_home_dir = @sandbox.get_user_home_dir
    expect(user_home_dir).to be_a(String)
    expect(user_home_dir).to include('/')

    work_dir = @sandbox.get_work_dir
    expect(work_dir).to be_a(String)
    expect(work_dir).to include('/')

    @sandbox.labels = { 'test' => 'e2e' }
    expect(@sandbox.labels).to include('test' => 'e2e')

    @sandbox.auto_stop_interval = 30
    expect(@sandbox.auto_stop_interval).to eq(30)

    @sandbox.refresh
    expect(@sandbox.state).to eq('started')
  end

  it 'performs file system operations' do
    @sandbox.fs.create_folder('test-dir', '755')
    @sandbox.fs.upload_file('hello world', 'test-dir/hello.txt')

    files = @sandbox.fs.list_files('test-dir')
    names = files.map(&:name)
    expect(names).to include('hello.txt')

    file_info = @sandbox.fs.get_file_info('test-dir/hello.txt')
    expect(file_info).not_to be_nil
    expect(file_info.name).to eq('hello.txt')

    downloaded_file = @sandbox.fs.download_file('test-dir/hello.txt')
    content = downloaded_file.open.read
    expect(content).to include('hello world')

    found_matches = @sandbox.fs.find_files('test-dir', 'hello')
    expect(found_matches).not_to be_nil
    expect(found_matches.length).to be > 0

    search_result = @sandbox.fs.search_files('test-dir', '*.txt')
    expect(search_result.files).to include('test-dir/hello.txt')

    replace_result = @sandbox.fs.replace_in_files(
      files: ['test-dir/hello.txt'],
      pattern: 'hello',
      new_value: 'world'
    )
    expect(replace_result).not_to be_nil

    replaced_content = @sandbox.fs.download_file('test-dir/hello.txt').open.read
    expect(replaced_content).to include('world world')

    @sandbox.fs.move_files('test-dir/hello.txt', 'test-dir/moved.txt')
    moved_info = @sandbox.fs.get_file_info('test-dir/moved.txt')
    expect(moved_info.name).to eq('moved.txt')

    @sandbox.fs.delete_file('test-dir/moved.txt')

    remaining = @sandbox.fs.list_files('test-dir').map(&:name)
    expect(remaining).not_to include('moved.txt')
  end

  it 'executes commands and code' do
    response = @sandbox.process.exec(command: 'echo hello')
    expect(response.exit_code).to eq(0)
    expect(response.result).to include('hello')

    cwd_response = @sandbox.process.exec(command: 'pwd', cwd: '/tmp')
    expect(cwd_response.exit_code).to eq(0)
    expect(cwd_response.result).to include('/tmp')

    env_response = @sandbox.process.exec(command: 'echo $MY_E2E_ENV', env: { 'MY_E2E_ENV' => 'works' })
    expect(env_response.exit_code).to eq(0)
    expect(env_response.result).to include('works')

    code_response = @sandbox.process.code_run(code: 'print("hello from python")')
    expect(code_response.exit_code).to eq(0)
    expect(code_response.result).to include('hello from python')

    fail_response = @sandbox.process.exec(command: 'exit 1')
    expect(fail_response.exit_code).not_to eq(0)
  end

  it 'manages process sessions' do
    @sandbox.process.create_session(SESSION_ID)

    session = @sandbox.process.get_session(SESSION_ID)
    expect(session).not_to be_nil
    expect(session.session_id).to eq(SESSION_ID)

    export_response = @sandbox.process.execute_session_command(
      session_id: SESSION_ID,
      req: Daytona::SessionExecuteRequest.new(command: 'export FOO=bar')
    )
    expect(export_response.exit_code).to eq(0)

    echo_response = @sandbox.process.execute_session_command(
      session_id: SESSION_ID,
      req: Daytona::SessionExecuteRequest.new(command: 'echo $FOO')
    )
    expect(echo_response.exit_code).to eq(0)
    expect(echo_response.stdout).to include('bar')

    sessions = @sandbox.process.list_sessions
    session_ids = sessions.map(&:session_id)
    expect(session_ids).to include(SESSION_ID)

    @sandbox.process.delete_session(SESSION_ID)
    session_ids_after_delete = @sandbox.process.list_sessions.map(&:session_id)
    expect(session_ids_after_delete).not_to include(SESSION_ID)
  end

  it 'runs git operations' do
    @sandbox.git.clone(url: 'https://github.com/octocat/Hello-World.git', path: 'hello-world')

    git_status = @sandbox.git.status('hello-world')
    expect(git_status).not_to be_nil

    git_branches = @sandbox.git.branches('hello-world')
    expect(git_branches).not_to be_nil
    expect(git_branches.branches).to be_a(Array)
    expect(git_branches.branches.length).to be > 0
  end

  it 'runs daytona client operations' do
    list_response = @daytona.list
    expect(list_response.total).to be > 0

    fetched = @daytona.get(@sandbox.id)
    expect(fetched.id).to eq(@sandbox.id)
  end
end
