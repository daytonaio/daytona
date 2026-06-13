# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

module Daytona
  class Git
    include Instrumentation

    # @return [String] The Sandbox ID
    attr_reader :sandbox_id

    # @return [DaytonaToolboxApiClient::GitApi] API client for Sandbox operations
    attr_reader :toolbox_api

    # Initializes a new Git handler instance.
    #
    # @param sandbox_id [String] The Sandbox ID.
    # @param toolbox_api [DaytonaToolboxApiClient::GitApi] API client for Sandbox operations.
    # @param otel_state [Daytona::OtelState, nil]
    def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @otel_state = otel_state
    end

    # Stages the specified files for the next commit, similar to
    # running 'git add' on the command line.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param files [Array<String>] List of file paths or directories to stage, relative to the repository root.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if adding files fails
    #
    # @example
    #   # Stage a single file
    #   sandbox.git.add("workspace/repo", ["file.txt"])
    #
    #   # Stage multiple files
    #   sandbox.git.add("workspace/repo", [
    #     "src/main.rb",
    #     "spec/main_spec.rb",
    #     "README.md"
    #   ])
    def add(path, files)
      toolbox_api.add_files(DaytonaToolboxApiClient::GitAddRequest.new(path:, files:))
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to add files')
    rescue StandardError => e
      raise Sdk::Error, "Failed to add files: #{e.message}"
    end

    # Lists branches in the repository.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @return [DaytonaApiClient::ListBranchResponse] List of branches in the repository.
    # @raise [Daytona::Sdk::Error] if listing branches fails
    #
    # @example
    #   response = sandbox.git.branches("workspace/repo")
    #   puts "Branches: #{response.branches}"
    def branches(path)
      toolbox_api.list_branches(path)
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to list branches')
    rescue StandardError => e
      raise Sdk::Error, "Failed to list branches: #{e.message}"
    end

    # Clones a Git repository into the specified path. It supports
    # cloning specific branches or commits, and can authenticate with the remote
    # repository if credentials are provided.
    #
    # @param url [String] Repository URL to clone from.
    # @param path [String] Path where the repository should be cloned. Relative paths are resolved
    #   based on the sandbox working directory.
    # @param branch [String, nil] Specific branch to clone. If not specified,
    #   clones the default branch.
    # @param commit_id [String, nil] Specific commit to clone. If specified,
    #   the repository will be left in a detached HEAD state at this commit.
    # @param username [String, nil] Git username for authentication.
    # @param password [String, nil] Git password or token for authentication.
    # @param insecure_skip_tls [Boolean, nil] Skip TLS certificate verification (insecure).
    #   Use only for trusted internal Git servers with self-signed or private-CA certs;
    #   credentials, if supplied, are transmitted over an unverified TLS connection.
    # @param depth [Integer, nil] Create a shallow clone truncated to the given number of commits.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if cloning repository fails
    #
    # @example
    #   # Clone the default branch
    #   sandbox.git.clone(
    #     url: "https://github.com/user/repo.git",
    #     path: "workspace/repo"
    #   )
    #
    #   # Clone a specific branch with authentication
    #   sandbox.git.clone(
    #     url: "https://github.com/user/private-repo.git",
    #     path: "workspace/private",
    #     branch: "develop",
    #     username: "user",
    #     password: "token"
    #   )
    #
    #   # Clone a specific commit
    #   sandbox.git.clone(
    #     url: "https://github.com/user/repo.git",
    #     path: "workspace/repo-old",
    #     commit_id: "abc123"
    #   )
    def clone(url:, path:, branch: nil, commit_id: nil, username: nil, password: nil, insecure_skip_tls: nil, depth: nil) # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists, Layout/LineLength
      toolbox_api.clone_repository(
        DaytonaToolboxApiClient::GitCloneRequest.new(
          url: url,
          branch: branch,
          path: path,
          username: username,
          password: password,
          commit_id: commit_id,
          insecure_skip_tls: insecure_skip_tls,
          depth: depth
        )
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to clone repository')
    rescue StandardError => e
      raise Sdk::Error, "Failed to clone repository: #{e.message}"
    end

    # Creates a new commit with the staged changes. Make sure to stage
    # changes using the add() method before committing.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param message [String] Commit message describing the changes.
    # @param author [String] Name of the commit author.
    # @param email [String] Email address of the commit author.
    # @param allow_empty [Boolean] Allow creating an empty commit when no changes are staged. Defaults to false.
    # @return [GitCommitResponse] Response containing the commit SHA.
    # @raise [Daytona::Sdk::Error] if committing changes fails
    #
    # @example
    #   # Stage and commit changes
    #   sandbox.git.add("workspace/repo", ["README.md"])
    #   commit_response = sandbox.git.commit(
    #     path: "workspace/repo",
    #     message: "Update documentation",
    #     author: "John Doe",
    #     email: "john@example.com",
    #     allow_empty: true
    #   )
    #   puts "Commit SHA: #{commit_response.sha}"
    def commit(path:, message:, author:, email:, allow_empty: false)
      response = toolbox_api.commit_changes(
        DaytonaToolboxApiClient::GitCommitRequest.new(path:, message:, author:, email:, allow_empty:)
      )
      GitCommitResponse.new(sha: response._hash)
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to commit changes')
    rescue StandardError => e
      raise Sdk::Error, "Failed to commit changes: #{e.message}"
    end

    # Pushes all local commits on the current branch to the remote
    # repository. If the remote repository requires authentication, provide
    # username and password/token.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param username [String, nil] Git username for authentication.
    # @param password [String, nil] Git password or token for authentication.
    # @param branch [String, nil] Branch to push. Defaults to the current branch.
    # @param remote [String, nil] Remote to push to. Defaults to "origin".
    # @param set_upstream [Boolean] Record the pushed branch as the upstream tracking branch. Defaults to false.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if pushing changes fails
    #
    # @example
    #   # Push without authentication (for public repos or SSH)
    #   sandbox.git.push("workspace/repo")
    #
    #   # Push with authentication
    #   sandbox.git.push(
    #     path: "workspace/repo",
    #     username: "user",
    #     password: "github_token"
    #   )
    #
    #   # Push a new branch and set its upstream
    #   sandbox.git.push(path: "workspace/repo", branch: "feature", set_upstream: true)
    def push(path:, username: nil, password: nil, branch: nil, remote: nil, set_upstream: false) # rubocop:disable Metrics/ParameterLists
      toolbox_api.push_changes(
        DaytonaToolboxApiClient::GitPushRequest.new(path:, username:, password:, branch:, remote:, set_upstream:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to push changes')
    rescue StandardError => e
      raise Sdk::Error, "Failed to push changes: #{e.message}"
    end

    # Pulls changes from the remote repository. If the remote repository requires authentication,
    # provide username and password/token.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param username [String, nil] Git username for authentication.
    # @param password [String, nil] Git password or token for authentication.
    # @param branch [String, nil] Branch to pull. Defaults to the current branch's upstream.
    # @param remote [String, nil] Remote to pull from. Defaults to "origin".
    # @return [void]
    # @raise [Daytona::Sdk::Error] if pulling changes fails
    #
    # @example
    #   # Pull without authentication
    #   sandbox.git.pull("workspace/repo")
    #
    #   # Pull with authentication
    #   sandbox.git.pull(
    #     path: "workspace/repo",
    #     username: "user",
    #     password: "github_token"
    #   )
    #
    #   # Pull a specific branch from a specific remote
    #   sandbox.git.pull(path: "workspace/repo", remote: "upstream", branch: "main")
    def pull(path:, username: nil, password: nil, branch: nil, remote: nil)
      toolbox_api.pull_changes(
        DaytonaToolboxApiClient::GitPullRequest.new(path:, username:, password:, branch:, remote:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to pull changes')
    rescue StandardError => e
      raise Sdk::Error, "Failed to pull changes: #{e.message}"
    end

    # Gets the current Git repository status.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @return [DaytonaToolboxApiClient::GitStatus] Repository status information including:
    # @raise [Daytona::Sdk::Error] if getting status fails
    #
    # @example
    #   status = sandbox.git.status("workspace/repo")
    #   puts "On branch: #{status.current_branch}"
    #   puts "Commits ahead: #{status.ahead}"
    #   puts "Commits behind: #{status.behind}"
    def status(path)
      toolbox_api.get_status(path)
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to get status')
    rescue StandardError => e
      raise Sdk::Error, "Failed to get status: #{e.message}"
    end

    # Checkout branch in the repository.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param branch [String] Name of the branch to checkout
    # @return [void]
    # @raise [Daytona::Sdk::Error] if checking out branch fails
    #
    # @example
    #   # Checkout a branch
    #   sandbox.git.checkout_branch("workspace/repo", "feature-branch")
    def checkout_branch(path, branch)
      toolbox_api.checkout_branch(
        DaytonaToolboxApiClient::GitCheckoutRequest.new(path:, branch:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to checkout branch')
    rescue StandardError => e
      raise Sdk::Error, "Failed to checkout branch: #{e.message}"
    end

    # Create branch in the repository.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param name [String] Name of the new branch to create
    # @return [void]
    # @raise [Daytona::Sdk::Error] if creating branch fails
    #
    # @example
    #   # Create a new branch
    #   sandbox.git.create_branch("workspace/repo", "new-feature")
    #
    def create_branch(path, name)
      toolbox_api.create_branch(
        DaytonaToolboxApiClient::GitBranchRequest.new(path:, name:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to create branch')
    rescue StandardError => e
      raise Sdk::Error, "Failed to create branch: #{e.message}"
    end

    # Delete branch in the repository.
    #
    # @param path [String] Path to the Git repository root. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param name [String] Name of the branch to delete
    # @return [void]
    # @raise [Daytona::Sdk::Error] if deleting branch fails
    #
    # @example
    #   # Delete a branch
    #   sandbox.git.delete_branch("workspace/repo", "old-feature")
    def delete_branch(path, name)
      toolbox_api.delete_branch(
        DaytonaToolboxApiClient::GitDeleteBranchRequest.new(path:, name:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to delete branch')
    rescue StandardError => e
      raise Sdk::Error, "Failed to delete branch: #{e.message}"
    end

    # Initializes a new Git repository at the specified path.
    #
    # @param path [String] Path where the repository should be initialized.
    # @param bare [Boolean] Create a bare repository without a working tree. Defaults to false.
    # @param initial_branch [String, nil] Name of the initial branch. If not specified, uses the Git default.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if initializing repository fails
    #
    # @example
    #   sandbox.git.init("workspace/repo", initial_branch: "main")
    def init(path, bare: false, initial_branch: nil)
      toolbox_api.init_repository(
        DaytonaToolboxApiClient::GitInitRequest.new(path:, bare:, initial_branch:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to initialize repository')
    rescue StandardError => e
      raise Sdk::Error, "Failed to initialize repository: #{e.message}"
    end

    # Resets the current HEAD to the specified state.
    #
    # @param path [String] Path to the Git repository root.
    # @param mode [String, nil] Reset mode, one of "soft", "mixed" (default), "hard", "merge" or "keep".
    # @param target [String, nil] Revision to reset to. Defaults to HEAD.
    # @param files [Array<String>, nil] Constrain the reset to the given paths.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if resetting fails
    #
    # @example
    #   # Unstage all changes (mixed reset to HEAD)
    #   sandbox.git.reset("workspace/repo")
    #
    #   # Hard reset to a previous commit
    #   sandbox.git.reset("workspace/repo", mode: "hard", target: "HEAD~1")
    def reset(path, mode: nil, target: nil, files: nil)
      toolbox_api.reset_changes(
        DaytonaToolboxApiClient::GitResetRequest.new(path:, mode:, target:, files:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to reset')
    rescue StandardError => e
      raise Sdk::Error, "Failed to reset: #{e.message}"
    end

    # Restores working tree files or unstages changes.
    #
    # @param path [String] Path to the Git repository root.
    # @param files [Array<String>] File paths to restore.
    # @param staged [Boolean, nil] Restore the staging index for the given files.
    # @param worktree [Boolean, nil] Restore the working tree for the given files. Defaults to true
    #   when neither staged nor worktree is provided.
    # @param source [String, nil] Restore file contents from the given revision instead of the index.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if restoring fails
    #
    # @example
    #   # Discard working tree changes
    #   sandbox.git.restore("workspace/repo", ["file.txt"])
    #
    #   # Unstage changes
    #   sandbox.git.restore("workspace/repo", ["file.txt"], staged: true)
    def restore(path, files, staged: nil, worktree: nil, source: nil)
      toolbox_api.restore_files(
        DaytonaToolboxApiClient::GitRestoreRequest.new(path:, files:, staged:, worktree:, source:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to restore files')
    rescue StandardError => e
      raise Sdk::Error, "Failed to restore files: #{e.message}"
    end

    # Adds (or overwrites) a remote in the repository.
    #
    # @param path [String] Path to the Git repository root.
    # @param name [String] Name of the remote.
    # @param url [String] URL of the remote.
    # @param fetch [Boolean] Fetch from the remote immediately after adding it. Defaults to false.
    # @param overwrite [Boolean] Replace an existing remote with the same name. Defaults to false.
    # @return [void]
    # @raise [Daytona::Sdk::Error] if adding the remote fails
    #
    # @example
    #   sandbox.git.remote_add("workspace/repo", "origin", "https://github.com/user/repo.git")
    def remote_add(path, name, url, fetch: false, overwrite: false)
      toolbox_api.add_remote(
        DaytonaToolboxApiClient::GitAddRemoteRequest.new(path:, name:, url:, fetch:, overwrite:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to add remote')
    rescue StandardError => e
      raise Sdk::Error, "Failed to add remote: #{e.message}"
    end

    # Lists the remotes configured in the repository.
    #
    # @param path [String] Path to the Git repository root.
    # @return [DaytonaToolboxApiClient::ListRemotesResponse] The configured remotes (name + URL).
    # @raise [Daytona::Sdk::Error] if listing remotes fails
    #
    # @example
    #   response = sandbox.git.remotes("workspace/repo")
    #   response.remotes.each { |r| puts "#{r.name}: #{r.url}" }
    def remotes(path)
      toolbox_api.list_remotes(path)
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to list remotes')
    rescue StandardError => e
      raise Sdk::Error, "Failed to list remotes: #{e.message}"
    end

    # Gets the URL of a remote, or nil when it does not exist.
    #
    # @param path [String] Path to the Git repository root.
    # @param name [String] Name of the remote.
    # @return [String, nil] The remote URL, or nil when the remote does not exist.
    # @raise [Daytona::Sdk::Error] if getting the remote fails
    #
    # @example
    #   url = sandbox.git.remote_get("workspace/repo", "origin")
    def remote_get(path, name)
      toolbox_api.list_remotes(path).remotes.find { |r| r.name == name }&.url
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to get remote')
    rescue StandardError => e
      raise Sdk::Error, "Failed to get remote: #{e.message}"
    end

    # Sets a Git config value at the given scope.
    #
    # @param key [String] Config key in dotted form (e.g. "user.name").
    # @param value [String] Config value.
    # @param scope [String] Config scope, one of "global" (default), "local" or "system".
    # @param path [String, nil] Repository path, required when scope is "local".
    # @return [void]
    # @raise [Daytona::Sdk::Error] if setting config fails
    #
    # @example
    #   sandbox.git.set_config("user.name", "John Doe")
    def set_config(key, value, scope: 'global', path: nil)
      toolbox_api.set_git_config(
        DaytonaToolboxApiClient::GitSetConfigRequest.new(key:, value:, scope:, path:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to set config')
    rescue StandardError => e
      raise Sdk::Error, "Failed to set config: #{e.message}"
    end

    # Gets a Git config value at the given scope, or nil when unset.
    #
    # @param key [String] Config key in dotted form (e.g. "user.name").
    # @param scope [String] Config scope, one of "global" (default), "local" or "system".
    # @param path [String, nil] Repository path, required when scope is "local".
    # @return [String, nil] The config value, or nil when the key is not set.
    # @raise [Daytona::Sdk::Error] if getting config fails
    #
    # @example
    #   name = sandbox.git.get_config("user.name")
    def get_config(key, scope: 'global', path: nil)
      toolbox_api.get_git_config(key, scope: scope, path: path).value
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to get config')
    rescue StandardError => e
      raise Sdk::Error, "Failed to get config: #{e.message}"
    end

    # Configures the Git user name and email at the given scope.
    #
    # @param name [String] User name (user.name).
    # @param email [String] User email (user.email).
    # @param scope [String] Config scope, one of "global" (default), "local" or "system".
    # @param path [String, nil] Repository path, required when scope is "local".
    # @return [void]
    # @raise [Daytona::Sdk::Error] if configuring user fails
    #
    # @example
    #   sandbox.git.configure_user("John Doe", "john@example.com")
    def configure_user(name, email, scope: 'global', path: nil)
      toolbox_api.configure_user(
        DaytonaToolboxApiClient::GitConfigureUserRequest.new(name:, email:, scope:, path:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to configure user')
    rescue StandardError => e
      raise Sdk::Error, "Failed to configure user: #{e.message}"
    end

    # Persists Git credentials globally so that subsequent operations against the
    # given host authenticate automatically.
    #
    # WARNING: This stores the password in plaintext on disk via the Git credential store.
    #
    # @param username [String] Git username.
    # @param password [String] Git password or token.
    # @param host [String, nil] Host to authenticate against. Defaults to "github.com".
    # @param protocol [String, nil] Protocol to authenticate against. Defaults to "https".
    # @return [void]
    # @raise [Daytona::Sdk::Error] if authenticating fails
    #
    # @example
    #   sandbox.git.dangerously_authenticate("user", "github_token")
    def dangerously_authenticate(username, password, host: nil, protocol: nil)
      toolbox_api.authenticate(
        DaytonaToolboxApiClient::GitAuthenticateRequest.new(username:, password:, host:, protocol:)
      )
    rescue DaytonaToolboxApiClient::ApiError => e
      raise map_api_error(e, 'Failed to authenticate')
    rescue StandardError => e
      raise Sdk::Error, "Failed to authenticate: #{e.message}"
    end

    instrument :add, :branches, :clone, :commit, :push, :pull, :status,
               :checkout_branch, :create_branch, :delete_branch,
               :init, :reset, :restore, :remote_add, :remotes, :remote_get,
               :set_config, :get_config, :configure_user, :dangerously_authenticate,
               component: 'Git'

    private

    # @return [Daytona::OtelState, nil]
    attr_reader :otel_state

    def map_api_error(api_error, prefix)
      msg = "#{prefix}: #{api_error.message}"
      case api_error.code
      when 400 then Sdk::ValidationError.new(msg)
      when 401 then Sdk::AuthenticationError.new(msg)
      when 403 then Sdk::ForbiddenError.new(msg)
      when 404 then Sdk::NotFoundError.new(msg)
      when 409 then Sdk::ConflictError.new(msg)
      when 429 then Sdk::RateLimitError.new(msg)
      when 500..599 then Sdk::ServerError.new(msg)
      else Sdk::Error.new(msg)
      end
    end
  end
end
