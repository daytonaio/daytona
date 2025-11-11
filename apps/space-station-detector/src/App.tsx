import { useState, useRef, useCallback, useEffect } from 'react'
import { DetectionCanvas } from '@/components/DetectionCanvas'
import { FileUpload } from '@/components/FileUpload'
import { WebcamCapture } from '@/components/WebcamCapture'
import { ResultsPanel } from '@/components/ResultsPanel'
import { ModelInfo } from '@/components/ModelInfo'
import { EnsembleDetector } from '@/lib/yolo'
import { Detection, DetectionMode } from '@/types/detection'

function App() {
  const [mode, setMode] = useState<DetectionMode>('image')
  const [detections, setDetections] = useState<Detection[]>([])
  const [inferenceTime, setInferenceTime] = useState<number>(0)
  const [fps, setFps] = useState<number>(0)
  const [isProcessing, setIsProcessing] = useState(false)
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.5)
  const [currentImage, setCurrentImage] = useState<HTMLImageElement | null>(null)
  const [isWebcamActive, setIsWebcamActive] = useState(false)

  const detectorRef = useRef<EnsembleDetector | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const fpsCounterRef = useRef<{ frames: number; lastTime: number }>({
    frames: 0,
    lastTime: Date.now(),
  })

  useEffect(() => {
    detectorRef.current = new EnsembleDetector()
    detectorRef.current.loadModels()
  }, [])

  const handleFileSelect = async (file: File) => {
    setIsProcessing(true)
    setDetections([])

    const img = new Image()
    img.onload = async () => {
      setCurrentImage(img)
      await processImage(img)
    }
    img.src = URL.createObjectURL(file)
  }

  const processImage = async (image: HTMLImageElement) => {
    if (!detectorRef.current) return

    const startTime = performance.now()
    const results = await detectorRef.current.detect(image, confidenceThreshold)
    const endTime = performance.now()

    setDetections(results)
    setInferenceTime(endTime - startTime)
    setIsProcessing(false)
  }

  const handleWebcamFrame = useCallback(
    async (video: HTMLVideoElement) => {
      if (!detectorRef.current || isProcessing) return

      videoRef.current = video
      setIsProcessing(true)

      const startTime = performance.now()
      const results = await detectorRef.current.detect(video, confidenceThreshold)
      const endTime = performance.now()

      setDetections(results)
      setInferenceTime(endTime - startTime)

      fpsCounterRef.current.frames++
      const now = Date.now()
      if (now - fpsCounterRef.current.lastTime >= 1000) {
        setFps(fpsCounterRef.current.frames)
        fpsCounterRef.current.frames = 0
        fpsCounterRef.current.lastTime = now
      }

      setIsProcessing(false)
    },
    [confidenceThreshold, isProcessing]
  )

  const toggleWebcam = () => {
    setIsWebcamActive(!isWebcamActive)
    if (!isWebcamActive) {
      setDetections([])
      setCurrentImage(null)
    }
  }

  const handleModeChange = (newMode: DetectionMode) => {
    setMode(newMode)
    setDetections([])
    setCurrentImage(null)
    if (newMode !== 'webcam') {
      setIsWebcamActive(false)
    }
  }

  return (
    <div className="min-h-screen bg-space-gradient text-foreground">
      <div className="container mx-auto px-4 py-8">
        <header className="text-center mb-12">
          <div className="flex items-center justify-center gap-4 mb-4">
            <div className="w-16 h-16 bg-gradient-to-br from-primary to-secondary rounded-full flex items-center justify-center">
              <svg
                className="w-10 h-10 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
                />
              </svg>
            </div>
          </div>
          <h1 className="text-5xl font-bold bg-gradient-to-r from-primary via-secondary to-accent bg-clip-text text-transparent mb-2">
            Space Station Inventory Detection
          </h1>
          <p className="text-muted-foreground text-lg">
            AI-Powered Detection System using YOLOv8s Ensemble
          </p>
        </header>

        <div className="mb-8">
          <div className="flex gap-4 justify-center mb-6">
            <button
              onClick={() => handleModeChange('image')}
              className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                mode === 'image'
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              }`}
            >
              Image Upload
            </button>
            <button
              onClick={() => handleModeChange('webcam')}
              className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                mode === 'webcam'
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              }`}
            >
              Webcam Detection
            </button>
          </div>

          <div className="max-w-md mx-auto mb-6">
            <label className="block text-sm font-medium text-foreground mb-2">
              Confidence Threshold: {(confidenceThreshold * 100).toFixed(0)}%
            </label>
            <input
              type="range"
              min="0"
              max="100"
              value={confidenceThreshold * 100}
              onChange={(e) => setConfidenceThreshold(Number(e.target.value) / 100)}
              className="w-full h-2 bg-muted rounded-lg appearance-none cursor-pointer accent-primary"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2">
            <div className="bg-card rounded-lg p-6 border border-border">
              {mode === 'image' && (
                <div className="space-y-6">
                  <FileUpload
                    onFileSelect={handleFileSelect}
                    disabled={isProcessing}
                  />
                  {currentImage && (
                    <div className="mt-6">
                      <DetectionCanvas
                        image={currentImage}
                        detections={detections}
                      />
                    </div>
                  )}
                  {!currentImage && (
                    <div className="text-center py-16 text-muted-foreground">
                      <svg
                        className="w-24 h-24 mx-auto mb-4 opacity-50"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                        />
                      </svg>
                      <p className="text-lg">Upload an image to start detection</p>
                    </div>
                  )}
                </div>
              )}

              {mode === 'webcam' && (
                <div className="space-y-6">
                  <button
                    onClick={toggleWebcam}
                    className={`w-full px-6 py-4 font-semibold rounded-lg transition-all ${
                      isWebcamActive
                        ? 'bg-red-500 hover:bg-red-600 text-white'
                        : 'bg-accent hover:bg-accent/90 text-white'
                    }`}
                  >
                    {isWebcamActive ? 'Stop Webcam' : 'Start Webcam'}
                  </button>

                  {isWebcamActive ? (
                    <div className="relative">
                      <WebcamCapture
                        onFrame={handleWebcamFrame}
                        isActive={isWebcamActive}
                        fps={15}
                      />
                      {videoRef.current && (
                        <div className="absolute inset-0 pointer-events-none">
                          <DetectionCanvas
                            image={videoRef.current}
                            detections={detections}
                            className="absolute inset-0"
                          />
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="text-center py-16 text-muted-foreground">
                      <svg
                        className="w-24 h-24 mx-auto mb-4 opacity-50"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
                        />
                      </svg>
                      <p className="text-lg">Click to start webcam detection</p>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>

          <div className="space-y-6">
            <ModelInfo />
            {(detections.length > 0 || inferenceTime > 0) && (
              <ResultsPanel
                detections={detections}
                inferenceTime={inferenceTime}
                fps={mode === 'webcam' ? fps : undefined}
              />
            )}
          </div>
        </div>

        <footer className="mt-16 text-center text-muted-foreground text-sm">
          <p>
            Powered by YOLOv8s Ensemble • mAP@0.5: 0.983 • Detecting Fire
            Extinguishers, Toolboxes, and Oxygen Tanks
          </p>
        </footer>
      </div>
    </div>
  )
}

export default App
