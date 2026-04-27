# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona do
  describe Daytona::OtelState do
    let(:histogram) { instance_double('Histogram') }
    let(:meter) { instance_double('Meter') }
    let(:tracer_provider) { instance_double('TracerProvider') }
    let(:meter_provider) { instance_double('MeterProvider') }
    let(:tracer) { instance_double('Tracer') }
    let(:state) do
      described_class.new(
        tracer_provider: tracer_provider,
        meter_provider: meter_provider,
        tracer: tracer,
        meter: meter
      )
    end

    it 'caches histograms by metric name' do
      allow(meter).to receive(:create_histogram).with('duration_ms', unit: 'ms').and_return(histogram)

      first = state.histogram('duration_ms')
      second = state.histogram('duration_ms')

      expect(first).to eq(histogram)
      expect(second).to eq(histogram)
      expect(meter).to have_received(:create_histogram).once
    end

    it 'shuts down both providers' do
      allow(tracer_provider).to receive(:shutdown)
      allow(meter_provider).to receive(:shutdown)

      state.shutdown

      expect(tracer_provider).to have_received(:shutdown)
      expect(meter_provider).to have_received(:shutdown)
    end
  end

  describe '.shutdown_otel' do
    it 'is nil-safe' do
      expect { described_class.shutdown_otel(nil) }.not_to raise_error
    end

    it 'delegates to the state shutdown method' do
      state = instance_double(Daytona::OtelState)
      allow(state).to receive(:shutdown)

      described_class.shutdown_otel(state)

      expect(state).to have_received(:shutdown)
    end
  end

  describe '.with_instrumentation' do
    it 'executes the block directly when otel is disabled' do
      expect(described_class.with_instrumentation(nil, 'Sandbox', 'start') { :ok }).to eq(:ok)
    end

    it 'records a successful duration metric' do
      histogram = instance_double('Histogram')
      tracer = instance_double('Tracer')
      state = instance_double(Daytona::OtelState, tracer: tracer)
      allow(state).to receive(:histogram).with('sandbox_start_duration').and_return(histogram)
      allow(tracer).to receive(:in_span).and_yield(nil)
      allow(histogram).to receive(:record)

      result = described_class.with_instrumentation(state, 'Sandbox', 'start') { :done }

      expect(result).to eq(:done)
      expect(tracer).to have_received(:in_span).with(
        'Sandbox.start',
        attributes: { 'component' => 'Sandbox', 'method' => 'start' }
      )
      expect(histogram).to have_received(:record) do |_duration, attributes:|
        expect(attributes['status']).to eq('success')
      end
    end

    it 'records an error status and re-raises failures' do
      histogram = instance_double('Histogram')
      tracer = instance_double('Tracer')
      state = instance_double(Daytona::OtelState, tracer: tracer)
      allow(state).to receive(:histogram).with('sandbox_start_duration').and_return(histogram)
      allow(tracer).to receive(:in_span).and_yield(nil)
      allow(histogram).to receive(:record)

      expect do
        described_class.with_instrumentation(state, 'Sandbox', 'start') { raise StandardError, 'boom' }
      end.to raise_error(StandardError, 'boom')

      expect(histogram).to have_received(:record) do |_duration, attributes:|
        expect(attributes['status']).to eq('error')
      end
    end
  end

  describe '.to_snake_case' do
    it 'converts dotted camel case names to snake case' do
      expect(described_class.to_snake_case('Sandbox.start')).to eq('sandbox_start')
      expect(described_class.to_snake_case('CodeInterpreter.runCode')).to eq('code_interpreter_run_code')
    end
  end

  describe '.install_typhoeus_propagation' do
    it 'does nothing when Typhoeus is unavailable' do
      hide_const('Typhoeus') if defined?(Typhoeus)

      expect { described_class.install_typhoeus_propagation }.not_to raise_error
    end

    it 'registers a before hook and injects headers' do
      typhoeus = Module.new do
        class << self
          attr_accessor :before_block
        end

        def self.before(&block)
          self.before_block = block
        end
      end
      propagation = double('Propagation')
      request = double('Request', options: {})

      stub_const('Typhoeus', typhoeus)
      open_telemetry = Module.new do
        class << self
          attr_accessor :propagation
        end
      end
      stub_const('OpenTelemetry', open_telemetry)
      OpenTelemetry.propagation = propagation
      allow(propagation).to receive(:inject)

      described_class.install_typhoeus_propagation
      Typhoeus.before_block.call(request)

      expect(propagation).to have_received(:inject).with(request.options[:headers])
    end
  end

  describe Daytona::Instrumentation do
    let(:test_class) do
      Class.new do
        include Daytona::Instrumentation

        def initialize(otel_state)
          @otel_state = otel_state
        end

        attr_reader :otel_state

        def ping(value)
          "pong #{value}"
        end

        def secret
          'secret'
        end
        private :secret

        instrument :ping, :secret, component: 'Dummy'
      end
    end

    it 'wraps public methods with Daytona.with_instrumentation' do
      otel_state = double('OtelState')
      instance = test_class.new(otel_state)
      allow(Daytona).to receive(:with_instrumentation).and_yield

      expect(instance.ping('value')).to eq('pong value')
      expect(Daytona).to have_received(:with_instrumentation).with(otel_state, 'Dummy', 'ping')
    end

    it 'preserves private method visibility' do
      instance = test_class.new(nil)

      expect(instance.private_methods).to include(:secret)
    end
  end
end
