# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::FileSystem do
  # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
  def multipart_response(parts, boundary: 'DAYTONA-FILE-BOUNDARY')
    body = String.new.b

    parts.each do |part|
      body << "--#{boundary}\r\n"
      body << %(Content-Disposition: form-data; name="#{part.fetch(:name)}")
      body << %(; filename="#{part[:filename]}") if part[:filename]
      body << "\r\n"
      part_body = part.fetch(:body)
      body << "Content-Type: #{part.fetch(:content_type, 'application/octet-stream')}\r\n"
      body << "Content-Length: #{part_body.bytesize}\r\n\r\n"
      body << part_body
      body << "\r\n"
    end

    body << "--#{boundary}--\r\n"
  end

  def stub_streaming_request(chunks:, headers: nil, success: true, code: 200, body: nil)
    headers ||= { 'Content-Type' => 'multipart/form-data; boundary=DAYTONA-FILE-BOUNDARY' }
    request = instance_double(Typhoeus::Request)
    callbacks = {}

    allow(Typhoeus::Request).to receive(:new).and_return(request)
    allow(request).to receive(:on_headers) { |&block| callbacks[:headers] = block }
    allow(request).to receive(:on_body) { |&block| callbacks[:body] = block }
    allow(request).to receive(:on_complete) { |&block| callbacks[:complete] = block }
    allow(request).to receive(:run) do
      callbacks[:headers]&.call(double('headers_response', headers: headers))
      chunks.each { |chunk| callbacks[:body]&.call(chunk) }
      callbacks[:complete]&.call(double('complete_response', success?: success, code: code, body: body))
    end

    request
  end

  def build_cancel_event
    Class.new do
      def initialize
        @set = false
      end

      def set!
        @set = true
      end

      def set?
        @set
      end
    end.new
  end

  def stub_upload_request(success: true, code: 200, body: '', return_code: :ok, timed_out: false, &block)
    request = instance_double(Typhoeus::Request)
    request_options = nil
    response = instance_double(
      Typhoeus::Response,
      success?: success,
      code: code,
      body: body,
      return_code: return_code,
      timed_out?: timed_out
    )

    allow(Typhoeus::Request).to receive(:new) do |_url, options|
      request_options = options
      request
    end
    allow(request).to receive(:run) do
      block&.call(request_options)
      response
    end

    [request, response, -> { request_options }]
  end
  # rubocop:enable Metrics/AbcSize, Metrics/MethodLength

  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::FileSystemApi) }
  let(:toolbox_api_config) { double('ToolboxConfig', base_url: 'https://toolbox.example.com', verify_ssl: true, verify_ssl_host: true) }
  let(:toolbox_api_client) do
    double('ToolboxApiClient', config: toolbox_api_config, default_headers: { 'Authorization' => 'Bearer token' })
  end
  let(:fs) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

  before do
    allow(toolbox_api).to receive(:api_client).and_return(toolbox_api_client)
  end

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

  describe '#download_file_stream' do
    it 'yields file content chunks when block given' do
      body = multipart_response([
                                  { name: 'file', filename: 'remote.txt', body: 'stream test content' }
                                ])
      stub_streaming_request(chunks: [body.byteslice(0, 96), body.byteslice(96, 8),
                                      body.byteslice(104, body.bytesize - 104)])

      chunks = []
      fs.download_file_stream('/remote.txt', timeout: 45) { |chunk| chunks << chunk }

      expect(chunks.join).to eq('stream test content')
      expect(Typhoeus::Request).to have_received(:new).with(
        'https://toolbox.example.com/files/bulk-download',
        hash_including(
          method: :post,
          timeout: 45,
          body: '{"paths":["/remote.txt"]}',
          headers: hash_including(
            'Authorization' => 'Bearer token',
            'Accept' => 'multipart/form-data',
            'Content-Type' => 'application/json'
          )
        )
      )
    end

    it 'returns enumerator when no block given' do
      body = multipart_response([
                                  { name: 'file', filename: 'remote.txt', body: 'enumerated content' }
                                ])
      stub_streaming_request(chunks: [body.byteslice(0, 70), body.byteslice(70, body.bytesize - 70)])

      enumerator = fs.download_file_stream('/remote.txt')

      expect(enumerator).to be_a(Enumerator)
      expect(enumerator.to_a.join).to eq('enumerated content')
    end

    it 'calls on_progress with bytes_received and total_bytes' do
      body = multipart_response([{ name: 'file', filename: 'remote.txt', body: 'hello world' }])
      stub_streaming_request(chunks: [body])

      progress_calls = []
      fs.download_file_stream('/remote.txt', on_progress: ->(progress) { progress_calls << progress }) { |_chunk| nil }

      expect(progress_calls).not_to be_empty
      expect(progress_calls.last.bytes_received).to eq('hello world'.bytesize)
      expect(progress_calls.last.total_bytes).to eq('hello world'.bytesize)
    end

    it 'raises error when file not found' do
      body = multipart_response([
                                  { name: 'error', content_type: 'application/json',
                                    body: '{"message":"file not found"}' }
                                ])
      stub_streaming_request(chunks: [body.byteslice(0, 82), body.byteslice(82, body.bytesize - 82)])

      expect { fs.download_file_stream('/missing.txt') { |_chunk| nil } }
        .to raise_error(Daytona::Sdk::Error, /Failed to download file: file not found/)
    end

    it 'aborts when cancel_event is set before the first chunk' do
      body = multipart_response([{ name: 'file', filename: 'remote.txt', body: 'hello world' }])
      stub_streaming_request(chunks: [body])
      cancel = double('CancelEvent', set?: true)

      expect { fs.download_file_stream('/remote.txt', cancel_event: cancel) { |_chunk| nil } }
        .to raise_error(Daytona::Sdk::Error, /Failed to download file: Download cancelled/)
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

  describe '#upload_file_stream' do
    it 'aborts when cancel_event is set before the upload starts' do
      cancel = double('CancelEvent', set?: true)
      io = StringIO.new('streamed-bytes')

      expect { fs.upload_file_stream(io, '/remote.bin', cancel_event: cancel) }
        .to raise_error(Daytona::Sdk::Error, /Failed to upload file: Upload cancelled/)
    end

    it 'cancels mid-upload from the libcurl progress callback' do
      cancel = build_cancel_event
      progress_calls = []
      stub_upload_request(success: false, code: 0, return_code: :aborted_by_callback) do |options|
        progress = options.fetch(:xferinfofunction)

        expect(progress.call(nil, 0, 0, 128, 32)).to eq(0)
        cancel.set!
        expect(progress.call(nil, 0, 0, 128, 64)).to eq(1)
      end

      expect do
        fs.upload_file_stream(
          StringIO.new('streamed-bytes'),
          '/remote.bin',
          cancel_event: cancel,
          on_progress: ->(p) { progress_calls << p }
        )
      end.to raise_error(Daytona::Sdk::Error, /Failed to upload file: Upload cancelled/)

      expect(progress_calls.map(&:bytes_sent)).to eq([32])
    end

    it 'closes the upload file handle after the request completes' do
      uploaded_file = nil
      stub_upload_request do |options|
        uploaded_file = options.dig(:body, 'files[0].file')
        expect(uploaded_file).to be_a(File)
        expect(uploaded_file.closed?).to be(false)
      end

      Dir.mktmpdir do |dir|
        source_path = File.join(dir, 'upload.bin')
        File.binwrite(source_path, 'payload')

        fs.upload_file_stream(source_path, '/remote.bin')
      end

      expect(uploaded_file.closed?).to be(true)
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
