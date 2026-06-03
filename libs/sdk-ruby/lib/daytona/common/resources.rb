# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

module Daytona
  class Resources
    # @return [Integer, nil] Number of CPU cores to allocate
    attr_reader :cpu

    # @return [Integer, nil] Amount of memory in GiB to allocate
    attr_reader :memory

    # @return [Integer, nil] Amount of disk space in GiB to allocate
    attr_reader :disk

    # @return [Integer, nil] Number of GPUs to allocate
    attr_reader :gpu

    # @return [String, Array<String>, nil] Preferred GPU type for the Sandbox
    attr_reader :gpu_type

    # @param cpu [Integer, nil] Number of CPU cores to allocate
    # @param memory [Integer, nil] Amount of memory in GiB to allocate
    # @param disk [Integer, nil] Amount of disk space in GiB to allocate
    # @param gpu [Integer, nil] Number of GPUs to allocate
    # @param gpu_type [String, Array<String>, nil] Preferred GPU type for the Sandbox
    def initialize(cpu: nil, memory: nil, disk: nil, gpu: nil, gpu_type: nil)
      @cpu = cpu
      @memory = memory
      @disk = disk
      @gpu = gpu
      @gpu_type = gpu_type
    end

    # @return [Hash] Hash representation of the resources
    def to_h = { cpu:, memory:, disk:, gpu:, gpu_type: }.compact
  end
end
