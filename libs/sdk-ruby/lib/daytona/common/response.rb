# frozen_string_literal: true

module Daytona
  # @deprecated Use {CursorPaginatedResource} instead.
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

  # Paginated list of Sandboxes using cursor-based pagination.
  class CursorPaginatedResource
    # @return [Array<Object>]
    attr_reader :items

    # @return [String, nil] Cursor for the next page of results. Nil if there are no more results.
    attr_reader :next_cursor

    # @param items [Array<Daytona::Sandbox>]
    # @param next_cursor [String, nil]
    def initialize(items:, next_cursor:)
      @items = items
      @next_cursor = next_cursor
    end
  end
end
