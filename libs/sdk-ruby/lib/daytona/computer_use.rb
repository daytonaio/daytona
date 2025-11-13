# frozen_string_literal: true

module Daytona
  class ComputerUse
    class Mouse
      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      def initialize(sandbox_id:, toolbox_api:)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
      end

      # Gets the current mouse cursor position.
      #
      # @return [DaytonaApiClient::MousePosition] Current mouse position with x and y coordinates
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   position = sandbox.computer_use.mouse.get_position
      #   puts "Mouse is at: #{position.x}, #{position.y}"
      def position
        toolbox_api.get_mouse_position(sandbox_id)
      rescue StandardError => e
        raise Sdk::Error, "Failed to get mouse position: #{e.message}"
      end

      # Moves the mouse cursor to the specified coordinates.
      #
      # @param x [Integer] The x coordinate to move to
      # @param y [Integer] The y coordinate to move to
      # @return [DaytonaApiClient::MouseMoveResponse] Move operation result
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   result = sandbox.computer_use.mouse.move(x: 100, y: 200)
      #   puts "Mouse moved to: #{result.x}, #{result.y}"
      def move(x:, y:) # rubocop:disable Naming/MethodParameterName
        request = DaytonaApiClient::MouseMoveRequest.new(x:, y:)
        toolbox_api.move_mouse(sandbox_id, request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to move mouse: #{e.message}"
      end

      # Clicks the mouse at the specified coordinates.
      #
      # @param x [Integer] The x coordinate to click at
      # @param y [Integer] The y coordinate to click at
      # @param button [String] The mouse button to click ('left', 'right', 'middle'). Defaults to 'left'
      # @param double [Boolean] Whether to perform a double-click. Defaults to false
      # @return [DaytonaApiClient::MouseClickResponse] Click operation result
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
      def click(x:, y:, button: 'left', double: false) # rubocop:disable Naming/MethodParameterName
        request = DaytonaApiClient::MouseClickRequest.new(x:, y:, button:, double:)
        toolbox_api.click_mouse(sandbox_id, request)
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
      # @return [DaytonaApiClient::MouseDragResponse] Drag operation result
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   result = sandbox.computer_use.mouse.drag(start_x: 50, start_y: 50, end_x: 150, end_y: 150)
      #   puts "Dragged from #{result.from_x},#{result.from_y} to #{result.to_x},#{result.to_y}"
      def drag(start_x:, start_y:, end_x:, end_y:, button: 'left')
        request = DaytonaApiClient::MouseDragRequest.new(start_x:, start_y:, end_x:, end_y:, button:)
        toolbox_api.drag_mouse(sandbox_id, request)
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
      def scroll(x:, y:, direction:, amount: 1) # rubocop:disable Naming/MethodParameterName
        request = DaytonaApiClient::MouseScrollRequest.new(x:, y:, direction:, amount:)
        toolbox_api.scroll_mouse(sandbox_id, request)
        true
      rescue StandardError => e
        raise Sdk::Error, "Failed to scroll mouse: #{e.message}"
      end
    end

    # Keyboard operations for computer use functionality.
    class Keyboard
      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      def initialize(sandbox_id:, toolbox_api:)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
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
        request = DaytonaApiClient::KeyboardTypeRequest.new(text:, delay:)
        toolbox_api.type_text(sandbox_id, request)
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
        request = DaytonaApiClient::KeyboardPressRequest.new(key:, modifiers: modifiers || [])
        toolbox_api.press_key(sandbox_id, request)
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
        request = DaytonaApiClient::KeyboardHotkeyRequest.new(keys:)
        toolbox_api.press_hotkey(sandbox_id, request)
      rescue StandardError => e
        raise Sdk::Error, "Failed to press hotkey: #{e.message}"
      end
    end

    # Screenshot operations for computer use functionality.
    class Screenshot
      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      def initialize(sandbox_id:, toolbox_api:)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
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
        toolbox_api.take_screenshot(sandbox_id, show_cursor:)
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
        toolbox_api.take_region_screenshot(sandbox_id, region.height, region.width, region.y, region.x, show_cursor:)
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
      def take_compressed_region(region:, options: nil) # rubocop:disable Metrics/MethodLength
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
    end

    # Display operations for computer use functionality.
    class Display
      # @return [String] The ID of the sandbox
      attr_reader :sandbox_id

      # @return [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      attr_reader :toolbox_api

      # @param sandbox_id [String] The ID of the sandbox
      # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
      def initialize(sandbox_id:, toolbox_api:)
        @sandbox_id = sandbox_id
        @toolbox_api = toolbox_api
      end

      # Gets information about the displays.
      #
      # @return [DaytonaApiClient::DisplayInfoResponse] Display information including primary display and all available displays
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
        toolbox_api.get_display_info(sandbox_id)
      rescue StandardError => e
        raise Sdk::Error, "Failed to get display info: #{e.message}"
      end

      # Gets the list of open windows.
      #
      # @return [DaytonaApiClient::WindowsResponse] List of open windows with their IDs and titles
      # @raise [Daytona::Sdk::Error] If the operation fails
      #
      # @example
      #   windows = sandbox.computer_use.display.get_windows
      #   puts "Found #{windows.count} open windows:"
      #   windows.windows.each do |window|
      #     puts "- #{window.title} (ID: #{window.id})"
      #   end
      def windows
        toolbox_api.get_windows(sandbox_id)
      rescue StandardError => e
        raise Sdk::Error, "Failed to get windows: #{e.message}"
      end
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
      def initialize(x:, y:, width:, height:) # rubocop:disable Naming/MethodParameterName
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

    # Initialize a new ComputerUse instance.
    #
    # @param sandbox_id [String] The ID of the sandbox
    # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for sandbox operations
    def initialize(sandbox_id:, toolbox_api:)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @mouse = Mouse.new(sandbox_id:, toolbox_api:)
      @keyboard = Keyboard.new(sandbox_id:, toolbox_api:)
      @screenshot = Screenshot.new(sandbox_id:, toolbox_api:)
      @display = Display.new(sandbox_id:, toolbox_api:)
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
      toolbox_api.start_computer_use(sandbox_id)
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
      toolbox_api.stop_computer_use(sandbox_id)
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
      toolbox_api.get_computer_use_status(sandbox_id)
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
  end
end
