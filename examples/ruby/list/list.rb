# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

limit = 2
states = %w[started stopped]

page1 = daytona.list(Daytona::ListSandboxesParams.new(limit: limit, states: states))
puts 'Listing page 1'
page1.items.each do |sandbox|
  puts "#{sandbox.id}: #{sandbox.state}"
end

if page1.next_cursor
  page2 = daytona.list(Daytona::ListSandboxesParams.new(cursor: page1.next_cursor, limit: limit, states: states))
  puts 'Listing page 2'
  page2.items.each do |sandbox|
    puts "#{sandbox.id}: #{sandbox.state}"
  end
end
