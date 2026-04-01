# frozen_string_literal: true

RSpec.describe Daytona::ComputerUse do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::ComputerUseApi) }
  let(:computer_use) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

  describe '#initialize' do
    it 'creates mouse, keyboard, screenshot, display, recording sub-objects' do
      expect(computer_use.mouse).to be_a(described_class::Mouse)
      expect(computer_use.keyboard).to be_a(described_class::Keyboard)
      expect(computer_use.screenshot).to be_a(described_class::Screenshot)
      expect(computer_use.display).to be_a(described_class::Display)
      expect(computer_use.recording).to be_a(described_class::Recording)
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
      expect { computer_use.start }.to raise_error(Daytona::Sdk::Error, /Failed to start computer use/)
    end
  end

  describe '#stop' do
    it 'delegates to toolbox_api.stop_computer_use' do
      result = double('StopResponse')
      allow(toolbox_api).to receive(:stop_computer_use).and_return(result)

      expect(computer_use.stop).to eq(result)
    end
  end

  describe '#status' do
    it 'returns computer use status' do
      status = double('StatusResponse')
      allow(toolbox_api).to receive(:get_computer_use_status).and_return(status)

      expect(computer_use.status).to eq(status)
    end
  end

  describe '#get_process_status' do
    it 'returns process status' do
      status = double('ProcessStatus')
      allow(toolbox_api).to receive(:get_process_status).with('xvfb', 'sandbox-123').and_return(status)

      expect(computer_use.get_process_status(process_name: 'xvfb')).to eq(status)
    end
  end

  describe '#restart_process' do
    it 'restarts a process' do
      result = double('RestartResponse')
      allow(toolbox_api).to receive(:restart_process).with('xfce4', 'sandbox-123').and_return(result)

      expect(computer_use.restart_process(process_name: 'xfce4')).to eq(result)
    end
  end

  describe '#get_process_logs' do
    it 'returns process logs' do
      logs = double('LogsResponse')
      allow(toolbox_api).to receive(:get_process_logs).with('novnc', 'sandbox-123').and_return(logs)

      expect(computer_use.get_process_logs(process_name: 'novnc')).to eq(logs)
    end
  end

  describe '#get_process_errors' do
    it 'returns process errors' do
      errors = double('ErrorsResponse')
      allow(toolbox_api).to receive(:get_process_errors).with('x11vnc', 'sandbox-123').and_return(errors)

      expect(computer_use.get_process_errors(process_name: 'x11vnc')).to eq(errors)
    end
  end

  describe Daytona::ComputerUse::Mouse do
    let(:mouse) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    describe '#position' do
      it 'returns mouse position' do
        pos = double('MousePosition', x: 100, y: 200)
        allow(toolbox_api).to receive(:get_mouse_position).and_return(pos)

        expect(mouse.position).to eq(pos)
      end
    end

    describe '#move' do
      it 'moves mouse to coordinates' do
        result = double('MoveResponse')
        allow(toolbox_api).to receive(:move_mouse).and_return(result)

        expect(mouse.move(x: 100, y: 200)).to eq(result)
      end
    end

    describe '#click' do
      it 'clicks at coordinates' do
        result = double('ClickResponse')
        allow(toolbox_api).to receive(:click).and_return(result)

        expect(mouse.click(x: 100, y: 200)).to eq(result)
      end

      it 'supports double click and right button' do
        result = double('ClickResponse')
        allow(toolbox_api).to receive(:click).and_return(result)

        mouse.click(x: 50, y: 50, button: 'right', double: true)
        expect(toolbox_api).to have_received(:click) do |req|
          expect(req.button).to eq('right')
          expect(req.double).to be(true)
        end
      end
    end

    describe '#drag' do
      it 'drags from start to end' do
        result = double('DragResponse')
        allow(toolbox_api).to receive(:drag).and_return(result)

        mouse.drag(start_x: 10, start_y: 20, end_x: 100, end_y: 200)
        expect(toolbox_api).to have_received(:drag)
      end
    end

    describe '#scroll' do
      it 'scrolls and returns true' do
        allow(toolbox_api).to receive(:scroll)

        expect(mouse.scroll(x: 100, y: 200, direction: 'up', amount: 3)).to be(true)
      end
    end
  end

  describe Daytona::ComputerUse::Keyboard do
    let(:keyboard) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    describe '#type' do
      it 'types text' do
        allow(toolbox_api).to receive(:type_text)

        keyboard.type(text: 'Hello')
        expect(toolbox_api).to have_received(:type_text) do |req|
          expect(req.text).to eq('Hello')
        end
      end
    end

    describe '#press' do
      it 'presses a key with modifiers' do
        allow(toolbox_api).to receive(:press_key)

        keyboard.press(key: 'c', modifiers: ['ctrl'])
        expect(toolbox_api).to have_received(:press_key) do |req|
          expect(req.key).to eq('c')
          expect(req.modifiers).to eq(['ctrl'])
        end
      end
    end

    describe '#hotkey' do
      it 'presses hotkey combination' do
        allow(toolbox_api).to receive(:press_hotkey)

        keyboard.hotkey(keys: 'ctrl+c')
        expect(toolbox_api).to have_received(:press_hotkey) do |req|
          expect(req.keys).to eq('ctrl+c')
        end
      end
    end
  end

  describe Daytona::ComputerUse::Screenshot do
    let(:screenshot) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    describe '#take_full_screen' do
      it 'takes a full screen screenshot' do
        result = double('ScreenshotResponse')
        allow(toolbox_api).to receive(:take_screenshot).with(show_cursor: false).and_return(result)

        expect(screenshot.take_full_screen).to eq(result)
      end

      it 'supports show_cursor option' do
        result = double('ScreenshotResponse')
        allow(toolbox_api).to receive(:take_screenshot).with(show_cursor: true).and_return(result)

        screenshot.take_full_screen(show_cursor: true)
      end
    end

    describe '#take_region' do
      it 'takes a region screenshot' do
        result = double('RegionScreenshotResponse')
        region = Daytona::ComputerUse::ScreenshotRegion.new(x: 10, y: 20, width: 300, height: 200)
        allow(toolbox_api).to receive(:take_region_screenshot)
          .with(200, 300, 20, 10, show_cursor: false)
          .and_return(result)

        expect(screenshot.take_region(region: region)).to eq(result)
      end
    end
  end

  describe Daytona::ComputerUse::Display do
    let(:display) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    describe '#info' do
      it 'returns display info' do
        info = double('DisplayInfoResponse')
        allow(toolbox_api).to receive(:get_display_info).and_return(info)

        expect(display.info).to eq(info)
      end
    end

    describe '#windows' do
      it 'returns window list' do
        windows = double('WindowsResponse')
        allow(toolbox_api).to receive(:get_windows).and_return(windows)

        expect(display.windows).to eq(windows)
      end
    end
  end

  describe Daytona::ComputerUse::Recording do
    let(:recording) { described_class.new(sandbox_id: 'sandbox-123', toolbox_api: toolbox_api) }

    describe '#start' do
      it 'starts recording' do
        result = double('Recording')
        allow(toolbox_api).to receive(:start_recording).and_return(result)

        expect(recording.start(label: 'my-recording')).to eq(result)
      end
    end

    describe '#stop' do
      it 'stops recording' do
        result = double('Recording')
        allow(toolbox_api).to receive(:stop_recording).and_return(result)

        expect(recording.stop(id: 'rec-1')).to eq(result)
      end
    end

    describe '#list' do
      it 'lists recordings' do
        result = double('ListRecordingsResponse')
        allow(toolbox_api).to receive(:list_recordings).and_return(result)

        expect(recording.list).to eq(result)
      end
    end

    describe '#get' do
      it 'gets a recording' do
        result = double('Recording')
        allow(toolbox_api).to receive(:get_recording).with('rec-1').and_return(result)

        expect(recording.get(id: 'rec-1')).to eq(result)
      end
    end

    describe '#delete' do
      it 'deletes a recording' do
        allow(toolbox_api).to receive(:delete_recording).with('rec-1')

        recording.delete(id: 'rec-1')
        expect(toolbox_api).to have_received(:delete_recording).with('rec-1')
      end
    end
  end

  describe Daytona::ComputerUse::ScreenshotRegion do
    it 'stores x, y, width, height' do
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
