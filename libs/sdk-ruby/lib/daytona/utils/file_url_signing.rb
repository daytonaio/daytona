# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

require 'base64'
require 'openssl'
require 'uri'

module Daytona
  module Utils
    module FileUrlSigning
      SIGNATURE_V1_PREFIX = 'v1_'
      private_constant :SIGNATURE_V1_PREFIX

      DEFAULT_TTL_SECONDS = 3600
      private_constant :DEFAULT_TTL_SECONDS

      module_function

      def compute_file_url_signature(signing_key, method, path, expires)
        canonical = "v1:files:#{method}:#{path}:#{expires}"
        digest = OpenSSL::HMAC.digest('sha256', signing_key, canonical)

        "#{SIGNATURE_V1_PREFIX}#{Base64.urlsafe_encode64(digest, padding: false)}"
      end

      def resolve_expires(ttl_seconds)
        return Time.now.to_i + DEFAULT_TTL_SECONDS if ttl_seconds.nil?

        return 0 if ttl_seconds <= 0

        Time.now.to_i + ttl_seconds
      end

      def build_signed_file_url(toolbox_proxy_url, sandbox_id, operation_path, method, file_path, signing_key,
                                ttl_seconds)
        if signing_key.nil? || signing_key.empty?
          raise Daytona::Sdk::Error,
                'Sandbox signing key is not available. Call refresh or fetch the sandbox by ID to load it.'
        end

        expires = resolve_expires(ttl_seconds)
        signature = compute_file_url_signature(signing_key, method, file_path, expires)
        query = URI.encode_www_form(path: file_path, expires: expires.to_s, signature:)

        "#{toolbox_proxy_url}/#{sandbox_id}#{operation_path}?#{query}"
      end
    end
  end
end
