# frozen_string_literal: true

RSpec.describe Daytona::Config do
  around do |example|
    env_keys = %w[DAYTONA_API_KEY DAYTONA_JWT_TOKEN DAYTONA_API_URL DAYTONA_TARGET DAYTONA_ORGANIZATION_ID]
    saved = env_keys.to_h { |k| [k, ENV.delete(k)] }
    example.run
  ensure
    saved.each { |k, v| v ? ENV[k] = v : ENV.delete(k) }
  end

  describe '#initialize' do
    it 'accepts explicit api_key' do
      config = described_class.new(api_key: 'my-key')
      expect(config.api_key).to eq('my-key')
    end

    it 'accepts explicit jwt_token and organization_id' do
      config = described_class.new(jwt_token: 'jwt-tok', organization_id: 'org-42')
      expect(config.jwt_token).to eq('jwt-tok')
      expect(config.organization_id).to eq('org-42')
    end

    it 'defaults api_url to API_URL constant' do
      config = described_class.new(api_key: 'k')
      expect(config.api_url).to eq(described_class::API_URL)
    end

    it 'reads api_key from ENV' do
      ENV['DAYTONA_API_KEY'] = 'env-key'
      config = described_class.new
      expect(config.api_key).to eq('env-key')
    end

    it 'reads api_url from ENV' do
      ENV['DAYTONA_API_URL'] = 'https://custom.api'
      config = described_class.new(api_key: 'k')
      expect(config.api_url).to eq('https://custom.api')
    end

    it 'reads target from ENV' do
      ENV['DAYTONA_TARGET'] = 'eu'
      config = described_class.new(api_key: 'k')
      expect(config.target).to eq('eu')
    end

    it 'prefers explicit params over ENV' do
      ENV['DAYTONA_API_KEY'] = 'env-key'
      config = described_class.new(api_key: 'explicit-key')
      expect(config.api_key).to eq('explicit-key')
    end

    it 'stores experimental config' do
      config = described_class.new(api_key: 'k', _experimental: { 'otel_enabled' => true })
      expect(config._experimental).to eq({ 'otel_enabled' => true })
    end
  end

  describe '#read_env' do
    it 'returns value for DAYTONA_-prefixed variable from ENV' do
      ENV['DAYTONA_CUSTOM_VAR'] = 'hello'
      config = described_class.new(api_key: 'k')
      expect(config.read_env('DAYTONA_CUSTOM_VAR')).to eq('hello')
    ensure
      ENV.delete('DAYTONA_CUSTOM_VAR')
    end

    it 'raises ArgumentError for non-DAYTONA_ variable names' do
      config = described_class.new(api_key: 'k')
      expect { config.read_env('OTHER_VAR') }.to raise_error(ArgumentError, /must start with 'DAYTONA_'/)
    end

    it 'returns nil for unset DAYTONA_ variable' do
      config = described_class.new(api_key: 'k')
      expect(config.read_env('DAYTONA_NONEXISTENT')).to be_nil
    end
  end
end
