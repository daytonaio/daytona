import { Detection } from '@/types/detection'
import { CLASS_COLORS } from '@/lib/utils'

interface ResultsPanelProps {
  detections: Detection[]
  inferenceTime: number
  fps?: number
}

export function ResultsPanel({
  detections,
  inferenceTime,
  fps,
}: ResultsPanelProps) {
  const groupedDetections = detections.reduce(
    (acc, detection) => {
      if (!acc[detection.class]) {
        acc[detection.class] = []
      }
      acc[detection.class].push(detection)
      return acc
    },
    {} as Record<string, Detection[]>
  )

  return (
    <div className="bg-card rounded-lg p-6 border border-border">
      <h3 className="text-xl font-bold text-foreground mb-4">
        Detection Results
      </h3>

      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-muted rounded-lg p-4">
          <p className="text-muted-foreground text-sm">Inference Time</p>
          <p className="text-2xl font-bold text-primary">
            {inferenceTime.toFixed(0)}ms
          </p>
        </div>
        {fps !== undefined && (
          <div className="bg-muted rounded-lg p-4">
            <p className="text-muted-foreground text-sm">FPS</p>
            <p className="text-2xl font-bold text-secondary">{fps.toFixed(1)}</p>
          </div>
        )}
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h4 className="font-semibold text-foreground">
            Objects Detected: {detections.length}
          </h4>
        </div>

        {Object.entries(groupedDetections).map(([className, items]) => (
          <div key={className} className="space-y-2">
            <div className="flex items-center gap-2">
              <div
                className="w-4 h-4 rounded"
                style={{ backgroundColor: CLASS_COLORS[className] }}
              />
              <span className="font-medium text-foreground">
                {className} ({items.length})
              </span>
            </div>
            <div className="space-y-1 ml-6">
              {items.map((detection, idx) => (
                <div
                  key={idx}
                  className="text-sm text-muted-foreground flex justify-between"
                >
                  <span>Detection {idx + 1}</span>
                  <span className="font-mono">
                    {(detection.confidence * 100).toFixed(1)}%
                  </span>
                </div>
              ))}
            </div>
          </div>
        ))}

        {detections.length === 0 && (
          <p className="text-muted-foreground text-center py-8">
            No objects detected. Try adjusting the confidence threshold.
          </p>
        )}
      </div>
    </div>
  )
}
