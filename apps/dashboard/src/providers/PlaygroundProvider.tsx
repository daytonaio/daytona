/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  PlaygroundContext,
  SandboxParams,
  SetSandboxParamsValue,
  VNCInteractionOptionsParams,
  SetVNCInteractionOptionsParamValue,
  RunningActionMethodName,
  ActionRuntimeError,
  ValidatePlaygroundActionRequiredParams,
  RunPlaygroundActionBasic,
  RunPlaygroundActionWithParams,
  ValidatePlaygroundActionWithParams,
  PlaygroundActionParamValueSetter,
  SetPlaygroundActionParamValue,
} from '@/contexts/PlaygroundContext'
import { ScreenshotFormatOption, MouseButton, MouseScrollDirection } from '@/enums/Playground'
import {
  DEFAULT_CPU_RESOURCES,
  DEFAULT_MEMORY_RESOURCES,
  DEFAULT_DISK_RESOURCES,
  SANDBOX_SNAPSHOT_DEFAULT_VALUE,
} from '@/constants/Playground'
import {
  Daytona,
  Sandbox,
  CreateSandboxBaseParams,
  CreateSandboxFromImageParams,
  CreateSandboxFromSnapshotParams,
  Image,
} from '@daytonaio/sdk'
import { useAuth } from 'react-oidc-context'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { getLanguageCodeToRun, objectHasAnyValue } from '@/lib/playground'
import { useState, useMemo, useCallback } from 'react'

export const PlaygroundProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sandboxParametersState, setSandboxParametersState] = useState<SandboxParams>({
    snapshotName: SANDBOX_SNAPSHOT_DEFAULT_VALUE,
    resources: {
      cpu: DEFAULT_CPU_RESOURCES,
      memory: DEFAULT_MEMORY_RESOURCES,
      disk: DEFAULT_DISK_RESOURCES,
    },
    createSandboxBaseParams: {
      autoStopInterval: 15,
      autoArchiveInterval: 7,
      autoDeleteInterval: -1,
    },
    listFilesParams: {
      directoryPath: 'workspace/new-dir',
    },
    createFolderParams: {
      folderDestinationPath: 'workspace/new-dir',
      permissions: '755',
    },
    deleteFileParams: {
      filePath: 'workspace/new-dir',
      recursive: true,
    },
    gitCloneParams: {
      repositoryURL: 'https://github.com/octocat/Hello-World.git',
      cloneDestinationPath: 'workspace/repo',
    },
    gitStatusParams: {
      repositoryPath: 'workspace/repo',
    },
    gitBranchesParams: {
      repositoryPath: 'workspace/repo',
    },
    codeRunParams: {
      languageCode: getLanguageCodeToRun(),
    },
    shellCommandRunParams: {
      shellCommand: 'ls -la', // Current default and fixed value
    },
  })
  const [VNCInteractionOptionsParamsState, setVNCInteractionOptionsParamsState] = useState<VNCInteractionOptionsParams>(
    {
      keyboardHotKeyParams: { keys: '' },
      keyboardPressParams: { key: '' },
      keyboardTypeParams: { text: '' },
      mouseClickParams: {
        x: 100,
        y: 100,
        button: MouseButton.LEFT,
        double: false,
      },
      mouseDragParams: {
        startX: 100,
        startY: 100,
        endX: 200,
        endY: 200,
        button: MouseButton.LEFT,
      },
      mouseMoveParams: {
        x: 100,
        y: 100,
      },
      mouseScrollParams: {
        x: 100,
        y: 100,
        direction: MouseScrollDirection.DOWN,
        amount: 1,
      },
      screenshotOptionsConfig: {
        showCursor: false,
        format: ScreenshotFormatOption.PNG,
        quality: 100,
        scale: 1,
      },
      screenshotRegionConfig: {
        x: 100,
        y: 100,
        width: 300,
        height: 200,
      },
      VNCUrl: null,
    },
  )

  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()

  const setSandboxParameterValue: SetSandboxParamsValue = useCallback((key, value) => {
    setSandboxParametersState((prev) => ({ ...prev, [key]: value }))
  }, [])

  const setVNCInteractionOptionsParamValue: SetVNCInteractionOptionsParamValue = useCallback((key, value) => {
    setVNCInteractionOptionsParamsState((prev) => ({ ...prev, [key]: value }))
  }, [])

  const setPlaygroundActionParamValue: SetPlaygroundActionParamValue = useCallback(
    (key, value) => {
      if (key in sandboxParametersState) {
        setSandboxParameterValue(key as keyof SandboxParams, value as SandboxParams[keyof SandboxParams])
      } else if (key in VNCInteractionOptionsParamsState) {
        setVNCInteractionOptionsParamValue(
          key as keyof VNCInteractionOptionsParams,
          value as VNCInteractionOptionsParams[keyof VNCInteractionOptionsParams],
        )
      } else {
        console.error(`Unknown parameter key: ${String(key)}`)
      }
    },
    [
      setSandboxParameterValue,
      setVNCInteractionOptionsParamValue,
      sandboxParametersState,
      VNCInteractionOptionsParamsState,
    ],
  )

  const [runningActionMethod, setRunningActionMethod] = useState<RunningActionMethodName>(null)
  const [actionRuntimeError, setActionRuntimeError] = useState<ActionRuntimeError>({})
  const [sandbox, setSandbox] = useState<Sandbox | null>(null)

  const validatePlaygroundActionRequiredParams: ValidatePlaygroundActionRequiredParams = useCallback(
    (actionParamsFormData, actionParamsState) => {
      if (actionParamsFormData.some((formItem) => formItem.required)) {
        const emptyFormItem = actionParamsFormData
          .filter((formItem) => formItem.required)
          .find((formItem) => {
            const value = actionParamsState[formItem.key]
            return value === '' || value === undefined
          })

        if (emptyFormItem) {
          return `${emptyFormItem.label} parameter is required for this action`
        }
      }

      return undefined
    },
    [],
  )

  const runPlaygroundAction: RunPlaygroundActionBasic = useCallback(async (actionFormData, invokeApi) => {
    setRunningActionMethod(actionFormData.methodName)
    // Reset error if exists
    setActionRuntimeError((prev) => ({
      ...prev,
      [actionFormData.methodName]: undefined,
    }))
    try {
      await invokeApi(actionFormData)
    } catch (error: unknown) {
      console.error('API call error', error)
      setActionRuntimeError((prev) => ({
        ...prev,
        [actionFormData.methodName]: error instanceof Error ? error.message : String(error),
      }))
    } finally {
      setRunningActionMethod(null)
    }
  }, [])

  const runPlaygroundActionWithParams: RunPlaygroundActionWithParams = useCallback(
    async (actionFormData, invokeApi) => {
      const validationError = validatePlaygroundActionRequiredParams(
        actionFormData.parametersFormItems,
        actionFormData.parametersState,
      )
      if (validationError) {
        setActionRuntimeError((prev) => ({
          ...prev,
          [actionFormData.methodName]: validationError,
        }))
        setRunningActionMethod(null)
        return
      }
      return await runPlaygroundAction(actionFormData, invokeApi)
    },
    [runPlaygroundAction, validatePlaygroundActionRequiredParams],
  )

  const validatePlaygroundActionWithParams: ValidatePlaygroundActionWithParams = useCallback(
    (actionFormData, parametersState) => {
      const validationError = validatePlaygroundActionRequiredParams(
        actionFormData.parametersFormItems,
        parametersState,
      )
      if (validationError) {
        setActionRuntimeError((prev) => ({
          ...prev,
          [actionFormData.methodName]: validationError,
        }))
      } // Reset error
      else
        setActionRuntimeError((prev) => ({
          ...prev,
          [actionFormData.methodName]: undefined,
        }))
    },
    [validatePlaygroundActionRequiredParams],
  )

  const playgroundActionParamValueSetter: PlaygroundActionParamValueSetter = useCallback(
    (actionFormData, paramFormData, setState, actionParamsKey, value) => {
      setState((prev) => {
        const newState = { ...prev, [paramFormData.key]: value }
        setPlaygroundActionParamValue(actionParamsKey, newState)
        // Validate action params
        if (!actionFormData.onChangeParamsValidationDisabled)
          validatePlaygroundActionWithParams(actionFormData, newState)
        return newState
      })
    },
    [setPlaygroundActionParamValue, validatePlaygroundActionWithParams],
  )

  const DaytonaClient = useMemo(() => {
    if (!user?.access_token) return null
    return new Daytona({
      jwtToken: user.access_token,
      apiUrl: import.meta.env.VITE_API_URL,
      organizationId: selectedOrganization?.id,
    })
  }, [user?.access_token, selectedOrganization?.id])

  const getSandboxParametersInfo = useCallback(() => {
    const useLanguageParam = !!sandboxParametersState['language']
    const resourceValuesExist = objectHasAnyValue(sandboxParametersState['resources'])
    const useResourcesCPU = resourceValuesExist && sandboxParametersState['resources']['cpu'] !== undefined
    const useResourcesMemory = resourceValuesExist && sandboxParametersState['resources']['memory'] !== undefined
    const useResourcesDisk = resourceValuesExist && sandboxParametersState['resources']['disk'] !== undefined
    const useDefaultResourceValues = !(
      (useResourcesCPU && sandboxParametersState['resources']['cpu'] !== DEFAULT_CPU_RESOURCES) ||
      (useResourcesMemory && sandboxParametersState['resources']['memory'] !== DEFAULT_MEMORY_RESOURCES) ||
      (useResourcesDisk && sandboxParametersState['resources']['disk'] !== DEFAULT_DISK_RESOURCES)
    )

    const createSandboxParamsExist = objectHasAnyValue(sandboxParametersState['createSandboxBaseParams'])
    const useAutoStopInterval =
      createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] !== undefined
    const useAutoArchiveInterval =
      createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] !== undefined
    const useAutoDeleteInterval =
      createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] !== undefined

    const createSandboxFromImageParams: CreateSandboxFromImageParams = { image: Image.debianSlim('3.13') } // Default and fixed image if CreateSandboxFromImageParams are used
    const snapshotName = sandboxParametersState['snapshotName']
    const useCustomSandboxSnapshotName = snapshotName !== undefined && snapshotName !== SANDBOX_SNAPSHOT_DEFAULT_VALUE
    const createSandboxFromSnapshotParams: CreateSandboxFromSnapshotParams = {
      snapshot: useCustomSandboxSnapshotName ? snapshotName : undefined,
    }
    const createSandboxFromSnapshot = useCustomSandboxSnapshotName || useDefaultResourceValues

    // Create from base image if default resource values are not used
    // Snapshot parameter has precedence over resources and createSandboxFromImage
    const createSandboxFromImage = !useDefaultResourceValues && !createSandboxFromSnapshot

    // We specifiy resources for sandbox creation if there is any specificed resource value which has value different from the default one and createSandboxFromSnapshot is false
    const useResources = !createSandboxFromSnapshot && resourceValuesExist && !useDefaultResourceValues
    const useSandboxCreateParams =
      useLanguageParam ||
      useResources ||
      createSandboxParamsExist ||
      createSandboxFromSnapshot ||
      createSandboxFromImage

    if (createSandboxFromImage) {
      // Set CreateSandboxFromImageParams specific params
      if (useResources) {
        createSandboxFromImageParams.resources = {}
        if (useResourcesCPU) createSandboxFromImageParams.resources.cpu = sandboxParametersState['resources']['cpu']
        if (useResourcesMemory)
          createSandboxFromImageParams.resources.memory = sandboxParametersState['resources']['memory']
        if (useResourcesDisk) createSandboxFromImageParams.resources.disk = sandboxParametersState['resources']['disk']
      }
    }
    let createSandboxParams: CreateSandboxBaseParams | CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams =
      {}
    if (createSandboxFromSnapshot) createSandboxParams = createSandboxFromSnapshotParams
    else if (createSandboxFromImage) createSandboxParams = createSandboxFromImageParams
    // Set CreateSandboxBaseParams params which are common for both params types
    if (useLanguageParam) createSandboxParams.language = sandboxParametersState['language']
    if (useAutoStopInterval)
      createSandboxParams.autoStopInterval = sandboxParametersState['createSandboxBaseParams']['autoStopInterval']
    if (useAutoArchiveInterval)
      createSandboxParams.autoArchiveInterval = sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']
    if (useAutoDeleteInterval)
      createSandboxParams.autoDeleteInterval = sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']
    createSandboxParams.labels = { 'daytona-playground': 'true' }
    if (useLanguageParam)
      createSandboxParams.labels['daytona-playground-language'] = sandboxParametersState['language'] as string // useLanguageParam guarantes that value isn't undefined so we put as string to silence TS compiler
    return {
      useLanguageParam,
      useResources,
      useResourcesCPU,
      useResourcesMemory,
      useResourcesDisk,
      createSandboxParamsExist,
      useAutoStopInterval,
      useAutoArchiveInterval,
      useAutoDeleteInterval,
      useSandboxCreateParams,
      useCustomSandboxSnapshotName,
      createSandboxFromImage,
      createSandboxFromSnapshot,
      createSandboxParams,
    }
  }, [sandboxParametersState])

  return (
    <PlaygroundContext.Provider
      value={{
        sandboxParametersState,
        setSandboxParameterValue,
        VNCInteractionOptionsParamsState,
        setVNCInteractionOptionsParamValue,
        runPlaygroundActionWithParams,
        runPlaygroundActionWithoutParams: runPlaygroundAction,
        validatePlaygroundActionWithParams,
        playgroundActionParamValueSetter,
        runningActionMethod,
        actionRuntimeError,
        DaytonaClient,
        sandbox,
        setSandbox,
        getSandboxParametersInfo,
      }}
    >
      {children}
    </PlaygroundContext.Provider>
  )
}
