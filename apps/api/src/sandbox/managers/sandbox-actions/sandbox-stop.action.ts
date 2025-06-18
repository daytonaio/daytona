import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SYNC_AGAIN } from '../sandbox.manager'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerSandboxState } from '../../runner-adapter/runnerAdapter'

@Injectable()
export class SandboxStopAction extends SandboxAction {
  async run(sandbox: Sandbox) {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    switch (sandbox.state) {
      case SandboxState.STARTED: {
        // stop sandbox
        const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
        await runnerAdapter.stop(sandbox.id)
        await this.updateSandboxState(sandbox.id, SandboxState.STOPPING)
        //  sync states again immediately for sandbox
        return SYNC_AGAIN
      }
      case SandboxState.STOPPING: {
        // check if sandbox is stopped
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
        const sandboxInfo = await runnerAdapter.info(sandbox.id)
        switch (sandboxInfo.state) {
          case RunnerSandboxState.STOPPED: {
            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.state = SandboxState.STOPPED
            sandboxToUpdate.backupState = BackupState.NONE
            await this.sandboxRepository.save(sandboxToUpdate)
            return SYNC_AGAIN
          }
          case RunnerSandboxState.ERROR: {
            await this.updateSandboxState(
              sandbox.id,
              SandboxState.ERROR,
              undefined,
              'Sandbox is in error state on runner',
            )
            return DONT_SYNC_AGAIN
          }
        }
        return SYNC_AGAIN
      }
      case SandboxState.ERROR: {
        if (sandbox.id.startsWith('err_')) {
          return DONT_SYNC_AGAIN
        }
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
        const sandboxInfo = await runnerAdapter.info(sandbox.id)
        if (sandboxInfo.state === RunnerSandboxState.STOPPED) {
          await this.updateSandboxState(sandbox.id, SandboxState.STOPPED)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }
}
