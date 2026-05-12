# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::ComputerUse do
  let(:toolbox_api) { double('ComputerUseApi') }
  let(:computer_use) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

  describe '#initialize' do
    it 'creates mouse, keyboard, screenshot, display, and recording helpers' do
      expect(computer_use.mouse).to be_a(described_class::Mouse)
      expect(computer_use.keyboard).to be_a(described_class::Keyboard)
      expect(computer_use.screenshot).to be_a(described_class::Screenshot)
      expect(computer_use.display).to be_a(described_class::Display)
      expect(computer_use.recording).to be_a(described_class::Recording)
      expect(computer_use.accessibility).to be_a(described_class::Accessibility)
    end
  end

  describe '#start' do
    it 'delegates to toolbox_api.start_computer_use' do
      result = double('StartResponse')
      allow(toolbox_api).to receive(:start_computer_use).and_return(result)

      expect(computer_use.start).to eq(result)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:start_computer_use).and_raise(StandardError, 'err')

      expect { computer_use.start }.to raise_error(Daytona::Sdk::Error, /Failed to start computer use: err/)
    end
  end

  describe '#stop' do
    it 'delegates to toolbox_api.stop_computer_use' do
      result = double('StopResponse')
      allow(toolbox_api).to receive(:stop_computer_use).and_return(result)

      expect(computer_use.stop).to eq(result)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:stop_computer_use).and_raise(StandardError, 'err')

      expect { computer_use.stop }.to raise_error(Daytona::Sdk::Error, /Failed to stop computer use: err/)
    end
  end

  describe '#status' do
    it 'returns computer use status' do
      status = double('StatusResponse')
      allow(toolbox_api).to receive(:get_computer_use_status).and_return(status)

      expect(computer_use.status).to eq(status)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_computer_use_status).and_raise(StandardError, 'err')

      expect { computer_use.status }.to raise_error(Daytona::Sdk::Error, /Failed to get computer use status: err/)
    end
  end

  describe '#get_process_status' do
    it 'returns process status' do
      status = double('ProcessStatus')
      allow(toolbox_api).to receive(:get_process_status).with('xvfb', 'sandbox-123').and_return(status)

      expect(computer_use.get_process_status(process_name: 'xvfb')).to eq(status)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_process_status).and_raise(StandardError, 'err')

      expect { computer_use.get_process_status(process_name: 'xvfb') }
        .to raise_error(Daytona::Sdk::Error, /Failed to get process status: err/)
    end
  end

  describe '#restart_process' do
    it 'restarts a process' do
      result = double('RestartResponse')
      allow(toolbox_api).to receive(:restart_process).with('xfce4', 'sandbox-123').and_return(result)

      expect(computer_use.restart_process(process_name: 'xfce4')).to eq(result)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:restart_process).and_raise(StandardError, 'err')

      expect { computer_use.restart_process(process_name: 'xfce4') }
        .to raise_error(Daytona::Sdk::Error, /Failed to restart process: err/)
    end
  end

  describe '#get_process_logs' do
    it 'returns process logs' do
      logs = double('LogsResponse')
      allow(toolbox_api).to receive(:get_process_logs).with('novnc', 'sandbox-123').and_return(logs)

      expect(computer_use.get_process_logs(process_name: 'novnc')).to eq(logs)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_process_logs).and_raise(StandardError, 'err')

      expect { computer_use.get_process_logs(process_name: 'novnc') }
        .to raise_error(Daytona::Sdk::Error, /Failed to get process logs: err/)
    end
  end

  describe '#get_process_errors' do
    it 'returns process errors' do
      errors = double('ErrorsResponse')
      allow(toolbox_api).to receive(:get_process_errors).with('x11vnc', 'sandbox-123').and_return(errors)

      expect(computer_use.get_process_errors(process_name: 'x11vnc')).to eq(errors)
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:get_process_errors).and_raise(StandardError, 'err')

      expect { computer_use.get_process_errors(process_name: 'x11vnc') }
        .to raise_error(Daytona::Sdk::Error, /Failed to get process errors: err/)
    end
  end

  describe Daytona::ComputerUse::Mouse do
    let(:mouse) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    it 'returns the current mouse position' do
      position = double('MousePosition', x: 100, y: 200)
      allow(toolbox_api).to receive(:get_mouse_position).and_return(position)

      expect(mouse.position).to eq(position)
    end

    it 'wraps position errors' do
      allow(toolbox_api).to receive(:get_mouse_position).and_raise(StandardError, 'err')

      expect { mouse.position }.to raise_error(Daytona::Sdk::Error, /Failed to get mouse position: err/)
    end

    it 'moves the mouse to coordinates' do
      result = double('MoveResponse')
      allow(toolbox_api).to receive(:move_mouse).and_return(result)

      expect(mouse.move(x: 100, y: 200)).to eq(result)
      expect(toolbox_api).to have_received(:move_mouse) do |req|
        expect(req.x).to eq(100)
        expect(req.y).to eq(200)
      end
    end

    it 'wraps move errors' do
      allow(toolbox_api).to receive(:move_mouse).and_raise(StandardError, 'err')

      expect { mouse.move(x: 1, y: 2) }.to raise_error(Daytona::Sdk::Error, /Failed to move mouse: err/)
    end

    it 'clicks with button and double-click settings' do
      result = double('ClickResponse')
      allow(toolbox_api).to receive(:click).and_return(result)

      expect(mouse.click(x: 50, y: 60, button: 'right', double: true)).to eq(result)
      expect(toolbox_api).to have_received(:click) do |req|
        expect(req.x).to eq(50)
        expect(req.y).to eq(60)
        expect(req.button).to eq('right')
        expect(req.double).to be(true)
      end
    end

    it 'wraps click errors' do
      allow(toolbox_api).to receive(:click).and_raise(StandardError, 'err')

      expect { mouse.click(x: 1, y: 2) }.to raise_error(Daytona::Sdk::Error, /Failed to click mouse: err/)
    end

    it 'drags from start to end coordinates' do
      result = double('DragResponse')
      allow(toolbox_api).to receive(:drag).and_return(result)

      expect(mouse.drag(start_x: 10, start_y: 20, end_x: 100, end_y: 200)).to eq(result)
      expect(toolbox_api).to have_received(:drag) do |req|
        expect(req.start_x).to eq(10)
        expect(req.end_y).to eq(200)
      end
    end

    it 'wraps drag errors' do
      allow(toolbox_api).to receive(:drag).and_raise(StandardError, 'err')

      expect { mouse.drag(start_x: 1, start_y: 2, end_x: 3, end_y: 4) }
        .to raise_error(Daytona::Sdk::Error, /Failed to drag mouse: err/)
    end

    it 'scrolls and returns true' do
      allow(toolbox_api).to receive(:scroll)

      expect(mouse.scroll(x: 100, y: 200, direction: 'up', amount: 3)).to be(true)
      expect(toolbox_api).to have_received(:scroll) do |req|
        expect(req.direction).to eq('up')
        expect(req.amount).to eq(3)
      end
    end

    it 'wraps scroll errors' do
      allow(toolbox_api).to receive(:scroll).and_raise(StandardError, 'err')

      expect { mouse.scroll(x: 1, y: 2, direction: 'down') }
        .to raise_error(Daytona::Sdk::Error, /Failed to scroll mouse: err/)
    end
  end

  describe Daytona::ComputerUse::Keyboard do
    let(:keyboard) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    it 'types text with an optional delay' do
      allow(toolbox_api).to receive(:type_text)

      keyboard.type(text: 'Hello', delay: 25)

      expect(toolbox_api).to have_received(:type_text) do |req|
        expect(req.text).to eq('Hello')
        expect(req.delay).to eq(25)
      end
    end

    it 'wraps type errors' do
      allow(toolbox_api).to receive(:type_text).and_raise(StandardError, 'err')

      expect { keyboard.type(text: 'Hello') }.to raise_error(Daytona::Sdk::Error, /Failed to type text: err/)
    end

    it 'presses a key with default empty modifiers' do
      allow(toolbox_api).to receive(:press_key)

      keyboard.press(key: 'Enter')

      expect(toolbox_api).to have_received(:press_key) do |req|
        expect(req.key).to eq('Enter')
        expect(req.modifiers).to eq([])
      end
    end

    it 'wraps press errors' do
      allow(toolbox_api).to receive(:press_key).and_raise(StandardError, 'err')

      expect { keyboard.press(key: 'c', modifiers: ['ctrl']) }
        .to raise_error(Daytona::Sdk::Error, /Failed to press key: err/)
    end

    it 'presses a hotkey combination' do
      allow(toolbox_api).to receive(:press_hotkey)

      keyboard.hotkey(keys: 'ctrl+c')

      expect(toolbox_api).to have_received(:press_hotkey) do |req|
        expect(req.keys).to eq('ctrl+c')
      end
    end

    it 'wraps hotkey errors' do
      allow(toolbox_api).to receive(:press_hotkey).and_raise(StandardError, 'err')

      expect { keyboard.hotkey(keys: 'ctrl+c') }.to raise_error(Daytona::Sdk::Error, /Failed to press hotkey: err/)
    end
  end

  describe Daytona::ComputerUse::Screenshot do
    let(:screenshot) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }
    let(:region) { Daytona::ComputerUse::ScreenshotRegion.new(x: 10, y: 20, width: 300, height: 200) }

    it 'takes a full screen screenshot' do
      result = double('ScreenshotResponse')
      allow(toolbox_api).to receive(:take_screenshot).with(show_cursor: true).and_return(result)

      expect(screenshot.take_full_screen(show_cursor: true)).to eq(result)
    end

    it 'wraps full screen screenshot errors' do
      allow(toolbox_api).to receive(:take_screenshot).and_raise(StandardError, 'err')

      expect { screenshot.take_full_screen }.to raise_error(Daytona::Sdk::Error, /Failed to take screenshot: err/)
    end

    it 'takes a region screenshot' do
      result = double('RegionScreenshotResponse')
      allow(toolbox_api).to receive(:take_region_screenshot)
        .with(200, 300, 20, 10, show_cursor: false)
        .and_return(result)

      expect(screenshot.take_region(region: region)).to eq(result)
    end

    it 'wraps region screenshot errors' do
      allow(toolbox_api).to receive(:take_region_screenshot).and_raise(StandardError, 'err')

      expect { screenshot.take_region(region: region) }
        .to raise_error(Daytona::Sdk::Error, /Failed to take region screenshot: err/)
    end

    it 'takes a compressed screenshot with default options' do
      result = double('CompressedScreenshotResponse')
      allow(toolbox_api).to receive(:take_compressed_screenshot).and_return(result)

      expect(screenshot.take_compressed).to eq(result)
      expect(toolbox_api).to have_received(:take_compressed_screenshot)
        .with('sandbox-123', scale: nil, quality: nil, format: nil, show_cursor: nil)
    end

    it 'takes a compressed region screenshot with explicit options' do
      result = double('CompressedRegionScreenshotResponse')
      options = Daytona::ComputerUse::ScreenshotOptions.new(show_cursor: true, format: 'jpeg', quality: 90, scale: 0.5)
      allow(toolbox_api).to receive(:take_compressed_region_screenshot).and_return(result)

      expect(screenshot.take_compressed_region(region: region, options: options)).to eq(result)
      expect(toolbox_api).to have_received(:take_compressed_region_screenshot)
        .with('sandbox-123', 200, 300, 20, 10, scale: 0.5, quality: 90, format: 'jpeg', show_cursor: true)
    end

    it 'wraps compressed screenshot errors' do
      allow(toolbox_api).to receive(:take_compressed_screenshot).and_raise(StandardError, 'err')

      expect do
        screenshot.take_compressed
      end.to raise_error(Daytona::Sdk::Error, /Failed to take compressed screenshot: err/)
    end

    it 'wraps compressed region screenshot errors' do
      allow(toolbox_api).to receive(:take_compressed_region_screenshot).and_raise(StandardError, 'err')

      expect { screenshot.take_compressed_region(region: region) }
        .to raise_error(Daytona::Sdk::Error, /Failed to take compressed region screenshot: err/)
    end
  end

  describe Daytona::ComputerUse::Display do
    let(:display) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    it 'returns display info' do
      info = double('DisplayInfoResponse')
      allow(toolbox_api).to receive(:get_display_info).and_return(info)

      expect(display.info).to eq(info)
    end

    it 'wraps info errors' do
      allow(toolbox_api).to receive(:get_display_info).and_raise(StandardError, 'err')

      expect { display.info }.to raise_error(Daytona::Sdk::Error, /Failed to get display info: err/)
    end

    it 'returns window list' do
      windows = double('WindowsResponse')
      allow(toolbox_api).to receive(:get_windows).and_return(windows)

      expect(display.windows).to eq(windows)
    end

    it 'wraps window errors' do
      allow(toolbox_api).to receive(:get_windows).and_raise(StandardError, 'err')

      expect { display.windows }.to raise_error(Daytona::Sdk::Error, /Failed to get windows: err/)
    end
  end

  describe Daytona::ComputerUse::Accessibility do
    let(:accessibility) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    it 'gets the accessibility tree with optional query values' do
      result = double('AccessibilityTreeResponse')
      allow(toolbox_api).to receive(:get_accessibility_tree).and_return(result)

      expect(accessibility.get_tree).to eq(result)
      expect(accessibility.get_tree(scope: 'pid', pid: 123, max_depth: 0)).to eq(result)
      expect(toolbox_api).to have_received(:get_accessibility_tree).with({})
      expect(toolbox_api).to have_received(:get_accessibility_tree).with(scope: 'pid', pid: 123, max_depth: 0)
    end

    it 'finds accessibility nodes with generated request fields' do
      result = double('AccessibilityNodesResponse')
      allow(toolbox_api).to receive(:find_accessibility_nodes).and_return(result)

      expect(accessibility.find_nodes(
               scope: 'all',
               role: 'button',
               name: 'Submit',
               name_match: 'exact',
               states: ['visible'],
               limit: 0
             )).to eq(result)
      expect(toolbox_api).to have_received(:find_accessibility_nodes) do |req|
        expect(req.scope).to eq('all')
        expect(req.role).to eq('button')
        expect(req.name).to eq('Submit')
        expect(req.name_match).to eq('exact')
        expect(req.states).to eq(['visible'])
        expect(req.limit).to eq(0)
      end
    end

    it 'delegates accessibility node actions' do
      allow(toolbox_api).to receive(:focus_accessibility_node)
      allow(toolbox_api).to receive(:invoke_accessibility_node)
      allow(toolbox_api).to receive(:set_accessibility_node_value)

      accessibility.focus_node(id: 'node-1')
      accessibility.invoke_node(id: 'node-2', action: 'click')
      accessibility.set_node_value(id: 'node-3', value: 'hello')

      expect(toolbox_api).to have_received(:focus_accessibility_node) do |req|
        expect(req.id).to eq('node-1')
      end
      expect(toolbox_api).to have_received(:invoke_accessibility_node) do |req|
        expect(req.id).to eq('node-2')
        expect(req.action).to eq('click')
      end
      expect(toolbox_api).to have_received(:set_accessibility_node_value) do |req|
        expect(req.id).to eq('node-3')
        expect(req.value).to eq('hello')
      end
    end

    it 'wraps accessibility errors' do
      allow(toolbox_api).to receive(:get_accessibility_tree).and_raise(StandardError, 'err')

      expect { accessibility.get_tree }
        .to raise_error(Daytona::Sdk::Error, /Failed to get accessibility tree: err/)
    end
  end

  describe Daytona::ComputerUse::Recording do
    let(:recording) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }
    let(:api_client) do
      double('ApiClient', default_headers: { 'Authorization' => 'Bearer token' },
                          config: double(base_url: 'https://toolbox.example.com', timeout: 10, verify_ssl: true, verify_ssl_host: true))
    end

    before do
      allow(toolbox_api).to receive(:api_client).and_return(api_client)
    end

    it 'starts recording with an optional label' do
      result = double('Recording')
      allow(toolbox_api).to receive(:start_recording).and_return(result)

      expect(recording.start(label: 'my-recording')).to eq(result)
      expect(toolbox_api).to have_received(:start_recording) do |request:|
        expect(request.label).to eq('my-recording')
      end
    end

    it 'wraps start errors' do
      allow(toolbox_api).to receive(:start_recording).and_raise(StandardError, 'err')

      expect { recording.start }.to raise_error(Daytona::Sdk::Error, /Failed to start recording: err/)
    end

    it 'stops recording' do
      result = double('Recording')
      allow(toolbox_api).to receive(:stop_recording).and_return(result)

      expect(recording.stop(id: 'rec-1')).to eq(result)
      expect(toolbox_api).to have_received(:stop_recording) do |request|
        expect(request.id).to eq('rec-1')
      end
    end

    it 'wraps stop errors' do
      allow(toolbox_api).to receive(:stop_recording).and_raise(StandardError, 'err')

      expect { recording.stop(id: 'rec-1') }.to raise_error(Daytona::Sdk::Error, /Failed to stop recording: err/)
    end

    it 'lists recordings' do
      result = double('ListRecordingsResponse')
      allow(toolbox_api).to receive(:list_recordings).and_return(result)

      expect(recording.list).to eq(result)
    end

    it 'wraps list errors' do
      allow(toolbox_api).to receive(:list_recordings).and_raise(StandardError, 'err')

      expect { recording.list }.to raise_error(Daytona::Sdk::Error, /Failed to list recordings: err/)
    end

    it 'gets a recording' do
      result = double('Recording')
      allow(toolbox_api).to receive(:get_recording).with('rec-1').and_return(result)

      expect(recording.get(id: 'rec-1')).to eq(result)
    end

    it 'wraps get errors' do
      allow(toolbox_api).to receive(:get_recording).and_raise(StandardError, 'err')

      expect { recording.get(id: 'rec-1') }.to raise_error(Daytona::Sdk::Error, /Failed to get recording: err/)
    end

    it 'deletes a recording' do
      allow(toolbox_api).to receive(:delete_recording).with('rec-1')

      recording.delete(id: 'rec-1')

      expect(toolbox_api).to have_received(:delete_recording).with('rec-1')
    end

    it 'wraps delete errors' do
      allow(toolbox_api).to receive(:delete_recording).and_raise(StandardError, 'err')

      expect { recording.delete(id: 'rec-1') }.to raise_error(Daytona::Sdk::Error, /Failed to delete recording: err/)
    end

    it 'downloads a recording to disk' do
      request_class = Class.new do
        class << self
          attr_accessor :instance
        end

        attr_reader :body_callback, :complete_callback

        def initialize(*, **)
          self.class.instance = self
        end

        def on_body(&block)
          @body_callback = block
        end

        def on_complete(&block)
          @complete_callback = block
        end

        def run
          body_callback.call('video-data')
          complete_callback.call(Class.new do
            def success? = true
            def code = 200
          end.new)
        end
      end

      stub_const('Typhoeus::Request', request_class)

      Dir.mktmpdir do |dir|
        path = File.join(dir, 'recordings', 'session.mp4')

        recording.download(id: 'rec-1', local_path: path)

        expect(File.binread(path)).to eq('video-data')
      end
    end

    it 'removes partial files when the download fails' do
      request_class = Class.new do
        class << self
          attr_accessor :instance
        end

        attr_reader :body_callback, :complete_callback

        def initialize(*, **)
          self.class.instance = self
        end

        def on_body(&block)
          @body_callback = block
        end

        def on_complete(&block)
          @complete_callback = block
        end

        def run
          body_callback.call('partial')
          complete_callback.call(Class.new do
            def success? = false
            def code = 500
          end.new)
        end
      end

      stub_const('Typhoeus::Request', request_class)

      Dir.mktmpdir do |dir|
        path = File.join(dir, 'session.mp4')

        expect { recording.download(id: 'rec-1', local_path: path) }
          .to raise_error(Daytona::Sdk::Error, /Failed to download recording: Failed to download recording: HTTP 500/)
        expect(File.exist?(path)).to be(false)
      end
    end
  end

  describe Daytona::ComputerUse::ScreenshotRegion do
    it 'stores x, y, width, and height' do
      region = described_class.new(x: 10, y: 20, width: 300, height: 200)

      expect(region.x).to eq(10)
      expect(region.y).to eq(20)
      expect(region.width).to eq(300)
      expect(region.height).to eq(200)
    end
  end

  describe Daytona::ComputerUse::ScreenshotOptions do
    it 'stores options with defaults' do
      opts = described_class.new

      expect(opts.show_cursor).to be_nil
      expect(opts.fmt).to be_nil
      expect(opts.quality).to be_nil
      expect(opts.scale).to be_nil
    end

    it 'accepts all options' do
      opts = described_class.new(show_cursor: true, format: 'jpeg', quality: 90, scale: 0.5)

      expect(opts.show_cursor).to be(true)
      expect(opts.fmt).to eq('jpeg')
      expect(opts.quality).to eq(90)
      expect(opts.scale).to eq(0.5)
    end
  end
end
