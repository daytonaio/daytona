import { Label } from '@/components/ui/label'
import { PlaygroundActionFormDataBasic } from '@/enums/Playground'
import PlaygroundActionRunButton from './ActionRunButton'

type PlaygroundActionFormProps<T> = {
  actionFormItem: PlaygroundActionFormDataBasic<T>
  onRunActionClick: () => Promise<void>
  runningActionMethodName: T | null
  actionError?: string
}

function PlaygroundActionForm<T>({
  actionFormItem,
  onRunActionClick,
  runningActionMethodName,
  actionError,
}: PlaygroundActionFormProps<T>) {
  const onRunActionButtonClick = async (actionMethodName: T) => {
    try {
      await onRunActionClick()
    } catch (error) {
      console.log('Action error', error)
    }
  }

  return (
    <>
      <div className="flex items-center justify-between">
        <div>
          <Label htmlFor={actionFormItem.methodName as string}>{actionFormItem.label}</Label>
          <p id={actionFormItem.methodName as string} className="text-sm text-muted-foreground mt-1 pl-1">
            {actionFormItem.description}
          </p>
        </div>
        <PlaygroundActionRunButton
          isDisabled={!!runningActionMethodName}
          isRunning={runningActionMethodName === actionFormItem.methodName}
          onRunActionClick={() => onRunActionButtonClick(actionFormItem.methodName)}
        />
      </div>
      <div>{actionError && <p className="text-sm text-red-500 mt-2">{actionError}</p>}</div>
    </>
  )
}

export default PlaygroundActionForm
