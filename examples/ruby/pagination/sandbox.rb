# frozen_string_literal: true

require 'daytona'

daytona = Daytona::Daytona.new

daytona.list(Daytona::ListSandboxesQuery.new(
               limit: 10,
               labels: { 'env' => 'dev' },
               states: [Daytona::SandboxState::STARTED],
               sort: Daytona::SandboxListSortField::CREATED_AT,
               order: Daytona::SandboxListSortDirection::DESC
             )).each do |sandbox|
  puts sandbox.id
end
