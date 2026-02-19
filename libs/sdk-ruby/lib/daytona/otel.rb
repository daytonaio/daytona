# frozen_string_literal: true

module Daytona
  # Holds OTel provider state for the SDK.
  class OtelState
    attr_reader :tracer_provider
    attr_reader :meter_provider
    attr_reader :tracer
    attr_reader :meter

    def initialize(tracer_provider:, meter_provider:, tracer:, meter:)
      @tracer_provider = tracer_provider
      @meter_provider = meter_provider
      @tracer = tracer
      @meter = meter
      @histograms = {}
      @histograms_mutex = Mutex.new
    end

    # Returns a cached histogram for the given metric name.
    def histogram(name)
      @histograms_mutex.synchronize do
        @histograms[name] ||= meter.create_histogram(name, unit: 'ms')
      end
    end

    def shutdown
      tracer_provider.shutdown
      meter_provider.shutdown
    end
  end

  # Initializes OTel providers, sets globals, installs Typhoeus propagation.
  # OTel gems are required lazily so they are never loaded when disabled.
  #
  # @param sdk_version [String]
  # @return [OtelState]
  def self.init_otel(sdk_version) # rubocop:disable Metrics/MethodLength
    require 'opentelemetry-sdk'
    require 'opentelemetry-metrics-sdk'
    require 'opentelemetry-exporter-otlp'
    require 'opentelemetry-exporter-otlp-metrics'

    resource = OpenTelemetry::SDK::Resources::Resource.create(
      'service.name' => 'daytona-ruby-sdk',
      'service.version' => sdk_version
    )

    tracer_provider = OpenTelemetry::SDK::Trace::TracerProvider.new(resource:)
    tracer_provider.add_span_processor(
      OpenTelemetry::SDK::Trace::Export::BatchSpanProcessor.new(
        OpenTelemetry::Exporter::OTLP::Exporter.new
      )
    )
    OpenTelemetry.tracer_provider = tracer_provider

    meter_provider = OpenTelemetry::SDK::Metrics::MeterProvider.new(resource:)
    meter_provider.add_metric_reader(
      OpenTelemetry::SDK::Metrics::Export::PeriodicMetricReader.new(
        exporter: OpenTelemetry::Exporter::OTLP::Metrics::MetricsExporter.new
      )
    )
    OpenTelemetry.meter_provider = meter_provider

    tracer = tracer_provider.tracer('daytona-ruby-sdk', sdk_version)
    meter = meter_provider.meter('daytona-ruby-sdk')

    # Install Typhoeus trace-context propagation
    install_typhoeus_propagation

    OtelState.new(tracer_provider:, meter_provider:, tracer:, meter:)
  end

  # Flushes and shuts down OTel providers.
  def self.shutdown_otel(state)
    state&.shutdown
  end

  # Wraps a block with OTel span creation and duration histogram recording.
  # When otel_state is nil (OTel disabled), calls the block directly.
  #
  # @param otel_state [OtelState, nil]
  # @param component [String]
  # @param method_name [String]
  # @return [Object] The block's return value
  def self.with_instrumentation(otel_state, component, method_name, &block) # rubocop:disable Metrics/MethodLength
    return block.call unless otel_state

    span_name = "#{component}.#{method_name}"
    metric_name = "#{to_snake_case(span_name)}_duration"
    status = 'success'

    otel_state.tracer.in_span(
      span_name,
      attributes: { 'component' => component, 'method' => method_name }
    ) do |_span|
      start_time = ::Process.clock_gettime(::Process::CLOCK_MONOTONIC)
      begin
        block.call
      rescue StandardError => e
        status = 'error'
        raise e
      ensure
        duration_ms = (::Process.clock_gettime(::Process::CLOCK_MONOTONIC) - start_time) * 1000.0
        otel_state.histogram(metric_name).record(
          duration_ms,
          attributes: { 'component' => component, 'method' => method_name, 'status' => status }
        )
      end
    end
  end

  # Converts "ClassName.method_name" to "class_name_method_name".
  def self.to_snake_case(str)
    result = +''
    str.each_char.with_index do |char, i|
      if char == '.'
        result << '_'
      elsif char =~ /[A-Z]/ && i > 0 && str[i - 1] != '.'
        result << '_' << char.downcase
      else
        result << char.downcase
      end
    end
    result
  end

  # Installs Typhoeus.before callback for W3C trace-context propagation.
  def self.install_typhoeus_propagation
    return unless defined?(Typhoeus)

    Typhoeus.before do |request|
      headers = request.options[:headers] ||= {}
      OpenTelemetry.propagation.inject(headers)
      true
    end
  end

  # Mixin that provides the `instrument` class macro for wrapping methods
  # with OTel spans and metrics.
  module Instrumentation
    def self.included(base)
      base.extend(ClassMethods)
    end

    module ClassMethods
      # Instruments the listed methods with OTel tracing/metrics.
      # Must be called after all target methods are defined.
      #
      # @param method_names [Array<Symbol>] methods to instrument
      # @param component [String] component name for span/metric attributes
      def instrument(*method_names, component:) # rubocop:disable Metrics/MethodLength
        method_names.each do |method_name|
          original = instance_method(method_name)

          # Detect original visibility
          visibility = if private_method_defined?(method_name, false)
                         :private
                       elsif protected_method_defined?(method_name, false)
                         :protected
                       else
                         :public
                       end

          define_method(method_name) do |*args, **kwargs, &blk|
            ::Daytona.with_instrumentation(otel_state, component, method_name.to_s) do
              original.bind_call(self, *args, **kwargs, &blk)
            end
          end

          # Restore visibility
          case visibility
          when :private   then private method_name
          when :protected then protected method_name
          end
        end
      end
    end
  end
end
