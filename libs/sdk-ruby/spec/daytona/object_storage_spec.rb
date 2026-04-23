# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

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
    it 'stores bucket_name and creates the S3 client' do
      expect(storage.bucket_name).to eq('test-bucket')
      expect(storage.s3_client).to eq(s3_client)
    end
  end

  describe '#upload' do
    it 'raises Errno::ENOENT for non-existent paths' do
      expect { storage.upload('/nonexistent/path', 'org-1') }
        .to raise_error(Errno::ENOENT, /Path does not exist/)
    end

    it 'skips upload if the file already exists in S3' do
      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'test.txt')
        File.write(file_path, 'content')
        allow(s3_client).to receive(:head_object).and_return(double('HeadResponse'))
        allow(s3_client).to receive(:put_object)

        result = storage.upload(file_path, 'org-1')

        expect(result).to be_a(String)
        expect(s3_client).not_to have_received(:put_object)
      end
    end

    it 'uploads a file as a tar when missing from S3' do
      Dir.mktmpdir do |dir|
        file_path = File.join(dir, 'test.txt')
        File.write(file_path, 'content')
        allow(s3_client).to receive(:head_object).and_raise(Aws::S3::Errors::NotFound.new(nil, 'not found'))
        allow(s3_client).to receive(:put_object)
        allow(storage).to receive(:system).with('tar', '-cf', anything, '-C', dir, 'test.txt').and_return(true)

        result = storage.upload(file_path, 'org-1')

        expect(result).to be_a(String)
        expect(s3_client).to have_received(:put_object)
      end
    end
  end

  describe '.compute_archive_base_path' do
    it 'returns basename for simple paths' do
      expect(described_class.compute_archive_base_path('/home/user/file.txt')).to eq('file.txt')
    end

    it 'removes Windows drive prefixes and separators from basenames' do
      expect(described_class.compute_archive_base_path('C:/Users/me/project')).to eq('project')
    end
  end

  describe 'private helpers' do
    it 'returns false for generic S3 head_object errors' do
      allow(s3_client).to receive(:head_object).and_raise(StandardError, 'broken')

      expect(storage.send(:file_exists_in_s3, 'org/hash/context.tar')).to be(false)
    end

    it 'computes different hashes for different file contents' do
      Dir.mktmpdir do |dir|
        path_a = File.join(dir, 'a.txt')
        path_b = File.join(dir, 'b.txt')
        File.write(path_a, 'alpha')
        File.write(path_b, 'beta')

        hash_a = storage.send(:compute_hash_for_path_md5, path_a)
        hash_b = storage.send(:compute_hash_for_path_md5, path_b)

        expect(hash_a).not_to eq(hash_b)
      end
    end

    it 'includes empty directories in directory hashes' do
      Dir.mktmpdir do |dir|
        Dir.mkdir(File.join(dir, 'empty'))

        hash = storage.send(:compute_hash_for_path_md5, dir)

        expect(hash).to match(/\A[0-9a-f]{32}\z/)
      end
    end
  end
end
