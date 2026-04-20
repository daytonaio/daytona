# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

daytona.list(Daytona::ListSandboxesQuery.new(
               limit: 10,
               labels: { 'env' => 'dev' },
               states: ['started'],
               sort: 'createdAt',
               order: 'desc'
             )).each do |sandbox|
  puts sandbox.id
end
