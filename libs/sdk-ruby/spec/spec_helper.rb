# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

require 'logger'
require 'webmock/rspec'
require 'daytona'

WebMock.disable_net_connect!(allow_localhost: true)

RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  config.shared_context_metadata_behavior = :apply_to_host_groups
  config.order = :random
  config.filter_run_when_matching :focus

  # Silence SDK logger during tests
  config.before(:suite) do
    Daytona::Sdk.logger.level = Logger::FATAL
  end
end

# ---------------------------------------------------------------------------
# Shared helpers for building mock DTOs
# ---------------------------------------------------------------------------

def build_sandbox_dto(overrides = {}) # rubocop:disable Metrics/MethodLength
  attrs = {
    id: 'sandbox-123',
    organization_id: 'org-1',
    snapshot: 'default-snapshot',
    user: 'daytona',
    env: {},
    labels: { 'code-toolbox-language' => 'python' },
    public: false,
    target: 'us',
    cpu: 4,
    gpu: 0,
    memory: 8,
    disk: 30,
    state: 'started',
    desired_state: 'started',
    error_reason: nil,
    backup_state: nil,
    backup_created_at: nil,
    auto_stop_interval: 15,
    auto_archive_interval: 10_080,
    auto_delete_interval: -1,
    volumes: [],
    build_info: nil,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    last_activity_at: '2025-01-01T00:00:00Z',
    daemon_version: '1.0.0',
    network_block_all: false,
    network_allow_list: nil,
    toolbox_proxy_url: 'https://proxy.example.com/'
  }.merge(overrides)

  instance_double(DaytonaApiClient::Sandbox, **attrs)
end

def build_volume_dto(overrides = {})
  attrs = {
    id: 'vol-123',
    name: 'test-volume',
    organization_id: 'org-1',
    state: 'ready',
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    last_used_at: '2025-01-01T00:00:00Z',
    error_reason: nil
  }.merge(overrides)

  instance_double(DaytonaApiClient::VolumeDto, **attrs)
end

def build_snapshot_dto(overrides = {})
  attrs = {
    id: 'snap-123',
    organization_id: 'org-1',
    general: false,
    name: 'test-snapshot',
    image_name: 'ubuntu:22.04',
    state: 'active',
    size: 1024,
    entrypoint: nil,
    cpu: 4,
    gpu: 0,
    mem: 8,
    disk: 30,
    error_reason: nil,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    last_used_at: nil,
    build_info: nil
  }.merge(overrides)

  instance_double(DaytonaApiClient::SnapshotDto, **attrs)
end

def build_config(overrides = {})
  attrs = {
    api_key: 'test-api-key',
    api_url: 'https://api.example.com',
    target: 'us'
  }.merge(overrides)

  Daytona::Config.new(**attrs)
end
