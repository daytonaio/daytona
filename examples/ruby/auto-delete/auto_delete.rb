# frozen_string_literal: true

daytona = Daytona::Daytona.new

# Auto delete disabled by default
first_sandbox = daytona.create
puts "Default auto delete interval: #{first_sandbox.auto_delete_interval}"

# Auto delete after the Sandbox has been stopped for 1 hour
first_sandbox.auto_delete_interval = 60
puts "Auto delete interval: #{first_sandbox.auto_delete_interval}"

# Delete immediately upon stopping
first_sandbox.auto_delete_interval = 0
puts "Auto delete interval: #{first_sandbox.auto_delete_interval}"

# Disable auto delete
first_sandbox.auto_delete_interval = -1
puts "Auto delete interval: #{first_sandbox.auto_delete_interval}"

# Auto delete after the Sandbox has been stopped for 1 day
second_sandbox = daytona.create(Daytona::CreateSandboxFromSnapshotParams.new(auto_delete_interval: 24 * 60))
puts "Auto delete interval: #{second_sandbox.auto_delete_interval}"

daytona.delete(first_sandbox)
daytona.delete(second_sandbox)
