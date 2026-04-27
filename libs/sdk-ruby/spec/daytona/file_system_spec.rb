# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::FileSystem do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::FileSystemApi) }
  let(:fs) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

  describe '#create_folder' do
    it 'delegates to toolbox_api' do
      allow(toolbox_api).to receive(:create_folder).with('/workspace/data', '755')

      fs.create_folder('/workspace/data', '755')

      expect(toolbox_api).to have_received(:create_folder).with('/workspace/data', '755')
    end

    it 'wraps errors in Sdk::Error' do
      allow(toolbox_api).to receive(:create_folder).and_raise(StandardError, 'fail')

      expect { fs.create_folder('/x', '755') }.to raise_error(Daytona::Sdk::Error, /Failed to create folder: fail/)
    end
  end

  describe '#delete_file' do
    it 'deletes a file' do
      allow(toolbox_api).to receive(:delete_file).with('/test.txt', { recursive: false })

      fs.delete_file('/test.txt')

      expect(toolbox_api).to have_received(:delete_file).with('/test.txt', { recursive: false })
    end

    it 'deletes a directory recursively' do
      allow(toolbox_api).to receive(:delete_file).with('/dir', { recursive: true })

      fs.delete_file('/dir', recursive: true)

      expect(toolbox_api).to have_received(:delete_file).with('/dir', { recursive: true })
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:delete_file).and_raise(StandardError, 'nope')

      expect { fs.delete_file('/x') }.to raise_error(Daytona::Sdk::Error, /Failed to delete file: nope/)
    end
  end

  describe '#get_file_info' do
    it 'returns file info' do
      info = double('FileInfo', size: 1024, is_dir: false)
      allow(toolbox_api).to receive(:get_file_info).with('/test.txt').and_return(info)

      expect(fs.get_file_info('/test.txt')).to eq(info)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_file_info).and_raise(StandardError, 'err')

      expect { fs.get_file_info('/x') }.to raise_error(Daytona::Sdk::Error, /Failed to get file info: err/)
    end
  end

  describe '#list_files' do
    it 'returns file list' do
      files = [double('FileInfo', name: 'a.txt'), double('FileInfo', name: 'b.rb')]
      allow(toolbox_api).to receive(:list_files).with({ path: '/workspace' }).and_return(files)

      expect(fs.list_files('/workspace')).to eq(files)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:list_files).and_raise(StandardError, 'err')

      expect { fs.list_files('/x') }.to raise_error(Daytona::Sdk::Error, /Failed to list files: err/)
    end
  end

  describe '#download_file' do
    it 'returns the file object when no local_path is given' do
      file_obj = double('File')
      allow(toolbox_api).to receive(:download_file).with('/remote.txt').and_return(file_obj)

      expect(fs.download_file('/remote.txt')).to eq(file_obj)
    end

    it 'saves the file to local_path and returns nil' do
      io = StringIO.new('content')
      file_obj = double('TempFile', open: io)
      allow(toolbox_api).to receive(:download_file).with('/remote.txt').and_return(file_obj)

      Dir.mktmpdir do |dir|
        local_path = File.join(dir, 'nested', 'local.txt')
        result = fs.download_file('/remote.txt', local_path)

        expect(result).to be_nil
        expect(File.read(local_path)).to eq('content')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:download_file).and_raise(StandardError, 'err')

      expect { fs.download_file('/x') }.to raise_error(Daytona::Sdk::Error, /Failed to download file: err/)
    end
  end

  describe '#upload_file' do
    it 'uploads string content via temp file' do
      allow(toolbox_api).to receive(:upload_file)

      fs.upload_file('hello world', '/remote/file.txt')

      expect(toolbox_api).to have_received(:upload_file).with('/remote/file.txt', anything)
    end

    it 'uploads a local file path by opening it in binary mode' do
      allow(toolbox_api).to receive(:upload_file)

      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'local.txt')
        File.binwrite(file_path, 'abc')

        fs.upload_file(file_path, '/remote/local.txt')

        expect(toolbox_api).to have_received(:upload_file).with('/remote/local.txt', kind_of(File))
      end
    end

    it 'uploads IO objects directly' do
      io = StringIO.new('data')
      allow(toolbox_api).to receive(:upload_file)

      fs.upload_file(io, '/remote/io.txt')

      expect(toolbox_api).to have_received(:upload_file).with('/remote/io.txt', io)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:upload_file).and_raise(StandardError, 'err')

      expect { fs.upload_file('data', '/x') }.to raise_error(Daytona::Sdk::Error, /Failed to upload file: err/)
    end
  end

  describe '#upload_files' do
    it 'uploads multiple files' do
      allow(toolbox_api).to receive(:upload_file)
      files = [
        Daytona::FileUpload.new('content1', '/dest1'),
        Daytona::FileUpload.new('content2', '/dest2')
      ]

      fs.upload_files(files)

      expect(toolbox_api).to have_received(:upload_file).twice
    end

    it 'wraps errors from individual uploads' do
      allow(fs).to receive(:upload_file).and_raise(StandardError, 'boom')

      expect { fs.upload_files([Daytona::FileUpload.new('content', '/dest')]) }
        .to raise_error(Daytona::Sdk::Error, /Failed to upload files: boom/)
    end
  end

  describe '#find_files' do
    it 'delegates to toolbox_api.find_in_files' do
      matches = [double('Match')]
      allow(toolbox_api).to receive(:find_in_files).with('/workspace', 'TODO:').and_return(matches)

      expect(fs.find_files('/workspace', 'TODO:')).to eq(matches)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:find_in_files).and_raise(StandardError, 'err')

      expect { fs.find_files('/x', 'pat') }.to raise_error(Daytona::Sdk::Error, /Failed to find files: err/)
    end
  end

  describe '#search_files' do
    it 'delegates to toolbox_api.search_files' do
      result = double('SearchResult', files: ['a.rb'])
      allow(toolbox_api).to receive(:search_files).with('/workspace', '*.rb').and_return(result)

      expect(fs.search_files('/workspace', '*.rb')).to eq(result)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:search_files).and_raise(StandardError, 'err')

      expect do
        fs.search_files('/workspace', '*.rb')
      end.to raise_error(Daytona::Sdk::Error, /Failed to search files: err/)
    end
  end

  describe '#move_files' do
    it 'delegates to toolbox_api.move_file' do
      allow(toolbox_api).to receive(:move_file).with('/old', '/new')

      fs.move_files('/old', '/new')

      expect(toolbox_api).to have_received(:move_file).with('/old', '/new')
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:move_file).and_raise(StandardError, 'err')

      expect { fs.move_files('/a', '/b') }.to raise_error(Daytona::Sdk::Error, /Failed to move files: err/)
    end
  end

  describe '#replace_in_files' do
    it 'sends ReplaceRequest to toolbox_api' do
      results = [double('ReplaceResult')]
      allow(toolbox_api).to receive(:replace_in_files).and_return(results)

      result = fs.replace_in_files(files: ['/f.rb'], pattern: 'old', new_value: 'new')

      expect(result).to eq(results)
      expect(toolbox_api).to have_received(:replace_in_files) do |request|
        expect(request.files).to eq(['/f.rb'])
        expect(request.pattern).to eq('old')
        expect(request.new_value).to eq('new')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:replace_in_files).and_raise(StandardError, 'err')

      expect { fs.replace_in_files(files: [], pattern: 'a', new_value: 'b') }
        .to raise_error(Daytona::Sdk::Error, /Failed to replace in files: err/)
    end
  end

  describe '#set_file_permissions' do
    it 'sets permissions with mode, owner, and group' do
      allow(toolbox_api).to receive(:set_file_permissions)
        .with('/script.sh', { mode: '755', owner: 'root', group: 'root' })

      fs.set_file_permissions(path: '/script.sh', mode: '755', owner: 'root', group: 'root')

      expect(toolbox_api).to have_received(:set_file_permissions).with('/script.sh',
                                                                       { mode: '755', owner: 'root', group: 'root' })
    end

    it 'omits nil options' do
      allow(toolbox_api).to receive(:set_file_permissions).with('/f.txt', { mode: '644' })

      fs.set_file_permissions(path: '/f.txt', mode: '644')

      expect(toolbox_api).to have_received(:set_file_permissions).with('/f.txt', { mode: '644' })
    end

    it 'sends an empty options hash when only the path is provided' do
      allow(toolbox_api).to receive(:set_file_permissions).with('/f.txt', {})

      fs.set_file_permissions(path: '/f.txt')

      expect(toolbox_api).to have_received(:set_file_permissions).with('/f.txt', {})
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:set_file_permissions).and_raise(StandardError, 'err')

      expect do
        fs.set_file_permissions(path: '/x')
      end.to raise_error(Daytona::Sdk::Error, /Failed to set file permissions: err/)
    end
  end
end
