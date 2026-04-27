# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Image do
  describe Daytona::Context do
    it 'stores the source and archive paths' do
      context = described_class.new(source_path: '/tmp/src', archive_path: 'archive/src')

      expect(context.source_path).to eq('/tmp/src')
      expect(context.archive_path).to eq('archive/src')
    end
  end

  describe '#initialize' do
    it 'defaults dockerfile to an empty string and context_list to an empty array' do
      image = described_class.new

      expect(image.dockerfile).to eq('')
      expect(image.context_list).to eq([])
    end
  end

  describe '#pip_install' do
    it 'returns self when no packages are provided' do
      image = described_class.base('python:3.12')

      expect(image.pip_install).to equal(image)
      expect(image.dockerfile).to eq("FROM python:3.12\n")
    end

    it 'sorts and flattens package arguments' do
      image = described_class.base('python:3.12')

      image.pip_install('requests', %w[numpy pandas])

      expect(image.dockerfile).to include('RUN python -m pip install numpy pandas requests')
    end

    it 'formats optional pip arguments' do
      image = described_class.base('python:3.12')

      image.pip_install(
        'requests',
        find_links: ['https://a.example.com'],
        index_url: 'https://pypi.example.com/simple',
        extra_index_urls: ['https://extra.example.com/simple'],
        pre: true,
        extra_options: ' --no-cache-dir '
      )

      expect(image.dockerfile).to include('--find-links https://a.example.com')
      expect(image.dockerfile).to include('--index-url https://pypi.example.com/simple')
      expect(image.dockerfile).to include('--extra-index-url https://extra.example.com/simple')
      expect(image.dockerfile).to include('--pre')
      expect(image.dockerfile).to include('--no-cache-dir')
    end

    it 'raises when non-string package values are provided' do
      image = described_class.base('python:3.12')

      expect { image.pip_install('requests', [1]) }
        .to raise_error(Daytona::Sdk::Error, /pip_install: packages must only contain strings/)
    end
  end

  describe '#pip_install_from_requirements' do
    it 'raises when the requirements file does not exist' do
      image = described_class.base('python:3.12')

      expect { image.pip_install_from_requirements('/missing/requirements.txt') }
        .to raise_error(Daytona::Sdk::Error, /does not exist/)
    end

    it 'adds the requirements file to context and dockerfile' do
      Dir.mktmpdir do |dir|
        requirements = File.join(dir, 'requirements.txt')
        File.write(requirements, "requests\n")

        image = described_class.base('python:3.12')
        image.pip_install_from_requirements(requirements, pre: true)

        expect(image.context_list.length).to eq(1)
        expect(image.context_list.first.source_path).to eq(File.expand_path(requirements))
        expect(image.dockerfile).to include('COPY requirements.txt /.requirements.txt')
        expect(image.dockerfile).to include('RUN python -m pip install -r /.requirements.txt --pre')
      end
    end
  end

  describe '#pip_install_from_pyproject' do
    it 'raises when project dependencies are missing' do
      Dir.mktmpdir do |dir|
        pyproject = File.join(dir, 'pyproject.toml')
        File.write(pyproject, "[project]\nname = \"demo\"\n")

        image = described_class.base('python:3.12')

        expect { image.pip_install_from_pyproject(pyproject) }
          .to raise_error(Daytona::Sdk::Error, /No \[project.dependencies\] section/)
      end
    end

    it 'installs project and optional dependencies' do
      Dir.mktmpdir do |dir|
        pyproject = File.join(dir, 'pyproject.toml')
        File.write(pyproject, <<~TOML)
          [project]
          name = "demo"
          dependencies = ["requests", "flask"]

          [project.optional-dependencies]
          dev = ["pytest"]
        TOML

        image = described_class.base('python:3.12')
        image.pip_install_from_pyproject(pyproject, optional_dependencies: ['dev'])

        expect(image.dockerfile).to include('RUN python -m pip install flask pytest requests')
      end
    end

    it 'returns nil when optional_dependencies is nil' do
      Dir.mktmpdir do |dir|
        pyproject = File.join(dir, 'pyproject.toml')
        File.write(pyproject, <<~TOML)
          [project]
          name = "demo"
          dependencies = ["requests"]
        TOML

        image = described_class.base('python:3.12')

        expect(image.pip_install_from_pyproject(pyproject, optional_dependencies: nil)).to be_nil
      end
    end
  end

  describe '#add_local_file' do
    it 'adds a local file and appends the basename when remote path ends with slash' do
      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'config.json')
        File.write(file_path, '{}')

        image = described_class.base('python:3.12')
        image.add_local_file(file_path, '/app/')

        expect(image.context_list.first.source_path).to eq(File.expand_path(file_path))
        expect(image.dockerfile).to include('COPY config.json /app//config.json')
      end
    end
  end

  describe '#add_local_dir' do
    it 'adds a local directory to the build context' do
      Dir.mktmpdir do |dir|
        nested = File.join(dir, 'src')
        Dir.mkdir(nested)

        image = described_class.base('python:3.12')
        image.add_local_dir(nested, '/workspace/src')

        expect(image.context_list.first.source_path).to eq(File.expand_path(nested))
        expect(image.dockerfile).to include('COPY src /workspace/src')
      end
    end
  end

  describe '#run_commands' do
    it 'adds string commands directly' do
      image = described_class.base('python:3.12')

      image.run_commands('echo hello', 'pwd')

      expect(image.dockerfile).to include("RUN echo hello\nRUN pwd\n")
    end

    it 'quotes array commands' do
      image = described_class.base('python:3.12')

      image.run_commands(%w[echo hello])

      expect(image.dockerfile).to include('RUN "echo" "hello"')
    end
  end

  describe '#env' do
    it 'appends environment variables to the dockerfile' do
      image = described_class.base('python:3.12')

      image.env('APP_ENV' => 'production', 'GREETING' => 'hello world')

      expect(image.dockerfile).to include('ENV APP_ENV=production')
      expect(image.dockerfile).to include('ENV GREETING=hello\ world')
    end

    it 'raises when a value is not a string' do
      image = described_class.base('python:3.12')

      expect { image.env('PORT' => 3000) }
        .to raise_error(Daytona::Sdk::Error, /Image ENV variables must be strings/)
    end
  end

  describe '#workdir' do
    it 'escapes the working directory path' do
      image = described_class.base('python:3.12')

      image.workdir('/workspace/my app')

      expect(image.dockerfile).to include('WORKDIR /workspace/my\ app')
    end
  end

  describe '#entrypoint' do
    it 'stores the entrypoint as a JSON array' do
      image = described_class.base('python:3.12')

      image.entrypoint(['/bin/bash', '-lc'])

      expect(image.dockerfile).to include('ENTRYPOINT ["/bin/bash", "-lc"]')
    end

    it 'raises for invalid entrypoint arguments' do
      image = described_class.base('python:3.12')

      expect { image.entrypoint('/bin/bash') }
        .to raise_error(Daytona::Sdk::Error, /entrypoint_commands must be a list of strings/)
    end
  end

  describe '#cmd' do
    it 'stores the command as a JSON array' do
      image = described_class.base('python:3.12')

      image.cmd(['ruby', 'app.rb'])

      expect(image.dockerfile).to include('CMD ["ruby", "app.rb"]')
    end

    it 'raises for invalid command arguments' do
      image = described_class.base('python:3.12')

      expect { image.cmd('ruby app.rb') }
        .to raise_error(Daytona::Sdk::Error, /Image CMD must be a list of strings/)
    end
  end

  describe '#dockerfile_commands' do
    it 'raises when the context directory does not exist' do
      image = described_class.base('python:3.12')

      expect { image.dockerfile_commands(['COPY . /app'], context_dir: '/missing/context') }
        .to raise_error(Daytona::Sdk::Error, /Context directory .* does not exist/)
    end

    it 'extracts COPY sources from commands into context_list' do
      Dir.mktmpdir do |dir|
        File.write(File.join(dir, 'Gemfile'), "source 'https://rubygems.org'\n")
        File.write(File.join(dir, 'Gemfile.lock'), '')

        image = described_class.base('python:3.12')
        image.dockerfile_commands(['COPY Gemfile* /app/'], context_dir: dir)

        expect(image.context_list.length).to eq(2)
        source_paths = image.context_list.map(&:source_path)
        expect(source_paths).to include(File.join(dir, 'Gemfile'))
        expect(source_paths).to include(File.join(dir, 'Gemfile.lock'))
        expect(image.dockerfile).to include('COPY Gemfile* /app/')
      end
    end
  end

  describe '.from_dockerfile' do
    it 'builds an image and extracts copy sources' do
      Dir.mktmpdir do |dir|
        src = File.join(dir, 'src')
        Dir.mkdir(src)
        File.write(File.join(src, 'app.rb'), 'puts :ok')
        dockerfile = File.join(dir, 'Dockerfile')
        File.write(dockerfile, "FROM ruby:3.4\nCOPY src/app.rb /app/app.rb\n")

        image = described_class.from_dockerfile(dockerfile)

        expect(image.dockerfile).to include('FROM ruby:3.4')
        expect(image.context_list.map(&:source_path)).to eq([File.join(src, 'app.rb')])
      end
    end
  end

  describe '.base' do
    it 'creates an image from a base image' do
      image = described_class.base('ruby:3.4')

      expect(image.dockerfile).to eq("FROM ruby:3.4\n")
    end
  end

  describe '.debian_slim' do
    it 'uses the latest supported Python series by default' do
      image = described_class.debian_slim

      expect(image.dockerfile).to include('FROM python:3.13.3-slim-bookworm')
      expect(image.dockerfile).to include('RUN pip install --upgrade pip')
    end

    it 'uses the latest micro version for a supported series' do
      image = described_class.debian_slim('3.12')

      expect(image.dockerfile).to include('FROM python:3.12.10-slim-bookworm')
    end

    it 'raises for unsupported Python versions' do
      expect { described_class.debian_slim('3.8') }
        .to raise_error(Daytona::Sdk::Error, /Unsupported Python version: 3.8/)
    end
  end
end
