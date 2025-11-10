# frozen_string_literal: true

require 'json'
require 'observer'

module Daytona
  class PtySize
    # @return [Integer] Number of terminal rows (height)
    attr_reader :rows

    # @return [Integer] Number of terminal columns (width)
    attr_reader :cols

    # Initialize a new PtySize
    #
    # @param rows [Integer] Number of terminal rows (height)
    # @param cols [Integer] Number of terminal columns (width)
    def initialize(rows:, cols:)
      @rows = rows
      @cols = cols
    end
  end

  class PtyResult
    # @return [Integer, nil] Exit code of the PTY process (0 for success, non-zero for errors).
    #                        nil if the process hasn't exited yet or exit code couldn't be determined.
    attr_reader :exit_code

    # @return [String, nil] Error message if the PTY failed or was terminated abnormally.
    #                       nil if no error occurred.
    attr_reader :error

    # Initialize a new PtyResult
    #
    # @param exit_code [Integer, nil] Exit code of the PTY process
    # @param error [String, nil] Error message if the PTY failed
    def initialize(exit_code: nil, error: nil)
      @exit_code = exit_code
      @error = error
    end
  end

  class PtyHandle # rubocop:disable Metrics/ClassLength
    include Observable

    # @return [String] Session ID of the PTY session
    attr_reader :session_id

    # @return [Integer, nil] Exit code of the PTY process (if terminated)
    attr_reader :exit_code

    # @return [String, nil] Error message if the PTY failed
    attr_reader :error

    # Initialize the PTY handle.
    #
    # @param websocket [WebSocket::Client::Simple::Client] Connected WebSocket client connection
    # @param session_id [String] Session ID of the PTY session
    # @param handle_resize [Proc, nil] Optional callback for resizing the PTY
    # @param handle_kill [Proc, nil] Optional callback for killing the PTY
    def initialize(websocket, session_id:, handle_resize: nil, handle_kill: nil)
      @websocket = websocket
      @session_id = session_id
      @handle_resize = handle_resize
      @handle_kill = handle_kill
      @exit_code = nil
      @error = nil
      @logger = Sdk.logger

      @status = Status::INIT
      subscribe
    end

    # Check if connected to the PTY session
    #
    # @return [Boolean] true if connected, false otherwise
    def connected? = websocket.open?

    # Wait for the PTY connection to be established
    #
    # @param timeout [Float] Maximum time in seconds to wait for connection. Defaults to 10.0
    # @return [void]
    # @raise [Daytona::Sdk::Error] If connection timeout is exceeded
    def wait_for_connection(timeout: DEFAULT_TIMEOUT)
      return if status == Status::CONNECTED

      start_time = Time.now

      sleep(SLEEP_INTERVAL) until status == Status::CONNECTED || (Time.now - start_time) > timeout

      raise Sdk::Error, 'PTY connection timeout' unless status == Status::CONNECTED
    end

    # Send input to the PTY session
    #
    # @param input [String] Input to send to the PTY
    # @return [void]
    def send_input(input)
      raise Sdk::Error, 'PTY session not connected' unless websocket.open?

      websocket.send(input)
    end

    # Resize the PTY terminal
    #
    # @param pty_size [PtySize] New terminal size
    # @return [DaytonaApiClient::PtySessionInfo] Updated PTY session information
    def resize(pty_size)
      raise Sdk::Error, 'No resize handler available' unless handle_resize

      handle_resize.call(pty_size)
    end

    # Delete the PTY session
    #
    # @return [void]
    def kill
      raise Sdk::Error, 'No kill handler available' unless handle_kill

      handle_kill.call
    end

    # Wait for the PTY session to complete
    #
    # @param on_data [Proc, nil] Optional callback to handle output data
    # @return [Daytona::PtyResult] Result containing exit code and error information
    def wait(timeout: nil, &on_data)
      timeout ||= Float::INFINITY
      return unless status == Status::CONNECTED

      start_time = Time.now
      add_observer(on_data, :call) if on_data

      sleep(SLEEP_INTERVAL) while status == Status::CONNECTED && (Time.now - start_time) <= timeout

      PtyResult.new(exit_code:, error:)
    ensure
      delete_observer(on_data) if on_data
    end

    # @yieldparam [WebSocket::Frame::Data]
    # @return [void]
    def each(&)
      return unless block_given?

      queue = Queue.new
      add_observer(proc { queue << _1 }, :call)

      while websocket.open?
        drain(queue, &)
        sleep(SLEEP_INTERVAL)
      end

      drain(queue, &)
    end

    # Disconnect from the PTY session
    #
    # @return [void]
    def disconnect = websocket.close

    private

    # @return [Symbol]
    attr_reader :status

    # @return [WebSocket::Client::Simple::Client]
    attr_reader :websocket

    # @return [Proc, Nil]
    attr_reader :handle_kill

    # @return [Proc, Nil]
    attr_reader :handle_resize

    # @return [Logger]
    attr_reader :logger

    # @return [void]
    def subscribe
      websocket.on(:open, &method(:on_websocket_open))
      websocket.on(:close, &method(:on_websocket_close))
      websocket.on(:message, &method(:on_websocket_message))
      websocket.on(:error, &method(:on_websocket_error))
    end

    # @return [void]
    def on_websocket_open
      logger.debug('[Websocket] open')
      @status = Status::OPEN
    end

    # @param error [Object, Nil]
    # @return [void]
    def on_websocket_close(error)
      logger.debug("[Websocket] close: #{error.inspect}")
      @status = Status::CLOSED
    end

    # @param error [WebSocket::Frame::Incoming::Client]
    # @return [void]
    def on_websocket_message(message)
      logger.debug("[Websocket] message(#{message.type}): #{message.data}")

      case message.type
      when :binary, :text
        process_websocket_text_message(message)
      when :close
        process_websocket_close_message(message)
      end
    end

    # @param error [Object]
    # @return [void]
    def on_websocket_error(error)
      logger.debug("[Websocket] error: #{error.inspect}")
      logger.debug("[Websocket] error: #{error.class}")
      @status = Status::ERROR
    end

    # @param message [WebSocket::Frame::Incoming::Client]
    # @return [void]
    def process_websocket_text_message(message)
      data = JSON.parse(message.data.to_s, symbolize_names: true)
      process_websocket_control_message(data) if data[:type] == WebSocketMessageType::CONTROL
    rescue JSON::ParserError, TypeError
      process_websocket_data_message(message.data.to_s)
    end

    # @param data [WebSocket::Frame::Data]
    # @return [void]
    def process_websocket_data_message(data)
      changed
      notify_observers(data)
    end

    # @param data [WebSocket::Frame::Data]
    # @return [void]
    def process_websocket_control_message(data) # rubocop:disable Metrics/MethodLength
      case data[:status]
      when WebSocketControlStatus::CONNECTED
        logger.debug('[control] connected')
        @status = Status::CONNECTED
      when WebSocketControlStatus::ERROR
        logger.debug("[control] error: #{error.inspect}")
        @status = Status::ERROR
        @error = data.fetch(:error, 'Unknown connection error')
      else
        websocket.close
        raise Sdk::Error, "Received invalid control message status: #{data[:status]}"
      end
    end

    # @param message [WebSocket::Frame::Incoming::Client]
    # @return [void]
    def process_websocket_close_message(message)
      data = JSON.parse(message.data.to_s, symbolize_names: true)
      @exit_code = data.fetch(:exitCode, nil)
      @error = data.fetch(:exitReason, nil)

      disconnect
    rescue JSON::ParserError, TypeError
      nil
    end

    # @param queue [Queue]
    # @yieldparam [WebSocket::Frame::Data]
    # @return [void]
    def drain(queue)
      data = nil

      yield data while (data = queue.pop(true))
    rescue ThreadError => _e
      nil
    end

    DEFAULT_TIMEOUT = 10.0
    private_constant :DEFAULT_TIMEOUT

    SLEEP_INTERVAL = 0.1
    private_constant :SLEEP_INTERVAL

    module Status
      ALL = [
        INIT = 'init',
        OPEN = 'open',
        CONNECTED = 'connected',
        CLOSED = 'closed',
        ERROR = 'error'
      ].freeze
    end

    module WebSocketMessageType
      ALL = [
        CONTROL = 'control'
      ].freeze
    end

    module WebSocketControlStatus
      ALL = [
        CONNECTED = 'connected',
        ERORR = 'error'
      ].freeze
    end
  end
end
