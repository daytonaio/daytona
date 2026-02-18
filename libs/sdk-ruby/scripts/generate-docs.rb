#!/usr/bin/env ruby
# frozen_string_literal: true

require 'fileutils'
require 'tmpdir'
require 'yard'

# Configuration
DOCS_OUTPUT_DIR = File.expand_path('../../../apps/docs/src/content/docs/en/ruby-sdk', __dir__)
LIB_DIR = File.expand_path('../lib/daytona', __dir__)

# Classes to document: [file_path, output_filename, class_name]
CLASSES_TO_DOCUMENT = [
  ['config.rb', 'config.mdx', 'Daytona::Config'],
  ['daytona.rb', 'daytona.mdx', 'Daytona::Daytona'],
  ['sandbox.rb', 'sandbox.mdx', 'Daytona::Sandbox'],
  ['file_system.rb', 'file-system.mdx', 'Daytona::FileSystem'],
  ['git.rb', 'git.mdx', 'Daytona::Git'],
  ['process.rb', 'process.mdx', 'Daytona::Process'],
  ['lsp_server.rb', 'lsp-server.mdx', 'Daytona::LspServer'],
  ['volume.rb', 'volume.mdx', 'Daytona::Volume'],
  ['object_storage.rb', 'object-storage.mdx', 'Daytona::ObjectStorage'],
  ['computer_use.rb', 'computer-use.mdx', 'Daytona::ComputerUse'],
  ['snapshot_service.rb', 'snapshot.mdx', 'Daytona::SnapshotService'],
  ['volume_service.rb', 'volume-service.mdx', 'Daytona::VolumeService'],
  ['common/charts.rb', 'charts.mdx', 'Daytona::Chart'],
  ['common/image.rb', 'image.mdx', 'Daytona::Image']
]

def extract_class_name_from_path(class_name)
  # Extract the simple class name from the full path
  class_name.split('::').last
end

def add_frontmatter(content, class_name)
  simple_name = extract_class_name_from_path(class_name)
  frontmatter = <<~FRONTMATTER
    ---
    title: "#{simple_name}"
    hideTitleOnPage: true
    ---

  FRONTMATTER

  frontmatter + content
end

def format_type(types)
  return 'Object' if types.nil? || types.empty?

  # Join types and escape special characters that break MDX
  type_str = types.join(', ')
  # Escape special chars that break MDX parsing
  # Replace :: with a single : to avoid MDX issues while keeping readability
  type_str = type_str.gsub('::', ':')
  # Escape angle brackets
  type_str.gsub('<', '\\<').gsub('>', '\\>')
end

def clean_description(description)
  return '' if description.nil? || description.empty?

  # Remove rubocop directive lines
  cleaned = description.to_s.lines.reject { |line| line.strip.start_with?('rubocop:') }.join
  # Convert YARD cross-reference links ({ClassName}) to backtick-quoted names so
  # they are not interpreted as JSX expressions by MDX.
  cleaned = cleaned.gsub(/\{([^}]+)\}/) { "`#{Regexp.last_match(1)}`" }
  # Convert YARD inline-code markers (+text+) to Markdown backticks.
  cleaned = cleaned.gsub(/\+([^+]+)\+/) { "`#{Regexp.last_match(1)}`" }
  cleaned.strip
end

def extract_class_description(obj)
  description = clean_description(obj.docstring)

  # If no class-level description, try to get it from the first method or constructor
  if description.empty? && obj.is_a?(YARD::CodeObjects::ClassObject)
    # Try to find a description from the constructor or first documented method
    constructor = obj.meths.find { |m| m.name == :initialize }
    if constructor && constructor.docstring && !constructor.docstring.empty?
      # Extract just the summary (first paragraph) from constructor docs
      constructor_desc = clean_description(constructor.docstring)
      first_paragraph = constructor_desc.split("\n\n").first
      if first_paragraph && !first_paragraph.empty?
        # Make it a class-level description
        extract_class_name_from_path(obj.path)
        description = first_paragraph.gsub(/^(Initializes|Creates a new)/, 'Main class for')
      end
    end

    # If still empty, generate a basic description
    if description.empty?
      simple_name = extract_class_name_from_path(obj.path)
      description = "#{simple_name} class for Daytona SDK."
    end
  end

  description
end

def extract_attributes(obj)
  attributes = []

  return attributes unless obj.is_a?(YARD::CodeObjects::ClassObject)

  # Collect all attributes
  read_attrs = obj.attributes[:read] || {}
  write_attrs = obj.attributes[:write] || {}
  all_attrs = (read_attrs.keys + write_attrs.keys).uniq

  all_attrs.each do |name|
    attr_obj = read_attrs[name] || write_attrs[name]
    type = attr_obj.docstring.tag(:return)&.types
    type_str = format_type(type)
    desc = attr_obj.docstring.to_s.split("\n").first || ''

    attributes << {
      name: name,
      type: type_str,
      description: desc
    }
  end

  attributes
end

def generate_markdown_for_object(obj)
  content = []

  # Add main heading
  content << "## #{obj.name}"
  content << ''

  # Add class description
  description = extract_class_description(obj)
  unless description.empty?
    content << description
    content << ''
  end

  # Add attributes/properties section (matching Python/TypeScript format)
  if obj.is_a?(YARD::CodeObjects::ClassObject)
    attributes = extract_attributes(obj)

    if attributes.any?
      content << '**Attributes**:'
      content << ''
      attributes.each do |attr|
        content << "- `#{attr[:name]}` _#{attr[:type]}_ - #{attr[:description]}"
      end
      content << ''
    end
  end

  # Add class-level examples (before constructors, matching Python/TypeScript)
  examples = obj.tags(:example)
  if examples.any?
    content << '**Examples:**'
    content << ''
    examples.each do |example|
      content << '```ruby'
      content << example.text.strip
      content << '```'
      content << ''
    end
  end

  # Add constructors section
  if obj.is_a?(YARD::CodeObjects::ClassObject)
    constructor = obj.meths.find { |m| m.name == :initialize }
    if constructor
      content << '### Constructors'
      content << ''
      content << "#### new #{extract_class_name_from_path(obj.path)}()"
      content << ''

      # Method signature
      content << '```ruby'
      params_str = constructor.parameters.map { |p| p[0] }.join(', ')
      content << "def initialize(#{params_str})"
      content << '```'
      content << ''

      # Constructor description
      if constructor.docstring && !constructor.docstring.empty?
        desc = clean_description(constructor.docstring)
        unless desc.empty?
          content << desc
          content << ''
        end
      end

      # Parameters
      params = constructor.tags(:param)
      if params.any?
        content << '**Parameters**:'
        content << ''
        params.each do |param|
          types = format_type(param.types)
          content << "- `#{param.name}` _#{types}_ - #{param.text}"
        end
        content << ''
      end

      # Returns
      return_tag = constructor.tag(:return)
      if return_tag
        types = format_type(return_tag.types)
        text = return_tag.text.to_s.strip
        content << '**Returns**:'
        content << ''
        content << if text.empty?
                     "- `#{types}`"
                   else
                     "- `#{types}` - #{text}"
                   end
        content << ''
      end

      # Raises
      raises = constructor.tags(:raise)
      if raises.any?
        content << '**Raises**:'
        content << ''
        raises.each do |raise_tag|
          types = format_type(raise_tag.types)
          content << "- `#{types}` - #{raise_tag.text}"
        end
        content << ''
      end
    end
  end

  # Add methods section
  if obj.is_a?(YARD::CodeObjects::ClassObject)
    methods = obj.meths.select { |m| m.scope == :instance && m.visibility == :public && m.name != :initialize }

    if methods.any?
      content << '### Methods'
      content << ''

      methods.each do |method|
        content << "#### #{method.name}()"
        content << ''

        overloads = method.tags(:overload)

        if overloads.any?
          # General method description (shared across all overloads)
          if method.docstring && !method.docstring.empty?
            desc = clean_description(method.docstring)
            unless desc.empty?
              content << desc
              content << ''
            end
          end

          overloads.each do |overload|
            # Overload signature
            content << '```ruby'
            content << "def #{overload.name}"
            content << '```'
            content << ''

            # Overload-specific description
            if overload.docstring && !overload.docstring.to_s.empty?
              desc = clean_description(overload.docstring)
              unless desc.empty?
                content << desc
                content << ''
              end
            end

            # Deprecated notice
            deprecated = overload.tag(:deprecated)
            if deprecated
              dep_text = deprecated.text.to_s.strip
              content << if dep_text.empty?
                           '**Deprecated**'
                         else
                           "**Deprecated**: #{dep_text}"
                         end
              content << ''
            end

            # Parameters
            params = overload.tags(:param)
            if params.any?
              content << '**Parameters**:'
              content << ''
              params.each do |param|
                types = format_type(param.types)
                content << "- `#{param.name}` _#{types}_ - #{param.text}"
              end
              content << ''
            end

            # Returns
            return_tag = overload.tag(:return)
            if return_tag
              types = format_type(return_tag.types)
              text = return_tag.text.to_s.strip
              content << '**Returns**:'
              content << ''
              content << if text.empty? || text.start_with?('Array<') || text.start_with?('Hash<')
                           "- `#{types}`"
                         else
                           "- `#{types}` - #{text}"
                         end
              content << ''
            end

            # Raises
            raises = overload.tags(:raise)
            next unless raises.any?

            content << '**Raises**:'
            content << ''
            raises.each do |raise_tag|
              types = format_type(raise_tag.types)
              content << "- `#{types}` - #{raise_tag.text}"
            end
            content << ''
          end
        else
          # Method signature
          content << '```ruby'
          params_str = method.parameters.map { |p| p[0] }.join(', ')
          content << "def #{method.name}(#{params_str})"
          content << '```'
          content << ''

          # Method description
          if method.docstring && !method.docstring.empty?
            desc = clean_description(method.docstring)
            unless desc.empty?
              content << desc
              content << ''
            end
          end

          # Parameters
          params = method.tags(:param)
          if params.any?
            content << '**Parameters**:'
            content << ''
            params.each do |param|
              types = format_type(param.types)
              content << "- `#{param.name}` _#{types}_ - #{param.text}"
            end
            content << ''
          end

          # Returns
          return_tag = method.tag(:return)
          if return_tag
            types = format_type(return_tag.types)
            text = return_tag.text.to_s.strip
            content << '**Returns**:'
            content << ''
            # Only add description if it's meaningful and not just repeating the type
            content << if text.empty? || text.start_with?('Array<') || text.start_with?('Hash<')
                         "- `#{types}`"
                       else
                         "- `#{types}` - #{text}"
                       end
            content << ''
          end

          # Raises
          raises = method.tags(:raise)
          if raises.any?
            content << '**Raises**:'
            content << ''
            raises.each do |raise_tag|
              types = format_type(raise_tag.types)
              content << "- `#{types}` - #{raise_tag.text}"
            end
            content << ''
          end
        end

        # Method-level examples
        examples = method.tags(:example)
        next unless examples.any?

        content << '**Examples:**'
        content << ''
        examples.each do |example|
          content << '```ruby'
          content << example.text.strip
          content << '```'
          content << ''
        end
      end
    end
  end

  content.join("\n")
end

def post_process_markdown(content)
  # Remove excessive blank lines (more than 2 consecutive)
  content = content.gsub(/\n{3,}/, "\n\n")

  # Remove blank lines inside code blocks
  content = content.gsub("```ruby\n\n", "```ruby\n")
  content = content.gsub("\n\n```", "\n```")

  # Ensure consistent spacing around sections
  content = content.gsub(/\n(\*\*[^*]+\*\*:)\n([^\n])/, "\n\\1\n\n\\2")

  # Ensure code blocks have proper spacing
  content = content.gsub(/([^\n])\n```/, "\\1\n\n```")
  content = content.gsub(/```\n([^\n])/, "```\n\n\\1")

  # Clean up trailing whitespace
  content = content.lines.map(&:rstrip).join("\n")

  # Ensure file ends with single newline
  content.strip + "\n"
end

def generate_docs_for_class(file_path, output_filename, class_name)
  full_path = File.join(LIB_DIR, file_path)

  unless File.exist?(full_path)
    puts "⚠️  File not found: #{full_path}"
    return
  end

  puts "📝 Generating docs for #{class_name}..."

  begin
    # Clear and parse with YARD
    YARD::Registry.clear
    YARD::Parser::SourceParser.parse(full_path)

    # Get the class object
    obj = YARD::Registry.at(class_name)

    unless obj
      puts "⚠️  Class #{class_name} not found in registry"
      return
    end

    # Generate markdown content
    markdown_content = generate_markdown_for_object(obj)

    # Post-process for consistency
    markdown_content = post_process_markdown(markdown_content)

    # Add frontmatter
    final_content = add_frontmatter(markdown_content, class_name)

    # Write to output file
    output_path = File.join(DOCS_OUTPUT_DIR, output_filename)
    File.write(output_path, final_content)

    puts "✅ Generated: #{output_filename}"
  rescue StandardError => e
    puts "❌ Error generating docs for #{class_name}: #{e.message}"
    puts e.backtrace.first(5).join("\n") if ENV['DEBUG']
  end
end

# Main execution
puts '🚀 Starting documentation generation...'
puts "📂 Output directory: #{DOCS_OUTPUT_DIR}"
puts ''

# Ensure output directory exists
FileUtils.mkdir_p(DOCS_OUTPUT_DIR)

# Generate docs for each class
CLASSES_TO_DOCUMENT.each do |file_path, output_filename, class_name|
  generate_docs_for_class(file_path, output_filename, class_name)
end

puts ''
puts '✨ Documentation generation complete!'
