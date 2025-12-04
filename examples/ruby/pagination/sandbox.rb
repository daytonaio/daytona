# frozen_string_literal: true

daytona = Daytona::Daytona.new

result = daytona.list({ 'my-label' => 'my-value' }, page: 2, limit: 10)
result.items.each do |sandbox|
  puts "#{sandbox.id} (#{sandbox.state})"
end
