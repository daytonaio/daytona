# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

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
