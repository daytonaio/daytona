# frozen_string_literal: true

RSpec.describe Daytona::CodeInterpreter do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::InterpreterApi) }
  let(:get_preview_link) do
    proc { |_port| double('PreviewUrl', url: 'https://preview.example.com', token: 'tok') }
  end

  let(:interpreter) do
    described_class.new(
      sandbox_id: 'sandbox-123',
      toolbox_api: toolbox_api,
      get_preview_link: get_preview_link
    )
  end

  describe '#create_context' do
    it 'creates an interpreter context' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:create_interpreter_context).and_return(ctx)

      result = interpreter.create_context
      expect(result).to eq(ctx)
    end

    it 'passes cwd option' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:create_interpreter_context).and_return(ctx)

      interpreter.create_context(cwd: '/workspace')
      expect(toolbox_api).to have_received(:create_interpreter_context) do |req|
        expect(req.cwd).to eq('/workspace')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:create_interpreter_context).and_raise(StandardError, 'err')
      expect { interpreter.create_context }.to raise_error(Daytona::Sdk::Error, /Failed to create interpreter context/)
    end
  end

  describe '#list_contexts' do
    it 'returns array of contexts' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      response = double('ListResponse', contexts: [ctx])
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_return(response)

      result = interpreter.list_contexts
      expect(result).to eq([ctx])
    end

    it 'returns empty array when nil contexts' do
      response = double('ListResponse', contexts: nil)
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_return(response)

      expect(interpreter.list_contexts).to eq([])
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_raise(StandardError, 'err')
      expect { interpreter.list_contexts }.to raise_error(Daytona::Sdk::Error, /Failed to list interpreter/)
    end
  end

  describe '#delete_context' do
    it 'deletes a context and returns nil' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:delete_interpreter_context).with('ctx-1')

      expect(interpreter.delete_context(ctx)).to be_nil
    end

    it 'wraps errors' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:delete_interpreter_context).and_raise(StandardError, 'err')
      expect { interpreter.delete_context(ctx) }.to raise_error(Daytona::Sdk::Error, /Failed to delete interpreter/)
    end
  end
end
