export interface Detection {
  class: string
  confidence: number
  bbox: {
    x: number
    y: number
    width: number
    height: number
  }
}

export interface ModelInfo {
  name: string
  mAP: number
  classes: string[]
}

export type DetectionMode = 'image' | 'webcam' | 'video'

export interface DetectionResult {
  detections: Detection[]
  inferenceTime: number
  fps?: number
}
