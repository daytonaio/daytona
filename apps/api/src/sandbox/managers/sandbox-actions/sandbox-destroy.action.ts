import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SYNC_AGAIN } from '../sandbox.manager'
import { RunnerState } from '../../enums/runner-state.enum'

@Injectable()
export class SandboxDestroyAction extends SandboxAction {
  async run(sandbox: Sandbox) {
    if (sandbox.state === SandboxState.ARCHIVED) {
      await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    switch (sandbox.state) {
      case SandboxState.DESTROYED:
        return DONT_SYNC_AGAIN
      case SandboxState.DESTROYING: {
        // check if sandbox is destroyed
        try {
          const sandboxInfo = await runnerAdapter.info(sandbox.id)
          if (sandboxInfo.state === SandboxState.DESTROYED || sandboxInfo.state === SandboxState.ERROR) {
            await runnerAdapter.removeDestroyed(sandbox.id)
          }
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
        return SYNC_AGAIN
      }
      default: {
        // destroy sandbox
        try {
          const sandboxInfo = await runnerAdapter.info(sandbox.id)
          if (sandboxInfo?.state === SandboxState.DESTROYED) {
            return DONT_SYNC_AGAIN
          }
          await runnerAdapter.destroy(sandbox.id)
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYING)
        return SYNC_AGAIN
      }
    }
  }
}
