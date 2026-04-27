# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Sdk do
  it 'defines Error as a StandardError' do
    expect(described_class::Error < StandardError).to be(true)
  end

  it 'defines TimeoutError as an SDK error' do
    expect(described_class::TimeoutError < described_class::Error).to be(true)
  end

  it 'memoizes the logger instance' do
    expect(described_class.logger).to be_a(Logger)
    expect(described_class.logger).to equal(described_class.logger)
  end
end
