/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BillingApiClient } from '@/billing-api/billingApiClient'
import {
  ApiKeysApi,
  Configuration,
  DockerRegistryApi,
  ImagesApi,
  OrganizationsApi,
  UsersApi,
  WorkspaceApi,
} from '@daytonaio/api-client'
import axios, { AxiosError } from 'axios'
import { DaytonaError } from './errors'

export class ApiClient {
  private config: Configuration
  private _imageApi: ImagesApi
  private _workspaceApi: WorkspaceApi
  private _userApi: UsersApi
  private _apiKeyApi: ApiKeysApi
  private _dockerRegistryApi: DockerRegistryApi
  private _organizationsApi: OrganizationsApi
  private _billingApi: BillingApiClient

  constructor(accessToken: string) {
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
    this._imageApi = new ImagesApi(this.config, undefined, axiosInstance)
    this._workspaceApi = new WorkspaceApi(this.config, undefined, axiosInstance)
    this._userApi = new UsersApi(this.config, undefined, axiosInstance)
    this._apiKeyApi = new ApiKeysApi(this.config, undefined, axiosInstance)
    this._dockerRegistryApi = new DockerRegistryApi(this.config, undefined, axiosInstance)
    this._organizationsApi = new OrganizationsApi(this.config, undefined, axiosInstance)
    this._billingApi = new BillingApiClient(import.meta.env.VITE_BILLING_API_URL || window.location.origin, accessToken)
  }

  public setAccessToken(accessToken: string) {
    this.config.accessToken = accessToken
  }

  public get imageApi() {
    return this._imageApi
  }

  public get workspaceApi() {
    return this._workspaceApi
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
}
