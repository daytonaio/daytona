# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

limit = 2
states = %w[started stopped]

page1 = daytona.list(ListSandboxesParams(limit: limit, states: states))
page1.items.each do |sandbox|
  puts "#{sandbox.id}: #{sandbox.state}"
end

if page1.next_cursor
  page2 = daytona.list(ListSandboxesParams(cursor: page1.next_cursor, limit: limit, states: states))
  page2.items.each do |sandbox|
    puts "#{sandbox.id}: #{sandbox.state}"
  end
end
