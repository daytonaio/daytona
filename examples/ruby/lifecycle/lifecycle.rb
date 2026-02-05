# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

puts 'Creating sandbox'
sandbox = daytona.create
puts 'Sandbox created'

puts 'Replacing sandbox labels'
sandbox.labels = { public: true }
puts "Sandbox labels: #{sandbox.labels}"

puts 'Stopping sandbox'
daytona.stop(sandbox)
puts "Sandbox #{sandbox.state}"

puts 'Starting sandbox'
daytona.start(sandbox)
puts "Sandbox #{sandbox.state}"

puts 'Getting existing sandbox'
sandbox = daytona.get(sandbox.id)
puts 'Retrieved existing sandbox'

response = sandbox.process.exec(command: 'echo "Hello World from exec!"', cwd: '/home/daytona', timeout: 10)
if response.exit_code == 0
  puts response.result
else
  puts "Error: #{response.exit_code} #{response.result}"
end

result = daytona.list
puts "Total sandboxes count: #{result.total.to_i}"

puts "Printing sandboxes[0] -> id: #{result.items.first.id} state: #{result.items.first.state}"

# Resize a started sandbox (CPU and memory can be increased)
puts 'Resizing started sandbox...'
sandbox.resize(Daytona::Resources.new(cpu: 2, memory: 2))
puts "Resize complete: CPU=#{sandbox.cpu}, Memory=#{sandbox.memory}GB"

# Resize a stopped sandbox (CPU, memory, and disk can be changed)
puts 'Stopping sandbox for resize...'
sandbox.stop
puts 'Resizing stopped sandbox...'
sandbox.resize(Daytona::Resources.new(cpu: 4, memory: 4, disk: 20))
puts "Resize complete: CPU=#{sandbox.cpu}, Memory=#{sandbox.memory}GB, Disk=#{sandbox.disk}GB"
sandbox.start
puts 'Sandbox restarted with new resources'

puts 'Removing sandbox'
daytona.delete(sandbox)
puts 'Sandbox removed'
