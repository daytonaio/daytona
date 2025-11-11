import { useEffect, useRef } from 'react'
import { Detection } from '@/types/detection'
import { drawDetections } from '@/lib/utils'

interface DetectionCanvasProps {
  image: HTMLImageElement | HTMLVideoElement | null
  detections: Detection[]
  className?: string
}

export function DetectionCanvas({
  image,
  detections,
  className = '',
}: DetectionCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    if (!canvasRef.current || !image) return

    drawDetections(canvasRef.current, image, detections)
  }, [image, detections])

  return (
    <canvas
      ref={canvasRef}
      className={`max-w-full h-auto rounded-lg border-2 border-primary/30 ${className}`}
    />
  )
}
