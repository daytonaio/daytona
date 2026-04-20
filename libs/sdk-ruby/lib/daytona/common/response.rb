# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

module Daytona
  class PaginatedResource
    # @return [Array<Object>]
    attr_reader :items

    # @return [Float]
    attr_reader :page

    # @return [Float]
    attr_reader :total

    # @return [Float]
    attr_reader :total_pages

    # @param items [Daytona::Sandbox]
    # @param page [Float]
    # @param total [Float]
    # @param total_pages [Float]
    def initialize(items:, page:, total:, total_pages:)
      @items = items
      @page = page
      @total = total
      @total_pages = total_pages
    end
  end

  # Query parameters for filtering and sorting when listing Sandboxes.
  #
  # @example
  #   query = Daytona::ListSandboxesQuery.new(labels: { 'env' => 'prod' }, limit: 10)
  #   daytona.list(query).each { |sandbox| puts sandbox.id }
  class ListSandboxesQuery
    # @return [Integer, nil] Per-page fetch size. Does NOT limit the total number of Sandboxes returned.
    attr_accessor :limit
    # @return [String, nil] Filter by ID prefix (case-insensitive)
    attr_accessor :id
    # @return [String, nil] Filter by name prefix (case-insensitive)
    attr_accessor :name
    # @return [Hash<String, String>, nil] Filter by labels
    attr_accessor :labels
    # @return [Array<String>, nil] Filter by states
    attr_accessor :states
    # @return [Array<String>, nil] Filter by snapshot names
    attr_accessor :snapshots
    # @return [Array<String>, nil] Filter by targets
    attr_accessor :targets
    # @return [Integer, nil] Filter by minimum CPU
    attr_accessor :min_cpu
    # @return [Integer, nil] Filter by maximum CPU
    attr_accessor :max_cpu
    # @return [Integer, nil] Filter by minimum memory in GiB
    attr_accessor :min_memory_gi_b
    # @return [Integer, nil] Filter by maximum memory in GiB
    attr_accessor :max_memory_gi_b
    # @return [Integer, nil] Filter by minimum disk space in GiB
    attr_accessor :min_disk_gi_b
    # @return [Integer, nil] Filter by maximum disk space in GiB
    attr_accessor :max_disk_gi_b
    # @return [Boolean, nil] Filter by public status
    attr_accessor :is_public
    # @return [Boolean, nil] Filter by recoverable status
    attr_accessor :is_recoverable
    # @return [String, nil] Include sandboxes created after this timestamp
    attr_accessor :created_at_after
    # @return [String, nil] Include sandboxes created before this timestamp
    attr_accessor :created_at_before
    # @return [String, nil] Include sandboxes with last activity after this timestamp
    attr_accessor :last_activity_after
    # @return [String, nil] Include sandboxes with last activity before this timestamp
    attr_accessor :last_activity_before
    # @return [String, nil] Sort by field (name, cpu, memoryGiB, diskGiB, lastActivityAt, createdAt)
    attr_accessor :sort
    # @return [String, nil] Sort direction (asc, desc)
    attr_accessor :order

    def initialize(**attrs) # rubocop:disable Metrics/MethodLength
      @limit = attrs[:limit]
      @id = attrs[:id]
      @name = attrs[:name]
      @labels = attrs[:labels]
      @states = attrs[:states]
      @snapshots = attrs[:snapshots]
      @targets = attrs[:targets]
      @min_cpu = attrs[:min_cpu]
      @max_cpu = attrs[:max_cpu]
      @min_memory_gi_b = attrs[:min_memory_gi_b]
      @max_memory_gi_b = attrs[:max_memory_gi_b]
      @min_disk_gi_b = attrs[:min_disk_gi_b]
      @max_disk_gi_b = attrs[:max_disk_gi_b]
      @is_public = attrs[:is_public]
      @is_recoverable = attrs[:is_recoverable]
      @created_at_after = attrs[:created_at_after]
      @created_at_before = attrs[:created_at_before]
      @last_activity_after = attrs[:last_activity_after]
      @last_activity_before = attrs[:last_activity_before]
      @sort = attrs[:sort]
      @order = attrs[:order]
    end
  end
end
