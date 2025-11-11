import { useEffect, useRef, useState } from 'react'

interface WebcamCaptureProps {
  onFrame: (video: HTMLVideoElement) => void
  isActive: boolean
  fps?: number
}

export function WebcamCapture({
  onFrame,
  isActive,
  fps = 15,
}: WebcamCaptureProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [stream, setStream] = useState<MediaStream | null>(null)
  const [error, setError] = useState<string>('')
  const intervalRef = useRef<number | undefined>(undefined)

  useEffect(() => {
    if (isActive) {
      startWebcam()
    } else {
      stopWebcam()
    }

    return () => {
      stopWebcam()
    }
  }, [isActive])

  useEffect(() => {
    if (isActive && videoRef.current && stream) {
      intervalRef.current = window.setInterval(() => {
        if (videoRef.current && videoRef.current.readyState === 4) {
          onFrame(videoRef.current)
        }
      }, 1000 / fps)
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [isActive, stream, fps, onFrame])

  const startWebcam = async () => {
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({
        video: { width: 1280, height: 720 },
      })
      setStream(mediaStream)
      if (videoRef.current) {
        videoRef.current.srcObject = mediaStream
      }
      setError('')
    } catch (err) {
      setError('Failed to access webcam. Please check permissions.')
      console.error('Webcam error:', err)
    }
  }

  const stopWebcam = () => {
    if (stream) {
      stream.getTracks().forEach((track) => track.stop())
      setStream(null)
    }
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
    }
  }

  return (
    <div className="relative">
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className="w-full h-auto rounded-lg border-2 border-primary/30 bg-black"
      />
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/80 rounded-lg">
          <p className="text-red-400 text-center px-4">{error}</p>
        </div>
      )}
    </div>
  )
}
