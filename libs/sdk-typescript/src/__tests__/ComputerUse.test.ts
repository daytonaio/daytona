import { createApiResponse } from './helpers'
import { ComputerUse } from '../ComputerUse'

const mockDynamicImport = jest.fn()

jest.mock('@daytonaio/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/Import', () => ({
  dynamicImport: (...args: unknown[]) => mockDynamicImport(...args),
}))

describe('ComputerUse', () => {
  const apiClient = {
    getMousePosition: jest.fn(),
    moveMouse: jest.fn(),
    click: jest.fn(),
    drag: jest.fn(),
    scroll: jest.fn(),
    typeText: jest.fn(),
    pressKey: jest.fn(),
    pressHotkey: jest.fn(),
    takeScreenshot: jest.fn(),
    takeRegionScreenshot: jest.fn(),
    takeCompressedScreenshot: jest.fn(),
    takeCompressedRegionScreenshot: jest.fn(),
    getDisplayInfo: jest.fn(),
    getWindows: jest.fn(),
    startRecording: jest.fn(),
    stopRecording: jest.fn(),
    listRecordings: jest.fn(),
    getRecording: jest.fn(),
    deleteRecording: jest.fn(),
    downloadRecording: jest.fn(),
    startComputerUse: jest.fn(),
    stopComputerUse: jest.fn(),
    getComputerUseStatus: jest.fn(),
    getProcessStatus: jest.fn(),
    restartProcess: jest.fn(),
    getProcessLogs: jest.fn(),
    getProcessErrors: jest.fn(),
  }

  const computerUse = new ComputerUse(apiClient as unknown as never)

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('delegates mouse/keyboard/screenshot/display operations', async () => {
    apiClient.getMousePosition.mockResolvedValue(createApiResponse({ x: 1, y: 2 }))
    apiClient.moveMouse.mockResolvedValue(createApiResponse({ x: 10, y: 20 }))
    apiClient.click.mockResolvedValue(createApiResponse({ success: true }))
    apiClient.drag.mockResolvedValue(createApiResponse({ success: true }))
    apiClient.scroll.mockResolvedValue(createApiResponse({ success: true }))
    apiClient.takeScreenshot.mockResolvedValue(createApiResponse({ data: 'a' }))
    apiClient.takeRegionScreenshot.mockResolvedValue(createApiResponse({ data: 'b' }))
    apiClient.takeCompressedScreenshot.mockResolvedValue(createApiResponse({ data: 'c' }))
    apiClient.takeCompressedRegionScreenshot.mockResolvedValue(createApiResponse({ data: 'd' }))
    apiClient.getDisplayInfo.mockResolvedValue(createApiResponse({ total_displays: 1 }))
    apiClient.getWindows.mockResolvedValue(createApiResponse({ count: 0, windows: [] }))

    await expect(computerUse.mouse.getPosition()).resolves.toEqual({ x: 1, y: 2 })
    await expect(computerUse.mouse.move(10, 20)).resolves.toEqual({ x: 10, y: 20 })
    await expect(computerUse.mouse.click(10, 20)).resolves.toEqual({ success: true })
    await expect(computerUse.mouse.drag(1, 2, 3, 4)).resolves.toEqual({ success: true })
    await expect(computerUse.mouse.scroll(1, 2, 'down')).resolves.toBe(true)

    await computerUse.keyboard.type('hello', 1)
    await computerUse.keyboard.press('Enter', ['ctrl'])
    await computerUse.keyboard.hotkey('ctrl+c')

    await expect(computerUse.screenshot.takeFullScreen()).resolves.toEqual({ data: 'a' })
    await expect(computerUse.screenshot.takeRegion({ x: 1, y: 2, width: 3, height: 4 })).resolves.toEqual({ data: 'b' })
    await expect(computerUse.screenshot.takeCompressed()).resolves.toEqual({ data: 'c' })
    await expect(
      computerUse.screenshot.takeCompressedRegion({ x: 1, y: 2, width: 3, height: 4 }, { format: 'png' }),
    ).resolves.toEqual({ data: 'd' })

    await expect(computerUse.display.getInfo()).resolves.toEqual({ total_displays: 1 })
    await expect(computerUse.display.getWindows()).resolves.toEqual({ count: 0, windows: [] })
  })

  it('delegates recording operations including download stream', async () => {
    apiClient.startRecording.mockResolvedValue(createApiResponse({ id: 'r1' }))
    apiClient.stopRecording.mockResolvedValue(createApiResponse({ id: 'r1', status: 'completed' }))
    apiClient.listRecordings.mockResolvedValue(createApiResponse({ recordings: [{ id: 'r1' }] }))
    apiClient.getRecording.mockResolvedValue(createApiResponse({ id: 'r1', status: 'completed' }))
    apiClient.deleteRecording.mockResolvedValue(createApiResponse(undefined))
    apiClient.downloadRecording.mockResolvedValue(createApiResponse({ pipe: jest.fn() }))

    const writer = {}
    const fsModule = {
      promises: { mkdir: jest.fn() },
      createWriteStream: jest.fn(() => writer),
    }
    const streamModule = { pipeline: jest.fn((_from: unknown, _to: unknown, cb: (err?: Error) => void) => cb()) }
    const utilModule = { promisify: (_f: unknown) => async () => undefined }

    mockDynamicImport.mockImplementation(async (moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'stream') return streamModule
      if (moduleName === 'util') return utilModule
      return {}
    })

    await expect(computerUse.recording.start('test')).resolves.toEqual({ id: 'r1' })
    await expect(computerUse.recording.stop('r1')).resolves.toEqual({ id: 'r1', status: 'completed' })
    await expect(computerUse.recording.list()).resolves.toEqual({ recordings: [{ id: 'r1' }] })
    await expect(computerUse.recording.get('r1')).resolves.toEqual({ id: 'r1', status: 'completed' })
    await computerUse.recording.delete('r1')
    await computerUse.recording.download('r1', '/tmp/recordings/r1.mp4')
  })

  it('delegates top-level computer-use process controls', async () => {
    apiClient.startComputerUse.mockResolvedValue(createApiResponse({ message: 'started' }))
    apiClient.stopComputerUse.mockResolvedValue(createApiResponse({ message: 'stopped' }))
    apiClient.getComputerUseStatus.mockResolvedValue(createApiResponse({ status: 'running' }))
    apiClient.getProcessStatus.mockResolvedValue(createApiResponse({ name: 'xvfb', running: true }))
    apiClient.restartProcess.mockResolvedValue(createApiResponse({ message: 'restarted' }))
    apiClient.getProcessLogs.mockResolvedValue(createApiResponse({ logs: 'ok' }))
    apiClient.getProcessErrors.mockResolvedValue(createApiResponse({ errors: '' }))

    await expect(computerUse.start()).resolves.toEqual({ message: 'started' })
    await expect(computerUse.stop()).resolves.toEqual({ message: 'stopped' })
    await expect(computerUse.getStatus()).resolves.toEqual({ status: 'running' })
    await expect(computerUse.getProcessStatus('xvfb')).resolves.toEqual({ name: 'xvfb', running: true })
    await expect(computerUse.restartProcess('xvfb')).resolves.toEqual({ message: 'restarted' })
    await expect(computerUse.getProcessLogs('xvfb')).resolves.toEqual({ logs: 'ok' })
    await expect(computerUse.getProcessErrors('xvfb')).resolves.toEqual({ errors: '' })
  })
})
