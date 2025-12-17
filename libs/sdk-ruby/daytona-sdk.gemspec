# frozen_string_literal: true

require_relative 'lib/daytona/sdk/version'

Gem::Specification.new do |spec|
  spec.name = 'daytona-sdk'
  spec.version = Daytona::Sdk::VERSION
  spec.authors = ['Daytona Platforms Inc.']
  spec.email = ['support@daytona.io']

  spec.summary = 'Ruby SDK for Daytona'
  spec.description = 'High-level Ruby SDK for Daytona: sandboxes, git, filesystem, LSP, process, and object storage.'
  spec.homepage = 'https://github.com/daytonaio/daytona'
  spec.required_ruby_version = '>= 3.2.0'

  spec.metadata['allowed_push_host'] = 'https://rubygems.org'

  spec.metadata['homepage_uri'] = spec.homepage
  spec.metadata['source_code_uri'] = 'https://github.com/daytonaio/daytona'
  spec.metadata['changelog_uri'] = 'https://github.com/daytonaio/daytona/releases'
  spec.metadata['rubygems_mfa_required'] = 'true'

  # Specify which files should be added to the gem when it is released.
  # The `git ls-files -z` loads the files in the RubyGem that have been added into git.
  gemspec = File.basename(__FILE__)
  spec.files = IO.popen(%w[git ls-files -z], chdir: __dir__, err: IO::NULL) do |ls|
    ls.readlines("\x0", chomp: true).reject do |f|
      (f == gemspec) ||
        f.start_with?(*%w[bin/ test/ spec/ features/ .git .github appveyor Gemfile])
    end
  end
  spec.bindir = 'exe'
  spec.executables = spec.files.grep(%r{\Aexe/}) { |f| File.basename(f) }
  spec.require_paths = ['lib']

  spec.add_dependency 'aws-sdk-s3', '~> 1.0'
  spec.add_dependency 'daytona_api_client', '>= 1.0.0'
  spec.add_dependency 'daytona_toolbox_api_client', '>= 1.0.0'
  spec.add_dependency 'dotenv'
  spec.add_dependency 'toml', '~> 0.3'
  spec.add_dependency 'websocket-client-simple', '~> 0.6'
end
