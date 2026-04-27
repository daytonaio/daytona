# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Git do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::GitApi) }
  let(:git) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

  describe '#add' do
    it 'stages files' do
      allow(toolbox_api).to receive(:add_files)

      git.add('/repo', ['file.txt', 'src/main.rb'])

      expect(toolbox_api).to have_received(:add_files) do |req|
        expect(req.path).to eq('/repo')
        expect(req.files).to eq(['file.txt', 'src/main.rb'])
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:add_files).and_raise(StandardError, 'err')

      expect { git.add('/repo', ['f']) }.to raise_error(Daytona::Sdk::Error, /Failed to add files: err/)
    end
  end

  describe '#branches' do
    it 'lists branches' do
      response = double('BranchResponse', branches: %w[main dev])
      allow(toolbox_api).to receive(:list_branches).with('/repo').and_return(response)

      expect(git.branches('/repo')).to eq(response)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:list_branches).and_raise(StandardError, 'err')

      expect { git.branches('/repo') }.to raise_error(Daytona::Sdk::Error, /Failed to list branches: err/)
    end
  end

  describe '#clone' do
    it 'clones a repository' do
      allow(toolbox_api).to receive(:clone_repository)

      git.clone(url: 'https://github.com/user/repo.git', path: '/workspace/repo')

      expect(toolbox_api).to have_received(:clone_repository) do |req|
        expect(req.url).to eq('https://github.com/user/repo.git')
        expect(req.path).to eq('/workspace/repo')
        expect(req.branch).to be_nil
        expect(req.commit_id).to be_nil
      end
    end

    it 'passes optional branch, commit_id, and credentials' do
      allow(toolbox_api).to receive(:clone_repository)

      git.clone(
        url: 'https://github.com/user/repo.git',
        path: '/repo',
        branch: 'dev',
        commit_id: 'abc123',
        username: 'user',
        password: 'token'
      )

      expect(toolbox_api).to have_received(:clone_repository) do |req|
        expect(req.branch).to eq('dev')
        expect(req.commit_id).to eq('abc123')
        expect(req.username).to eq('user')
        expect(req.password).to eq('token')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:clone_repository).and_raise(StandardError, 'err')

      expect { git.clone(url: 'x', path: 'y') }.to raise_error(Daytona::Sdk::Error, /Failed to clone repository: err/)
    end
  end

  describe '#commit' do
    it 'creates a commit and returns GitCommitResponse' do
      commit_response = DaytonaToolboxApiClient::GitCommitResponse.new(_hash: 'abc123def')
      allow(toolbox_api).to receive(:commit_changes).and_return(commit_response)

      result = git.commit(path: '/repo', message: 'Initial', author: 'Dev', email: 'dev@example.com')

      expect(result).to be_a(Daytona::GitCommitResponse)
      expect(result.sha).to eq('abc123def')
      expect(toolbox_api).to have_received(:commit_changes) do |req|
        expect(req.path).to eq('/repo')
        expect(req.message).to eq('Initial')
      end
    end

    it 'passes allow_empty through to the API' do
      commit_response = DaytonaToolboxApiClient::GitCommitResponse.new(_hash: 'abc123def')
      allow(toolbox_api).to receive(:commit_changes).and_return(commit_response)

      git.commit(path: '/repo', message: 'Initial', author: 'Dev', email: 'dev@example.com', allow_empty: true)

      expect(toolbox_api).to have_received(:commit_changes) do |req|
        expect(req.allow_empty).to be(true)
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:commit_changes).and_raise(StandardError, 'err')

      expect { git.commit(path: '/r', message: 'm', author: 'a', email: 'e') }
        .to raise_error(Daytona::Sdk::Error, /Failed to commit changes: err/)
    end
  end

  describe '#push' do
    it 'pushes changes with optional credentials' do
      allow(toolbox_api).to receive(:push_changes)

      git.push(path: '/repo', username: 'user', password: 'token')

      expect(toolbox_api).to have_received(:push_changes) do |req|
        expect(req.path).to eq('/repo')
        expect(req.username).to eq('user')
        expect(req.password).to eq('token')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:push_changes).and_raise(StandardError, 'err')

      expect { git.push(path: '/r') }.to raise_error(Daytona::Sdk::Error, /Failed to push changes: err/)
    end
  end

  describe '#pull' do
    it 'pulls changes with optional credentials' do
      allow(toolbox_api).to receive(:pull_changes)

      git.pull(path: '/repo', username: 'user', password: 'token')

      expect(toolbox_api).to have_received(:pull_changes) do |req|
        expect(req.path).to eq('/repo')
        expect(req.username).to eq('user')
        expect(req.password).to eq('token')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:pull_changes).and_raise(StandardError, 'err')

      expect { git.pull(path: '/r') }.to raise_error(Daytona::Sdk::Error, /Failed to pull changes: err/)
    end
  end

  describe '#status' do
    it 'returns git status' do
      status_response = double('GitStatus', current_branch: 'main', ahead: 0, behind: 0)
      allow(toolbox_api).to receive(:get_status).with('/repo').and_return(status_response)

      expect(git.status('/repo')).to eq(status_response)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_status).and_raise(StandardError, 'err')

      expect { git.status('/r') }.to raise_error(Daytona::Sdk::Error, /Failed to get status: err/)
    end
  end

  describe '#checkout_branch' do
    it 'checks out a branch' do
      allow(toolbox_api).to receive(:checkout_branch)

      git.checkout_branch('/repo', 'feature')

      expect(toolbox_api).to have_received(:checkout_branch) do |req|
        expect(req.path).to eq('/repo')
        expect(req.branch).to eq('feature')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:checkout_branch).and_raise(StandardError, 'err')

      expect { git.checkout_branch('/r', 'b') }.to raise_error(Daytona::Sdk::Error, /Failed to checkout branch: err/)
    end
  end

  describe '#create_branch' do
    it 'creates a branch' do
      allow(toolbox_api).to receive(:create_branch)

      git.create_branch('/repo', 'new-feature')

      expect(toolbox_api).to have_received(:create_branch) do |req|
        expect(req.path).to eq('/repo')
        expect(req.name).to eq('new-feature')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:create_branch).and_raise(StandardError, 'err')

      expect { git.create_branch('/r', 'b') }.to raise_error(Daytona::Sdk::Error, /Failed to create branch: err/)
    end
  end

  describe '#delete_branch' do
    it 'deletes a branch' do
      allow(toolbox_api).to receive(:delete_branch)

      git.delete_branch('/repo', 'old-feature')

      expect(toolbox_api).to have_received(:delete_branch) do |req|
        expect(req.path).to eq('/repo')
        expect(req.name).to eq('old-feature')
      end
    end

    it 'wraps errors from the API call' do
      allow(toolbox_api).to receive(:delete_branch).and_raise(StandardError, 'err')

      expect { git.delete_branch('/r', 'b') }.to raise_error(Daytona::Sdk::Error, /Failed to delete branch: err/)
    end
  end
end
