# frozen_string_literal: true

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
if response.exit_code != 0
  puts "Error: #{response.exit_code} #{response.result}"
else
  puts response.result
end

result = daytona.list
puts "Total sandboxes count: #{result.total.to_i}"

puts "Printing sandboxes[0] -> id: #{sandboxes.items.first.id} state: #{sandboxes.items.first.state}"

puts 'Removing sandbox'
daytona.delete(sandbox)
puts 'Sandbox removed'
