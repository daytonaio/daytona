import { Button } from '@/components/ui/button'
import { Loader2, Play } from 'lucide-react'

type PlaygroundActionRunButtonProps = {
  isDisabled: boolean
  isRunning: boolean
  onRunActionClick: () => void
}

const PlaygroundActionRunButton: React.FC<PlaygroundActionRunButtonProps> = ({
  isDisabled,
  isRunning,
  onRunActionClick,
}) => {
  return (
    <div>
      <Button disabled={isDisabled} variant="outline" title="Run" onClick={onRunActionClick}>
        {isRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="w-4 h-4" />}
      </Button>
    </div>
  )
}

export default PlaygroundActionRunButton
