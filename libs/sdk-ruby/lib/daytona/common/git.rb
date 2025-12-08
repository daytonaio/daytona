# frozen_string_literal: true

module Daytona
  # Response from the git commit.
  class GitCommitResponse
    # @return [String] The SHA of the commit
    attr_reader :sha

    # Initialize a new GitCommitResponse
    #
    # @param sha [String] The SHA of the commit
    def initialize(sha:)
      @sha = sha
    end
  end
end
