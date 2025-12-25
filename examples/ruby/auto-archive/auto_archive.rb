# frozen_string_literal: true

daytona = Daytona::Daytona.new

# Default interval
first_sandbox = daytona.create
puts "Default auto archive interval: #{first_sandbox.auto_archive_interval}"

# Set interval to 1 hour
first_sandbox.auto_archive_interval = 60
puts "Auto archive interval: #{first_sandbox.auto_archive_interval}"

# Max interval
second_sandbox = daytona.create(Daytona::CreateSandboxFromSnapshotParams.new(auto_archive_interval: 0))
puts "Max auto archive interval: #{second_sandbox.auto_archive_interval}"

# 1 day interval
third_sandbox = daytona.create(Daytona::CreateSandboxFromSnapshotParams.new(auto_archive_interval: 24 * 60))
puts "Auto archive interval: #{third_sandbox.auto_archive_interval}"

daytona.delete(first_sandbox)
daytona.delete(second_sandbox)
daytona.delete(third_sandbox)
