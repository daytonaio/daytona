# frozen_string_literal: true

daytona = Daytona::Daytona.new

result = daytona.snapshot.list(page: 2, limit: 10)
result.items.each do |snapshot|
  puts "#{snapshot.name} (#{snapshot.image_name})"
end
