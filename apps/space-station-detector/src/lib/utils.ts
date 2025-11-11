import { Detection } from '@/types/detection'

export const CLASS_NAMES = ['Fire Extinguisher', 'Toolbox', 'Oxygen Tank']

export const CLASS_COLORS: Record<string, string> = {
  'Fire Extinguisher': '#ef4444',
  Toolbox: '#f59e0b',
  'Oxygen Tank': '#10b981',
}

export function drawDetections(
  canvas: HTMLCanvasElement,
  image: HTMLImageElement | HTMLVideoElement,
  detections: Detection[]
) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  canvas.width = image.width || (image as HTMLVideoElement).videoWidth
  canvas.height = image.height || (image as HTMLVideoElement).videoHeight

  ctx.drawImage(image, 0, 0, canvas.width, canvas.height)

  detections.forEach((detection) => {
    const { bbox, class: className, confidence } = detection
    const color = CLASS_COLORS[className] || '#3b82f6'

    ctx.strokeStyle = color
    ctx.lineWidth = 3
    ctx.strokeRect(bbox.x, bbox.y, bbox.width, bbox.height)

    ctx.fillStyle = color
    ctx.fillRect(bbox.x, bbox.y - 25, bbox.width, 25)

    ctx.fillStyle = '#ffffff'
    ctx.font = 'bold 14px Arial'
    ctx.fillText(
      `${className} ${(confidence * 100).toFixed(1)}%`,
      bbox.x + 5,
      bbox.y - 8
    )
  })
}

export function nonMaxSuppression(
  boxes: Detection[],
  iouThreshold: number = 0.45
): Detection[] {
  if (boxes.length === 0) return []

  const sorted = boxes.sort((a, b) => b.confidence - a.confidence)
  const selected: Detection[] = []

  while (sorted.length > 0) {
    const current = sorted.shift()!
    selected.push(current)

    for (let i = sorted.length - 1; i >= 0; i--) {
      const iou = calculateIoU(current.bbox, sorted[i].bbox)
      if (iou > iouThreshold) {
        sorted.splice(i, 1)
      }
    }
  }

  return selected
}

function calculateIoU(
  box1: Detection['bbox'],
  box2: Detection['bbox']
): number {
  const x1 = Math.max(box1.x, box2.x)
  const y1 = Math.max(box1.y, box2.y)
  const x2 = Math.min(box1.x + box1.width, box2.x + box2.width)
  const y2 = Math.min(box1.y + box1.height, box2.y + box2.height)

  const intersection = Math.max(0, x2 - x1) * Math.max(0, y2 - y1)
  const area1 = box1.width * box1.height
  const area2 = box2.width * box2.height
  const union = area1 + area2 - intersection

  return intersection / union
}

export function preprocessImage(
  image: HTMLImageElement | HTMLVideoElement,
  targetSize: number = 640
): { tensor: Float32Array; scale: number; offsetX: number; offsetY: number } {
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')!

  const imgWidth = image.width || (image as HTMLVideoElement).videoWidth
  const imgHeight = image.height || (image as HTMLVideoElement).videoHeight

  const scale = Math.min(targetSize / imgWidth, targetSize / imgHeight)
  const scaledWidth = Math.round(imgWidth * scale)
  const scaledHeight = Math.round(imgHeight * scale)

  const offsetX = Math.floor((targetSize - scaledWidth) / 2)
  const offsetY = Math.floor((targetSize - scaledHeight) / 2)

  canvas.width = targetSize
  canvas.height = targetSize

  ctx.fillStyle = '#000000'
  ctx.fillRect(0, 0, targetSize, targetSize)
  ctx.drawImage(image, offsetX, offsetY, scaledWidth, scaledHeight)

  const imageData = ctx.getImageData(0, 0, targetSize, targetSize)
  const pixels = imageData.data

  const tensor = new Float32Array(3 * targetSize * targetSize)

  for (let i = 0; i < pixels.length; i += 4) {
    const pixelIndex = i / 4
    tensor[pixelIndex] = pixels[i] / 255.0
    tensor[pixelIndex + targetSize * targetSize] = pixels[i + 1] / 255.0
    tensor[pixelIndex + 2 * targetSize * targetSize] = pixels[i + 2] / 255.0
  }

  return { tensor, scale, offsetX, offsetY }
}
