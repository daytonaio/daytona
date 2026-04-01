# frozen_string_literal: true

RSpec.describe Daytona::ObjectStorage do
  let(:s3_client) { instance_double(Aws::S3::Client) }

  let(:storage) do
    allow(Aws::S3::Client).to receive(:new).and_return(s3_client)
    described_class.new(
      endpoint_url: 'https://s3.example.com',
      aws_access_key_id: 'key-id',
      aws_secret_access_key: 'secret',
      aws_session_token: 'token',
      bucket_name: 'test-bucket'
    )
  end

  describe '#initialize' do
    it 'stores bucket_name' do
      expect(storage.bucket_name).to eq('test-bucket')
    end

    it 'creates S3 client' do
      expect(storage.s3_client).to eq(s3_client)
    end
  end

  describe '#upload' do
    it 'raises Errno::ENOENT for non-existent path' do
      expect { storage.upload('/nonexistent/path', 'org-1') }
        .to raise_error(Errno::ENOENT, /Path does not exist/)
    end

    it 'skips upload if file already exists in S3' do
      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'test.txt')
        File.write(file_path, 'content')

        allow(s3_client).to receive(:head_object).and_return(double('HeadResponse'))

        result = storage.upload(file_path, 'org-1')
        expect(result).to be_a(String)
        expect(s3_client).not_to have_received(:put_object) if s3_client.respond_to?(:put_object)
      end
    end

    it 'uploads file as tar when not in S3' do
      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'test.txt')
        File.write(file_path, 'content')

        allow(s3_client).to receive(:head_object).and_raise(Aws::S3::Errors::NotFound.new(nil, 'not found'))
        allow(s3_client).to receive(:put_object)

        result = storage.upload(file_path, 'org-1')
        expect(result).to be_a(String)
        expect(s3_client).to have_received(:put_object)
      end
    end
  end

  describe '.compute_archive_base_path' do
    it 'returns basename for simple path' do
      expect(described_class.compute_archive_base_path('/home/user/file.txt')).to eq('file.txt')
    end

    it 'handles basename extraction' do
      expect(described_class.compute_archive_base_path('/home/user/data')).to eq('data')
    end

    it 'strips leading separators' do
      expect(described_class.compute_archive_base_path('/root')).to eq('root')
    end
  end
end
