import { CLASS_COLORS } from '@/lib/utils'

export function ModelInfo() {
  return (
    <div className="bg-card rounded-lg p-6 border border-border">
      <h3 className="text-xl font-bold text-foreground mb-4">Model Information</h3>

      <div className="space-y-4">
        <div>
          <p className="text-sm text-muted-foreground">Architecture</p>
          <p className="text-lg font-semibold text-foreground">
            YOLOv8s Ensemble (3 models)
          </p>
        </div>

        <div>
          <p className="text-sm text-muted-foreground">Performance</p>
          <p className="text-lg font-semibold text-accent">
            mAP@0.5: 0.983 (98.3%)
          </p>
        </div>

        <div>
          <p className="text-sm text-muted-foreground mb-2">Detected Classes</p>
          <div className="space-y-2">
            {Object.entries(CLASS_COLORS).map(([className, color]) => (
              <div key={className} className="flex items-center gap-2">
                <div
                  className="w-4 h-4 rounded"
                  style={{ backgroundColor: color }}
                />
                <span className="text-foreground">{className}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="pt-4 border-t border-border">
          <p className="text-xs text-muted-foreground">
            This model uses a multi-model ensemble approach with three YOLOv8s
            models to achieve high accuracy in detecting space station inventory
            items.
          </p>
        </div>
      </div>
    </div>
  )
}
