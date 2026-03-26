# frozen_string_literal: true

module Daytona
  class ComputerUse
    class Mouse
      include Instrumentation

      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      # @param otel_state [Daytona::OtelState, nil]
      def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
        @otel_state = otel_state
      end

      # Gets the current mouse cursor position.
      #
      # @return [DaytonaToolboxApiClient::MousePosition] Current mouse position with x and y coordinates
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   position = sandbox.computer_use.mouse.get_position
      #   puts "Mouse is at: #{position.x}, #{position.y}"
      def position
        toolbox_api.get_mouse_position
      rescue StandardError => e
        raise Sdk::Error, "Failed to get mouse position: #{e.message}"
      end

      # Moves the mouse cursor to the specified coordinates.
      #
      # @param x [Integer] The x coordinate to move to
      # @param y [Integer] The y coordinate to move to
      # @return [DaytonaToolboxApiClient::MouseMoveResponse] Move operation result
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   result = sandbox.computer_use.mouse.move(x: 100, y: 200)
      #   puts "Mouse moved to: #{result.x}, #{result.y}"
      def move(x:, y:)
        request = DaytonaToolboxApiClient::MouseMoveRequest.new(x:, y:)
        toolbox_api.move_mouse(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to move mouse: #{e.message}"
      end

      # Clicks the mouse at the specified coordinates.
      #
      # @param x [Integer] The x coordinate to click at
      # @param y [Integer] The y coordinate to click at
      # @param button [String] The mouse button to click ('left', 'right', 'middle'). Defaults to 'left'
      # @param double [Boolean] Whether to perform a double-click. Defaults to false
      # @return [DaytonaToolboxApiClient::MouseClickResponse] Click operation result
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Single left click
      #   result = sandbox.computer_use.mouse.click(x: 100, y: 200)
      #
      #   # Double click
      #   double_click = sandbox.computer_use.mouse.click(x: 100, y: 200, button: 'left', double: true)
      #
      #   # Right click
      #   right_click = sandbox.computer_use.mouse.click(x: 100, y: 200, button: 'right')
      def click(x:, y:, button: 'left', double: false)
        request = DaytonaToolboxApiClient::MouseClickRequest.new(x:, y:, button:, double:)
        toolbox_api.click(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to click mouse: #{e.message}"
      end

      # Drags the mouse from start coordinates to end coordinates.
      #
      # @param start_x [Integer] The starting x coordinate
      # @param start_y [Integer] The starting y coordinate
      # @param end_x [Integer] The ending x coordinate
      # @param end_y [Integer] The ending y coordinate
      # @param button [String] The mouse button to use for dragging. Defaults to 'left'
      # @return [DaytonaToolboxApiClient::MouseDragResponse] Drag operation result
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   result = sandbox.computer_use.mouse.drag(start_x: 50, start_y: 50, end_x: 150, end_y: 150)
      #   puts "Dragged from #{result.from_x},#{result.from_y} to #{result.to_x},#{result.to_y}"
      def drag(start_x:, start_y:, end_x:, end_y:, button: 'left')
        request = DaytonaToolboxApiClient::MouseDragRequest.new(start_x:, start_y:, end_x:, end_y:, button:)
        toolbox_api.drag(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to drag mouse: #{e.message}"
      end

      # Scrolls the mouse wheel at the specified coordinates.
      #
      # @param x [Integer] The x coordinate to scroll at
      # @param y [Integer] The y coordinate to scroll at
      # @param direction [String] The direction to scroll ('up' or 'down')
      # @param amount [Integer] The amount to scroll. Defaults to 1
      # @return [Boolean] Whether the scroll operation was successful
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Scroll up
      #   scroll_up = sandbox.computer_use.mouse.scroll(x: 100, y: 200, direction: 'up', amount: 3)
      #
      #   # Scroll down
      #   scroll_down = sandbox.computer_use.mouse.scroll(x: 100, y: 200, direction: 'down', amount: 5)
      def scroll(x:, y:, direction:, amount: 1)
        request = DaytonaToolboxApiClient::MouseScrollRequest.new(x:, y:, direction:, amount:)
        toolbox_api.scroll(request)
        true
      rescue StandardError => e
        raise Sdk::Error, "Failed to scroll mouse: #{e.message}"
      end

      instrument :position, :move, :click, :drag, :scroll, component: 'Mouse'

      private

      # @return [Daytona::OtelState, nil]
      attr_reader :otel_state
    end

    # Keyboard operations for computer use functionality.
    class Keyboard
      include Instrumentation

      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      # @param otel_state [Daytona::OtelState, nil]
      def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
        @otel_state = otel_state
      end

      # Types the specified text.
      #
      # @param text [String] The text to type
      # @param delay [Integer, nil] Delay between characters in milliseconds
      # @return [void]
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   sandbox.computer_use.keyboard.type("Hello, World!")
      #
      #   # With delay between characters
      #   sandbox.computer_use.keyboard.type("Slow typing", delay: 100)
      def type(text:, delay: nil)
        request = DaytonaToolboxApiClient::KeyboardTypeRequest.new(text:, delay:)
        toolbox_api.type_text(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to type text: #{e.message}"
      end

      # Presses a key with optional modifiers.
      #
      # @param key [String] The key to press (e.g., 'Enter', 'Escape', 'Tab', 'a', 'A')
      # @param modifiers [Array<String>, nil] Modifier keys ('ctrl', 'alt', 'meta', 'shift')
      # @return [void]
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Press Enter
      #   sandbox.computer_use.keyboard.press("Return")
      #
      #   # Press Ctrl+C
      #   sandbox.computer_use.keyboard.press("c", modifiers: ["ctrl"])
      #
      #   # Press Ctrl+Shift+T
      #   sandbox.computer_use.keyboard.press("t", modifiers: ["ctrl", "shift"])
      def press(key:, modifiers: nil)
        request = DaytonaToolboxApiClient::KeyboardPressRequest.new(key:, modifiers: modifiers || [])
        toolbox_api.press_key(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to press key: #{e.message}"
      end

      # Presses a hotkey combination.
      #
      # @param keys [String] The hotkey combination (e.g., 'ctrl+c', 'alt+tab', 'cmd+shift+t')
      # @return [void]
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Copy
      #   sandbox.computer_use.keyboard.hotkey("ctrl+c")
      #
      #   # Paste
      #   sandbox.computer_use.keyboard.hotkey("ctrl+v")
      #
      #   # Alt+Tab
      #   sandbox.computer_use.keyboard.hotkey("alt+tab")
      def hotkey(keys:)
        request = DaytonaToolboxApiClient::KeyboardHotkeyRequest.new(keys:)
        toolbox_api.press_hotkey(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to press hotkey: #{e.message}"
      end

      instrument :type, :press, :hotkey, component: 'Keyboard'

      private

      # @return [Daytona::OtelState, nil]
      attr_reader :otel_state
    end

    # Screenshot operations for computer use functionality.
    class Screenshot
      include Instrumentation

      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      # @param otel_state [Daytona::OtelState, nil]
      def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
        @otel_state = otel_state
      end

      # Takes a screenshot of the entire screen.
      #
      # @param show_cursor [Boolean] Whether to show the cursor in the screenshot. Defaults to false
      # @return [DaytonaApiClient::ScreenshotResponse] Screenshot data with base64 encoded image
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   screenshot = sandbox.computer_use.screenshot.take_full_screen
      #   puts "Screenshot size: #{screenshot.width}x#{screenshot.height}"
      #
      #   # With cursor visible
      #   with_cursor = sandbox.computer_use.screenshot.take_full_screen(show_cursor: true)
      def take_full_screen(show_cursor: false)
        toolbox_api.take_screenshot(show_cursor:)
      rescue StandardError => e
        raise Sdk::Error, "Failed to take screenshot: #{e.message}"
      end

      # Takes a screenshot of a specific region.
      #
      # @param region [ScreenshotRegion] The region to capture
      # @param show_cursor [Boolean] Whether to show the cursor in the screenshot. Defaults to false
      # @return [DaytonaApiClient::RegionScreenshotResponse] Screenshot data with base64 encoded image
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   region = ScreenshotRegion.new(x: 100, y: 100, width: 300, height: 200)
      #   screenshot = sandbox.computer_use.screenshot.take_region(region)
      #   puts "Captured region: #{screenshot.region.width}x#{screenshot.region.height}"
      def take_region(region:, show_cursor: false)
        toolbox_api.take_region_screenshot(region.height, region.width, region.y, region.x, show_cursor:)
      rescue StandardError => e
        raise Sdk::Error, "Failed to take region screenshot: #{e.message}"
      end

      # Takes a compressed screenshot of the entire screen.
      #
      # @param options [ScreenshotOptions, nil] Compression and display options
      # @return [DaytonaApiClient::CompressedScreenshotResponse] Compressed screenshot data
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Default compression
      #   screenshot = sandbox.computer_use.screenshot.take_compressed
      #
      #   # High quality JPEG
      #   jpeg = sandbox.computer_use.screenshot.take_compressed(
      #     options: ScreenshotOptions.new(format: "jpeg", quality: 95, show_cursor: true)
      #   )
      #
      #   # Scaled down PNG
      #   scaled = sandbox.computer_use.screenshot.take_compressed(
      #     options: ScreenshotOptions.new(format: "png", scale: 0.5)
      #   )
      def take_compressed(options: nil)
        options ||= ScreenshotOptions.new
        toolbox_api.take_compressed_screenshot(
          sandbox_id,
          scale: options.scale,
          quality: options.quality,
          format: options.fmt,
          show_cursor: options.show_cursor
        )
      rescue StandardError => e
        raise Sdk::Error, "Failed to take compressed screenshot: #{e.message}"
      end

      # Takes a compressed screenshot of a specific region.
      #
      # @param region [ScreenshotRegion] The region to capture
      # @param options [ScreenshotOptions, nil] Compression and display options
      # @return [DaytonaApiClient::CompressedScreenshotResponse] Compressed screenshot data
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   region = ScreenshotRegion.new(x: 0, y: 0, width: 800, height: 600)
      #   screenshot = sandbox.computer_use.screenshot.take_compressed_region(
      #     region,
      #     options: ScreenshotOptions.new(format: "webp", quality: 80, show_cursor: true)
      #   )
      #   puts "Compressed size: #{screenshot.size_bytes} bytes"
      def take_compressed_region(region:, options: nil)
        options ||= ScreenshotOptions.new
        toolbox_api.take_compressed_region_screenshot(
          sandbox_id,
          region.height,
          region.width,
          region.y,
          region.x,
          scale: options.scale,
          quality: options.quality,
          format: options.fmt,
          show_cursor: options.show_cursor
        )
      rescue StandardError => e
        raise Sdk::Error, "Failed to take compressed region screenshot: #{e.message}"
      end

      instrument :take_full_screen, :take_region, :take_compressed, :take_compressed_region,
                 component: 'Screenshot'

      private

      # @return [Daytona::OtelState, nil]
      attr_reader :otel_state
    end

    # Display operations for computer use functionality.
    class Display
      include Instrumentation

      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      # @param otel_state [Daytona::OtelState, nil]
      def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
        @otel_state = otel_state
      end

      # Gets information about the displays.
      #
      # @return [DaytonaToolboxApiClient::DisplayInfoResponse] Display information including primary display and all available displays
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   info = sandbox.computer_use.display.get_info
      #   puts "Primary display: #{info.primary_display.width}x#{info.primary_display.height}"
      #   puts "Total displays: #{info.total_displays}"
      #   info.displays.each_with_index do |display, i|
      #     puts "Display #{i}: #{display.width}x#{display.height} at #{display.x},#{display.y}"
      #   end
      def info
        toolbox_api.get_display_info
      rescue StandardError => e
        raise Sdk::Error, "Failed to get display info: #{e.message}"
      end

      # Gets the list of open windows.
      #
      # @return [DaytonaToolboxApiClient::WindowsResponse] List of open windows with their IDs and titles
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   windows = sandbox.computer_use.display.get_windows
      #   puts "Found #{windows.count} open windows:"
      #   windows.windows.each do |window|
      #     puts "- #{window.title} (ID: #{window.id})"
      #   end
      def windows
        toolbox_api.get_windows
      rescue StandardError => e
        raise Sdk::Error, "Failed to get windows: #{e.message}"
      end

      instrument :info, :windows, component: 'Display'

      private

      # @return [Daytona::OtelState, nil]
      attr_reader :otel_state
    end

    # Region coordinates for screenshot operations.
    class ScreenshotRegion
      # @return [Integer] X coordinate of the region
      attr_accessor :x

      # @return [Integer] Y coordinate of the region
      attr_accessor :y

      # @return [Integer] Width of the region
      attr_accessor :width

      # @return [Integer] Height of the region
      attr_accessor :height

      # @param x [Integer] X coordinate of the region
      # @param y [Integer] Y coordinate of the region
      # @param width [Integer] Width of the region
      # @param height [Integer] Height of the region
      def initialize(x:, y:, width:, height:)
        @x = x
        @y = y
        @width = width
        @height = height
      end
    end

    # Options for screenshot compression and display.
    class ScreenshotOptions
      # @return [Boolean, nil] Whether to show the cursor in the screenshot
      attr_accessor :show_cursor

      # @return [String, nil] Image format (e.g., 'png', 'jpeg', 'webp')
      attr_accessor :fmt

      # @return [Integer, nil] Compression quality (0-100)
      attr_accessor :quality

      # @return [Float, nil] Scale factor for the screenshot
      attr_accessor :scale

      # @param show_cursor [Boolean, nil] Whether to show the cursor in the screenshot
      # @param format [String, nil] Image format (e.g., 'png', 'jpeg', 'webp')
      # @param quality [Integer, nil] Compression quality (0-100)
      # @param scale [Float, nil] Scale factor for the screenshot
      def initialize(show_cursor: nil, format: nil, quality: nil, scale: nil)
        @show_cursor = show_cursor
        @fmt = format
        @quality = quality
        @scale = scale
      end
    end

    # Recording operations for computer use functionality.
    class Recording
      include Instrumentation

      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaToolboxApiClient::ComputerUseApi] API client for sandbox operations
      # @param otel_state [Daytona::OtelState, nil]
      def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
        @otel_state = otel_state
      end

      # Starts a new screen recording session.
      #
      # @param label [String, nil] Optional custom label for the recording
      # @return [DaytonaToolboxApiClient::Recording] Started recording details
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   # Start a recording with a label
      #   recording = sandbox.computer_use.recording.start(label: "my-test-recording")
      #   puts "Recording started: #{recording.id}"
      #   puts "File: #{recording.file_path}"
      def start(label: nil)
        request = DaytonaToolboxApiClient::StartRecordingRequest.new(label:)
        toolbox_api.start_recording(request: request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to start recording: #{e.message}"
      end

      # Stops an active screen recording session.
      #
      # @param id [String] The ID of the recording to stop
      # @return [DaytonaToolboxApiClient::Recording] Stopped recording details
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   result = sandbox.computer_use.recording.stop(id: recording.id)
      #   puts "Recording stopped: #{result.duration_seconds} seconds"
      #   puts "Saved to: #{result.file_path}"
      def stop(id:)
        request = DaytonaToolboxApiClient::StopRecordingRequest.new(id: id)
        toolbox_api.stop_recording(request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to stop recording: #{e.message}"
      end

      # Lists all recordings (active and completed).
      #
      # @return [DaytonaToolboxApiClient::ListRecordingsResponse] List of all recordings
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   recordings = sandbox.computer_use.recording.list
      #   puts "Found #{recordings.recordings.length} recordings"
      #   recordings.recordings.each do |rec|
      #     puts "- #{rec.file_name}: #{rec.status}"
      #   end
      def list
        toolbox_api.list_recordings
      rescue StandardError => e
        raise Sdk::Error, "Failed to list recordings: #{e.message}"
      end

      # Gets details of a specific recording by ID.
      #
      # @param id [String] The ID of the recording to retrieve
      # @return [DaytonaToolboxApiClient::Recording] Recording details
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   recording = sandbox.computer_use.recording.get(id: recording_id)
      #   puts "Recording: #{recording.file_name}"
      #   puts "Status: #{recording.status}"
      #   puts "Duration: #{recording.duration_seconds} seconds"
      def get(id:)
        toolbox_api.get_recording(id)
      rescue StandardError => e
        raise Sdk::Error, "Failed to get recording: #{e.message}"
      end

      # Deletes a recording by ID.
      #
      # @param id [String] The ID of the recording to delete
      # @return [void]
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   sandbox.computer_use.recording.delete(id: recording_id)
      #   puts "Recording deleted"
      def delete(id:)
        toolbox_api.delete_recording(id)
      rescue StandardError => e
        raise Sdk::Error, "Failed to delete recording: #{e.message}"
      end

      # Downloads a recording file and saves it to a local path.
      #
      # The file is streamed directly to disk without loading the entire content into memory.
      #
      # @param id [String] The ID of the recording to download
      # @param local_path [String] Path to save the recording file locally
      # @return [void]
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   sandbox.computer_use.recording.download(id: recording_id, local_path: "local_recording.mp4")
      #   puts "Recording downloaded"
      def download(id:, local_path:)
        require 'fileutils'
        require 'typhoeus'

        # Get the API configuration and build the download URL
        api_client = toolbox_api.api_client
        config = api_client.config
        base_url = config.base_url
        download_url = "#{base_url}/computeruse/recordings/#{id}/download"

        # Create parent directory if it doesn't exist
        parent_dir = File.dirname(local_path)
        FileUtils.mkdir_p(parent_dir) unless parent_dir.empty?

        # Stream the download directly to file
        file = File.open(local_path, 'wb')
        request = Typhoeus::Request.new(
          download_url,
          method: :get,
          headers: api_client.default_headers,
          timeout: config.timeout,
          ssl_verifypeer: config.verify_ssl,
          ssl_verifyhost: config.verify_ssl_host ? 2 : 0
        )

        # Stream chunks directly to file
        request.on_body do |chunk|
          file.write(chunk)
        end

        request.on_complete do |response|
          file.close
          unless response.success?
            File.delete(local_path) if File.exist?(local_path)
            raise Sdk::Error, "Failed to download recording: HTTP #{response.code}"
          end
        end

        request.run
      rescue StandardError => e
        file&.close
        File.delete(local_path) if File.exist?(local_path)
        raise Sdk::Error, "Failed to download recording: #{e.message}"
      end

      instrument :start, :stop, :list, :get, :delete, :download, component: 'Recording'

      private

      # @return [Daytona::OtelState, nil]
      attr_reader :otel_state
    end

    include Instrumentation

    # @return [String] The ID of the sandbox
    attr_reader :sandbox_id

    # @return [DaytonaApiClient::ToolboxApi] API client for sandbox operations
    attr_reader :toolbox_api

    # @return [Mouse] Mouse operations interface
    attr_reader :mouse

    # @return [Keyboard] Keyboard operations interface
    attr_reader :keyboard

    # @return [Screenshot] Screenshot operations interface
    attr_reader :screenshot

    # @return [Display] Display operations interface
    attr_reader :display

    # @return [Recording] Screen recording operations interface
    attr_reader :recording

    # Initialize a new ComputerUse instance.
    #
    # @param sandbox_id [String] The ID of the sandbox
    # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
    # @param otel_state [Daytona::OtelState, nil]
    def initialize(sandbox_id:, toolbox_api:, otel_state: nil)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @otel_state = otel_state
      @mouse = Mouse.new(sandbox_id:, toolbox_api:, otel_state:)
      @keyboard = Keyboard.new(sandbox_id:, toolbox_api:, otel_state:)
      @screenshot = Screenshot.new(sandbox_id:, toolbox_api:, otel_state:)
      @display = Display.new(sandbox_id:, toolbox_api:, otel_state:)
      @recording = Recording.new(sandbox_id:, toolbox_api:, otel_state:)
    end

    # Starts all computer use processes (Xvfb, xfce4, x11vnc, novnc).
    #
    # @return [DaytonaApiClient::ComputerUseStartResponse] Computer use start response
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   result = sandbox.computer_use.start
    #   puts "Computer use processes started: #{result.message}"
    def start
      toolbox_api.start_computer_use
    rescue StandardError => e
      raise Sdk::Error, "Failed to start computer use: #{e.message}"
    end

    # Stops all computer use processes.
    #
    # @return [DaytonaApiClient::ComputerUseStopResponse] Computer use stop response
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   result = sandbox.computer_use.stop
    #   puts "Computer use processes stopped: #{result.message}"
    def stop
      toolbox_api.stop_computer_use
    rescue StandardError => e
      raise Sdk::Error, "Failed to stop computer use: #{e.message}"
    end

    # Gets the status of all computer use processes.
    #
    # @return [DaytonaApiClient::ComputerUseStatusResponse] Status information about all VNC desktop processes
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   response = sandbox.computer_use.get_status
    #   puts "Computer use status: #{response.status}"
    def status
      toolbox_api.get_computer_use_status
    rescue StandardError => e
      raise Sdk::Error, "Failed to get computer use status: #{e.message}"
    end

    # Gets the status of a specific VNC process.
    #
    # @param process_name [String] Name of the process to check
    # @return [DaytonaApiClient::ProcessStatusResponse] Status information about the specific process
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   xvfb_status = sandbox.computer_use.get_process_status("xvfb")
    #   no_vnc_status = sandbox.computer_use.get_process_status("novnc")
    def get_process_status(process_name:)
      toolbox_api.get_process_status(process_name, sandbox_id)
    rescue StandardError => e
      raise Sdk::Error, "Failed to get process status: #{e.message}"
    end

    # Restarts a specific VNC process.
    #
    # @param process_name [String] Name of the process to restart
    # @return [DaytonaApiClient::ProcessRestartResponse] Process restart response
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   result = sandbox.computer_use.restart_process("xfce4")
    #   puts "XFCE4 process restarted: #{result.message}"
    def restart_process(process_name:)
      toolbox_api.restart_process(process_name, sandbox_id)
    rescue StandardError => e
      raise Sdk::Error, "Failed to restart process: #{e.message}"
    end

    # Gets logs for a specific VNC process.
    #
    # @param process_name [String] Name of the process to get logs for
    # @return [DaytonaApiClient::ProcessLogsResponse] Process logs
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   logs = sandbox.computer_use.get_process_logs("novnc")
    #   puts "NoVNC logs: #{logs}"
    def get_process_logs(process_name:)
      toolbox_api.get_process_logs(process_name, sandbox_id)
    rescue StandardError => e
      raise Sdk::Error, "Failed to get process logs: #{e.message}"
    end

    # Gets error logs for a specific VNC process.
    #
    # @param process_name [String] Name of the process to get error logs for
    # @return [DaytonaApiClient::ProcessErrorsResponse] Process error logs
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   errors = sandbox.computer_use.get_process_errors("x11vnc")
    #   puts "X11VNC errors: #{errors}"
    def get_process_errors(process_name:)
      toolbox_api.get_process_errors(process_name, sandbox_id)
    rescue StandardError => e
      raise Sdk::Error, "Failed to get process errors: #{e.message}"
    end

    instrument :start, :stop, :status, :get_process_status, :restart_process,
               :get_process_logs, :get_process_errors,
               component: 'ComputerUse'

    private

    # @return [Daytona::OtelState, nil]
    attr_reader :otel_state
  end
end
