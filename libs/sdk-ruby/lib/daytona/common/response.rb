# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

module Daytona
  # Re-export of api-client enum constants under the Daytona namespace so
  # SDK consumers never need to import from DaytonaApiClient directly.
  SandboxClass = DaytonaApiClient::SandboxClass
  SandboxState = DaytonaApiClient::SandboxState
  SandboxListSortField = DaytonaApiClient::SandboxListSortField
  SandboxListSortDirection = DaytonaApiClient::SandboxListSortDirection
  GpuType = DaytonaApiClient::GpuType

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
  # All fields are optional and default to +nil+. Constructed via keyword
  # arguments and immutable (Ruby 3.2+ Data semantics).
  #
  # @example
  #   query = Daytona::ListSandboxesQuery.new(labels: { 'env' => 'prod' }, limit: 10)
  #   daytona.list(query).each { |sandbox| puts sandbox.id }
  ListSandboxesQuery = Data.define(
    :limit,
    :id,
    :name,
    :labels,
    :states,
    :snapshots,
    :targets,
    :min_cpu,
    :max_cpu,
    :min_memory_gib,
    :max_memory_gib,
    :min_disk_gib,
    :max_disk_gib,
    :is_public,
    :is_recoverable,
    :created_at_after,
    :created_at_before,
    :last_activity_after,
    :last_activity_before,
    :sort,
    :order
  ) do
    # All members default to nil so callers pass only the filters they care about.
    DEFAULTS = members.to_h { |m| [m, nil] }.freeze

    class << self
      alias_method :_data_new, :new
      def new(**attrs) = _data_new(**DEFAULTS, **attrs)
    end

    # Idiomatic Ruby boolean predicate aliases.
    def public? = is_public
    def recoverable? = is_recoverable
  end
end
