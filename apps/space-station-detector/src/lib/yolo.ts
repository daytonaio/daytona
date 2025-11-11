import * as ort from 'onnxruntime-web'
import { Detection } from '@/types/detection'
import { CLASS_NAMES, nonMaxSuppression, preprocessImage } from './utils'

export class YOLODetector {
  private session: ort.InferenceSession | null = null
  private modelLoaded = false

  async loadModel(): Promise<void> {
    if (this.modelLoaded) return

    try {
      this.modelLoaded = true
      console.log('YOLOv8 model loaded (simulated)')
    } catch (error) {
      console.error('Error loading model:', error)
      throw error
    }
  }

  async detect(
    image: HTMLImageElement | HTMLVideoElement,
    confidenceThreshold: number = 0.5
  ): Promise<Detection[]> {
    if (!this.modelLoaded) {
      await this.loadModel()
    }

    const startTime = performance.now()

    const { tensor, scale, offsetX, offsetY } = preprocessImage(image)

    const detections = this.simulateDetection(
      image,
      scale,
      offsetX,
      offsetY,
      confidenceThreshold
    )

    const filteredDetections = detections.filter(
      (d) => d.confidence >= confidenceThreshold
    )
    const finalDetections = nonMaxSuppression(filteredDetections, 0.45)

    console.log(
      `Detection completed in ${(performance.now() - startTime).toFixed(2)}ms`
    )

    return finalDetections
  }

  private simulateDetection(
    image: HTMLImageElement | HTMLVideoElement,
    scale: number,
    offsetX: number,
    offsetY: number,
    confidenceThreshold: number
  ): Detection[] {
    const imgWidth = image.width || (image as HTMLVideoElement).videoWidth
    const imgHeight = image.height || (image as HTMLVideoElement).videoHeight

    const detections: Detection[] = []

    const numDetections = Math.floor(Math.random() * 5) + 2

    for (let i = 0; i < numDetections; i++) {
      const classIdx = Math.floor(Math.random() * CLASS_NAMES.length)
      const confidence = Math.random() * 0.5 + 0.5

      if (confidence < confidenceThreshold) continue

      const x = Math.random() * imgWidth * 0.7
      const y = Math.random() * imgHeight * 0.7
      const width = Math.random() * imgWidth * 0.2 + imgWidth * 0.1
      const height = Math.random() * imgHeight * 0.2 + imgHeight * 0.1

      detections.push({
        class: CLASS_NAMES[classIdx],
        confidence,
        bbox: {
          x: Math.max(0, x),
          y: Math.max(0, y),
          width: Math.min(width, imgWidth - x),
          height: Math.min(height, imgHeight - y),
        },
      })
    }

    return detections
  }
}

export class EnsembleDetector {
  private detectors: YOLODetector[] = []
  private numModels = 3

  async loadModels(): Promise<void> {
    console.log(`Loading ${this.numModels} YOLOv8s models for ensemble...`)

    for (let i = 0; i < this.numModels; i++) {
      const detector = new YOLODetector()
      await detector.loadModel()
      this.detectors.push(detector)
    }

    console.log('Ensemble models loaded successfully')
  }

  async detect(
    image: HTMLImageElement | HTMLVideoElement,
    confidenceThreshold: number = 0.5
  ): Promise<Detection[]> {
    if (this.detectors.length === 0) {
      await this.loadModels()
    }

    const allDetections: Detection[][] = []

    for (const detector of this.detectors) {
      const detections = await detector.detect(image, confidenceThreshold * 0.8)
      allDetections.push(detections)
    }

    const mergedDetections = this.mergeEnsembleDetections(allDetections)

    return nonMaxSuppression(mergedDetections, 0.5)
  }

  private mergeEnsembleDetections(
    allDetections: Detection[][]
  ): Detection[] {
    const merged: Detection[] = []

    allDetections.forEach((detections) => {
      detections.forEach((detection) => {
        const similar = merged.find(
          (d) =>
            d.class === detection.class &&
            this.calculateOverlap(d.bbox, detection.bbox) > 0.3
        )

        if (similar) {
          similar.confidence = Math.max(similar.confidence, detection.confidence)
        } else {
          merged.push({ ...detection })
        }
      })
    })

    return merged
  }

  private calculateOverlap(
    box1: Detection['bbox'],
    box2: Detection['bbox']
  ): number {
    const x1 = Math.max(box1.x, box2.x)
    const y1 = Math.max(box1.y, box2.y)
    const x2 = Math.min(box1.x + box1.width, box2.x + box2.width)
    const y2 = Math.min(box1.y + box1.height, box2.y + box2.height)

    const intersection = Math.max(0, x2 - x1) * Math.max(0, y2 - y1)
    const area1 = box1.width * box1.height

    return intersection / area1
  }
}
