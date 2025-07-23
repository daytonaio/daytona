/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillingApiClient } from '@/billing-api/billingApiClient'
import {
  ApiKeysApi,
  Configuration,
  DockerRegistryApi,
  SnapshotsApi,
  OrganizationsApi,
  UsersApi,
  VolumesApi,
  SandboxApi,
  ToolboxApi,
  AuditApi,
  DaytonaConfiguration,
} from '@daytonaio/api-client'
import axios, { AxiosError } from 'axios'
import { DaytonaError } from './errors'

export class ApiClient {
  private config: Configuration
  private _snapshotApi: SnapshotsApi
  private _sandboxApi: SandboxApi
  private _userApi: UsersApi
  private _apiKeyApi: ApiKeysApi
  private _dockerRegistryApi: DockerRegistryApi
  private _organizationsApi: OrganizationsApi
  private _billingApi: BillingApiClient
  private _volumeApi: VolumesApi
  private _toolboxApi: ToolboxApi
  private _auditApi: AuditApi

  constructor(config: DaytonaConfiguration, accessToken: string) {
    this.config = new Configuration({
      basePath: import.meta.env.VITE_API_URL,
      accessToken: accessToken,
    })

    const axiosInstance = axios.create()
    axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        let errorMessage: string

        if (error instanceof AxiosError && error.message.includes('timeout of')) {
          errorMessage = 'Operation timed out'
        } else {
          errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)
        }

        throw DaytonaError.fromString(String(errorMessage))
      },
    )

    // Initialize APIs
    this._snapshotApi = new SnapshotsApi(this.config, undefined, axiosInstance)
    this._sandboxApi = new SandboxApi(this.config, undefined, axiosInstance)
    this._userApi = new UsersApi(this.config, undefined, axiosInstance)
    this._apiKeyApi = new ApiKeysApi(this.config, undefined, axiosInstance)
    this._dockerRegistryApi = new DockerRegistryApi(this.config, undefined, axiosInstance)
    this._organizationsApi = new OrganizationsApi(this.config, undefined, axiosInstance)
    this._billingApi = new BillingApiClient(config.billingApiUrl || window.location.origin, accessToken)
    this._volumeApi = new VolumesApi(this.config, undefined, axiosInstance)
    this._toolboxApi = new ToolboxApi(this.config, undefined, axiosInstance)
    this._auditApi = new AuditApi(this.config, undefined, axiosInstance)
  }

  public setAccessToken(accessToken: string) {
    this.config.accessToken = accessToken
  }

  public get snapshotApi() {
    return this._snapshotApi
  }

  public get sandboxApi() {
    return this._sandboxApi
  }

  public get userApi() {
    return this._userApi
  }

  public get apiKeyApi() {
    return this._apiKeyApi
  }

  public get dockerRegistryApi() {
    return this._dockerRegistryApi
  }

  public get organizationsApi() {
    return this._organizationsApi
  }

  public get billingApi() {
    return this._billingApi
  }

  public get volumeApi() {
    return this._volumeApi
  }

  public get toolboxApi() {
    return this._toolboxApi
  }

  public get auditApi() {
    return this._auditApi
  }
}
