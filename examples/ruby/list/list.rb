# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

states_filter = %w[started stopped]

page1 = daytona.list_v2(limit: 2, states: states_filter)
page1.items.each do |sandbox|
  puts "#{sandbox.id}: #{sandbox.state}"
end

if page1.next_cursor
  page2 = daytona.list_v2(cursor: page1.next_cursor, limit: 2, states: states_filter)
  page2.items.each do |sandbox|
    puts "#{sandbox.id}: #{sandbox.state}"
  end
end
