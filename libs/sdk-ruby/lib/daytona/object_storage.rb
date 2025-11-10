# frozen_string_literal: true

require 'digest'
require 'fileutils'
require 'pathname'
require 'tempfile'
require 'zlib'
require 'aws-sdk-s3'

module Daytona
  class ObjectStorage
    # @return [String] The name of the S3 bucket used for object storage
    attr_reader :bucket_name

    # @return [Aws::S3::Client] The S3 client
    attr_reader :s3_client

    # Initialize ObjectStorage with S3-compatible credentials
    #
    # @param endpoint_url [String] The endpoint URL for the object storage service
    # @param aws_access_key_id [String] The access key ID for the object storage service
    # @param aws_secret_access_key [String] The secret access key for the object storage service
    # @param aws_session_token [String] The session token for the object storage service
    # @param bucket_name [String] The name of the bucket to use (defaults to "daytona-volume-builds")
    # @param region [String] AWS region (defaults to us-east-1)
    def initialize(endpoint_url:, aws_access_key_id:, aws_secret_access_key:, aws_session_token:, # rubocop:disable Metrics/ParameterLists
                   bucket_name: DEFAULT_BUCKET_NAME, region: DEFAULT_REGION)
      @bucket_name = bucket_name
      @s3_client = Aws::S3::Client.new(
        region:,
        endpoint: endpoint_url,
        access_key_id: aws_access_key_id,
        secret_access_key: aws_secret_access_key,
        session_token: aws_session_token
      )
    end

    # Uploads a file to the object storage service
    #
    # @param path [String] The path to the file to upload
    # @param organization_id [String] The organization ID to use
    # @param archive_base_path [String, nil] The base path to use for the archive
    # @return [String] The hash of the uploaded file
    # @raise [Errno::ENOENT] If the path does not exist
    def upload(path, organization_id, archive_base_path = nil)
      raise Errno::ENOENT, "Path does not exist: #{path}" unless File.exist?(path)

      path_hash = compute_hash_for_path_md5(path, archive_base_path)
      s3_key = "#{organization_id}/#{path_hash}/context.tar"

      return path_hash if file_exists_in_s3(s3_key)

      upload_as_tar(s3_key, path, archive_base_path)

      path_hash
    end

    # Compute the base path for an archive. Returns normalized path without the root
    # (drive letter or leading slash).
    #
    # @param path_str [String] The path to compute the base path for
    # @return [String] The base path for the given path
    def self.compute_archive_base_path(path_str)
      normalized_path = File.basename(path_str)

      # Remove drive letter for Windows paths (e.g., C:)
      path_without_drive = normalized_path.gsub(/^[A-Za-z]:/, '')

      # Remove leading separators (both / and \)
      path_without_drive.gsub(%r{^[/\\]+}, '')
    end

    private

    # Computes the MD5 hash for a given path
    #
    # @param path_str [String] The path to compute the hash for
    # @param archive_base_path [String, nil] The base path to use for the archive
    # @return [String] The MD5 hash for the given path
    def compute_hash_for_path_md5(path_str, archive_base_path = nil) # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/MethodLength, Metrics/PerceivedComplexity
      md5_hasher = Digest::MD5.new
      abs_path_str = File.expand_path(path_str)

      archive_base_path = self.class.compute_archive_base_path(path_str) if archive_base_path.nil?
      md5_hasher.update(archive_base_path)

      if File.file?(abs_path_str)
        File.open(abs_path_str, 'rb') do |f|
          while (chunk = f.read(8192))
            md5_hasher.update(chunk)
          end
        end
      else
        Dir.glob(File.join(abs_path_str, '**', '*')).each do |file_path|
          next unless File.file?(file_path)

          rel_path = Pathname.new(file_path).relative_path_from(Pathname.new(abs_path_str)).to_s

          md5_hasher.update(rel_path)

          File.open(file_path, 'rb') do |f|
            while (chunk = f.read(8192))
              md5_hasher.update(chunk)
            end
          end
        end

        # Handle empty directories
        Dir
          .glob(File.join(abs_path_str, '**', '*'))
          .select { |path| File.directory?(path) && Dir.empty?(path) }
          .each do |empty_dir|
          rel_dir = Pathname.new(empty_dir).relative_path_from(Pathname.new(abs_path_str)).to_s
          md5_hasher.update(rel_dir)
        end
      end

      md5_hasher.hexdigest
    end

    # Checks whether a specific object exists at the given path
    #
    # @param file_path [String] Full object path, e.g. "org/abcd123/context.tar"
    # @return [Boolean] True if the object exists, False otherwise
    def file_exists_in_s3(file_path)
      s3_client.head_object(bucket: bucket_name, key: file_path)
      true
    rescue Aws::S3::Errors::NotFound
      false
    rescue StandardError => e
      Sdk.logger.debug("Error checking file existence: #{e.message}")
      false
    end

    # Uploads a file to the object storage service as a tar
    #
    # @param s3_key [String] The key to upload the file to
    # @param source_path [String] The path to the file to upload
    # @param archive_base_path [String, nil] The base path to use for the archive
    def upload_as_tar(s3_key, source_path, archive_base_path = nil) # rubocop:disable Metrics/MethodLength
      source_path = File.expand_path(source_path)

      self.class.compute_archive_base_path(source_path) if archive_base_path.nil?

      temp_file = Tempfile.new(['context', '.tar'])

      begin
        system('tar', '-cf', temp_file.path, '-C', File.dirname(source_path), File.basename(source_path))

        File.open(temp_file.path, 'rb') do |file|
          s3_client.put_object(
            bucket: bucket_name,
            key: s3_key,
            body: file
          )
        end
      ensure
        temp_file.close
        temp_file.unlink
      end
    end

    DEFAULT_BUCKET_NAME = 'daytona-volume-builds'
    private_constant :DEFAULT_BUCKET_NAME

    DEFAULT_REGION = 'us-east-1'
    private_constant :DEFAULT_REGION
  end
end
