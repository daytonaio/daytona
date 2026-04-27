# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::LspServer do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::LspApi) }
  let(:lsp) do
    described_class.new(
      language_id: :python,
      path_to_project: '/workspace',
      toolbox_api: toolbox_api,
      sandbox_id: 'sandbox-123'
    )
  end

  describe '#initialize' do
    it 'stores language_id, path_to_project, and sandbox_id' do
      expect(lsp.language_id).to eq(:python)
      expect(lsp.path_to_project).to eq('/workspace')
      expect(lsp.sandbox_id).to eq('sandbox-123')
    end
  end

  describe '#start' do
    it 'starts the LSP server' do
      allow(toolbox_api).to receive(:start)

      lsp.start

      expect(toolbox_api).to have_received(:start) do |req|
        expect(req.language_id).to eq(:python)
        expect(req.path_to_project).to eq('/workspace')
      end
    end
  end

  describe '#stop' do
    it 'stops the LSP server' do
      allow(toolbox_api).to receive(:stop)

      lsp.stop

      expect(toolbox_api).to have_received(:stop) do |req|
        expect(req.language_id).to eq(:python)
        expect(req.path_to_project).to eq('/workspace')
      end
    end
  end

  describe '#did_open' do
    it 'notifies server of file open' do
      allow(toolbox_api).to receive(:did_open)

      lsp.did_open('/workspace/main.py')

      expect(toolbox_api).to have_received(:did_open) do |req|
        expect(req.uri).to eq('file:///workspace/main.py')
      end
    end
  end

  describe '#did_close' do
    it 'notifies server of file close' do
      allow(toolbox_api).to receive(:did_close)

      lsp.did_close('/workspace/main.py')

      expect(toolbox_api).to have_received(:did_close) do |req|
        expect(req.uri).to eq('file:///workspace/main.py')
      end
    end
  end

  describe '#completions' do
    it 'returns completion list for a position' do
      completions = double('CompletionList')
      allow(toolbox_api).to receive(:completions).and_return(completions)

      result = lsp.completions(path: '/workspace/main.py',
                               position: Daytona::LspServer::Position.new(line: 10,
                                                                          character: 5))

      expect(result).to eq(completions)
      expect(toolbox_api).to have_received(:completions) do |req|
        expect(req.language_id).to eq(:python)
        expect(req.path_to_project).to eq('/workspace')
        expect(req.uri).to eq('file:///workspace/main.py')
        expect(req.position.line).to eq(10)
        expect(req.position.character).to eq(5)
      end
    end
  end

  describe '#document_symbols' do
    it 'returns document symbols' do
      symbols = [double('LspSymbol')]
      allow(toolbox_api).to receive(:document_symbols).with(:python, '/workspace',
                                                            'file:///workspace/main.py').and_return(symbols)

      expect(lsp.document_symbols('/workspace/main.py')).to eq(symbols)
    end
  end

  describe '#sandbox_symbols' do
    it 'returns workspace symbols matching a query' do
      symbols = [double('LspSymbol')]
      allow(toolbox_api).to receive(:workspace_symbols).with('MyClass', :python, '/workspace').and_return(symbols)

      expect(lsp.sandbox_symbols('MyClass')).to eq(symbols)
    end
  end

  describe 'Language constants' do
    it 'defines supported language constants' do
      expect(Daytona::LspServer::Language::ALL).to eq(%i[javascript python typescript])
    end
  end

  describe 'Position' do
    it 'is a data class with line and character' do
      pos = Daytona::LspServer::Position.new(line: 5, character: 10)

      expect(pos.line).to eq(5)
      expect(pos.character).to eq(10)
    end
  end
end
