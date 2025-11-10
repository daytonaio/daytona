# frozen_string_literal: true

require 'base64'

CODE = <<~PYTHON
  import matplotlib.pyplot as plt
  import numpy as np

  # Sample data
  x = np.linspace(0, 10, 30)
  y = np.sin(x)
  categories = ['A', 'B', 'C', 'D', 'E']
  values = [40, 63, 15, 25, 8]
  box_data = [np.random.normal(0, std, 100) for std in range(1, 6)]

  # 1. Line Chart
  plt.figure(figsize=(8, 5))
  plt.plot(x, y, 'b-', linewidth=2)
  plt.title('Line Chart')
  plt.xlabel('X-axis (seconds)')
  plt.ylabel('Y-axis (amplitude)')
  plt.grid(True)
  plt.show()

  # 2. Scatter Plot
  plt.figure(figsize=(8, 5))
  plt.scatter(x, y, c=y, cmap='viridis', s=100*np.abs(y))
  plt.colorbar(label='Value (normalized)')
  plt.title('Scatter Plot')
  plt.xlabel('X-axis (time in seconds)')
  plt.ylabel('Y-axis (signal strength)')
  plt.show()

  # 3. Bar Chart
  plt.figure(figsize=(10, 6))
  plt.bar(categories, values, color='skyblue', edgecolor='navy')
  plt.title('Bar Chart')
  plt.xlabel('Categories')
  plt.ylabel('Values (count)')
  plt.show()

  # 4. Pie Chart
  plt.figure(figsize=(8, 8))
  plt.pie(values, labels=categories,
          autopct='%1.1f%%',
          colors=plt.cm.Set3.colors, shadow=True, startangle=90)
  plt.title('Pie Chart (Distribution in %)')
  plt.axis('equal')
  plt.legend()
  plt.show()

  # 5. Box and Whisker Plot
  plt.figure(figsize=(10, 6))
  plt.boxplot(box_data, patch_artist=True,
              boxprops=dict(facecolor='lightblue'),
              medianprops=dict(color='red', linewidth=2))
  plt.title('Box and Whisker Plot')
  plt.xlabel('Groups (Experiment IDs)')
  plt.ylabel('Values (measurement units)')
  plt.grid(True, linestyle='--', alpha=0.7)
  plt.show()
PYTHON

def main # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
  daytona = Daytona::Daytona.new

  sandbox = daytona.create(
    Daytona::CreateSandboxFromImageParams.new(
      image: Daytona::Image.debian_slim('3.13').pip_install('matplotlib')
    ),
    on_snapshot_create_logs: proc { print _1 }
  )
  response = sandbox.process.code_run(code: CODE)

  if response.exit_code != 0
    puts "Error: #{response.exit_code} #{response.result}"
  else
    response.artifacts.charts.each do |chart|
      img_data = Base64.decode64(chart.png)
      title = chart.title || Time.now.to_i
      filename = "#{title}.png"
      File.binwrite(File.expand_path(filename, __dir__), img_data)

      puts "Image saved as #{filename}"
      print_chart(chart)
    end
  end

  daytona.delete(sandbox)
end

def print_chart(chart) # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/MethodLength
  puts "Type: #{chart.type}"
  puts "Title: #{chart.title}"

  case chart.type
  when Daytona::Charts::ChartType::LINE
    puts "X Label: #{chart.x_label}"
    puts "Y Label: #{chart.title}"
    puts "X Ticks: #{chart.x_ticks}"
    puts "X Tick Labels: #{chart.x_tick_labels}"
    puts "X Scale: #{chart.x_scale}"
    puts "Y Ticks: #{chart.y_ticks}"
    puts "Y Tick Labels: #{chart.y_tick_labels}"
    puts "Y Scale: #{chart.y_scale}"
    puts 'Elements:'
    chart.elements.each do |element|
      puts "\n  Label: #{element.label}"
      puts "  Points: #{element.points}"
    end
  when Daytona::Charts::ChartType::SCATTER
    puts "X Label: #{chart.x_label}"
    puts "Y Label: #{chart.y_label}"
    puts "X Ticks: #{chart.x_ticks}"
    puts "X Tick Labels: #{chart.x_tick_labels}"
    puts "X Scale: #{chart.x_scale}"
    puts "Y Ticks: #{chart.y_ticks}"
    puts "Y Tick Labels: #{chart.y_tick_labels}"
    puts "Y Scale: #{chart.y_scale}"
    puts 'Elements:'
    chart.elements.each do |element|
      puts "\n  Label: #{element.label}"
      puts "  Points: #{element.points}"
    end
  when Daytona::Charts::ChartType::BAR
    puts "X Label: #{chart.x_label}"
    puts "Y Label: #{chart.y_label}"
    puts 'Elements:'
    chart.elements.each do |element|
      puts "\n  Label: #{element.label}"
      puts "  Group: #{element.group}"
      puts "  Value: #{element.value}"
    end
  when Daytona::Charts::ChartType::PIE
    puts 'Elements:'
    chart.elements.each do |element|
      puts "\n  Label: #{element.label}"
      puts "  Angle: #{element.angle}"
      puts "  Radius: #{element.radius}"
      puts "  Autopct: #{element.autopct}"
    end
  when Daytona::Charts::ChartType::BOX_AND_WHISKER
    puts "X Label: #{chart.x_label}"
    puts "Y Label: #{chart.y_label}"
    puts 'Elements:'
    chart.elements.each do |element|
      puts "\n  Label: #{element.label}"
      puts "  Min: #{element.min}"
      puts '  First Quartile: {element.first_quartile}'
      puts "  Median: #{element.median}"
      puts "  Third Quartile: #{element.third_quartile}"
      puts "  Max: #{element.max}"
      puts "  Outliers: #{element.outliers}"
    end
  when Daytona::Charts::ChartType::COMPOSITE_CHART
    puts "Elements:\n"
    chart.element.each { print_chart(_1) }
  else
    raise ArgumentError, "Unknown chart: #{chart.type}"
  end
end

main
