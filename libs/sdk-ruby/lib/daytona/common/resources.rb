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

    # @param cpu [Integer, nil] Number of CPU cores to allocate
    # @param memory [Integer, nil] Amount of memory in GiB to allocate
    # @param disk [Integer, nil] Amount of disk space in GiB to allocate
    # @param gpu [Integer, nil] Number of GPUs to allocate
    def initialize(cpu: nil, memory: nil, disk: nil, gpu: nil)
      @cpu = cpu
      @memory = memory
      @disk = disk
      @gpu = gpu
    end

    # @return [Hash] Hash representation of the resources
    def to_h = { cpu:, memory:, disk:, gpu: }.compact
  end
end
