# frozen_string_literal: true

module Daytona
  class Git
    # @return [String] The Sandbox ID
    attr_reader :sandbox_id

    # @return [DaytonaToolboxApiClient::GitApi] API client for Sandbox operations
    attr_reader :toolbox_api

    # Initializes a new Git handler instance.
    #
    # @param sandbox_id [String] The Sandbox ID.
    # @param toolbox_api [DaytonaToolboxApiClient::GitApi] API client for Sandbox operations.
    def initialize(sandbox_id:, toolbox_api:)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
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
      toolbox_api.git_add_files(DaytonaToolboxApiClient::GitAddRequest.new(path:, files:))
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
      toolbox_api.git_list_branches(path)
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
    def clone(url:, path:, branch: nil, commit_id: nil, username: nil, password: nil) # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists
      toolbox_api.git_clone_repository(
        DaytonaToolboxApiClient::GitCloneRequest.new(
          url: url,
          branch: branch,
          path: path,
          username: username,
          password: password,
          commit_id: commit_id
        )
      )
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
      response = toolbox_api.git_commit_changes(
        DaytonaToolboxApiClient::GitCommitRequest.new(path:, message:, author:, email:, allow_empty:)
      )
      GitCommitResponse.new(sha: response.hash)
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
    def push(path:, username: nil, password: nil)
      toolbox_api.git_push_changes(
        DaytonaToolboxApiClient::GitRepoRequest.new(path:, username:, password:)
      )
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
    def pull(path:, username: nil, password: nil)
      toolbox_api.git_pull_changes(
        DaytonaToolboxApiClient::GitRepoRequest.new(path:, username:, password:)
      )
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
      toolbox_api.git_get_status(path)
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
      toolbox_api.git_checkout_branch(
        DaytonaToolboxApiClient::GitCheckoutRequest.new(path:, branch:)
      )
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
      toolbox_api.git_create_branch(
        DaytonaToolboxApiClient::GitBranchRequest.new(path:, name:)
      )
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
      toolbox_api.git_delete_branch(
        DaytonaToolboxApiClient::GitDeleteBranchRequest.new(path:, name:)
      )
    rescue StandardError => e
      raise Sdk::Error, "Failed to delete branch: #{e.message}"
    end
  end
end
