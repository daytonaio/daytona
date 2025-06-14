/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { VolumeDto, VolumesApi } from '@daytonaio/api-client'
import { DaytonaNotFoundError } from './errors/DaytonaError'

/**
 * Represents a Daytona Volume which is a shared storage volume for Sandboxes.
 *
 * @property {string} id - Unique identifier for the Volume
 * @property {string} name - Name of the Volume
 * @property {string} organizationId - Organization ID that owns the Volume
 * @property {string} state - Current state of the Volume
 * @property {string} createdAt - Date and time when the Volume was created
 * @property {string} updatedAt - Date and time when the Volume was last updated
 * @property {string} lastUsedAt - Date and time when the Volume was last used
 */
export type Volume = VolumeDto & { __brand: 'Volume' }

/**
 * Service for managing Daytona Volumes.
 *
 * This service provides methods to list, get, create, and delete Volumes.
 *
 * @class
 */
export class VolumeService {
  constructor(private volumesApi: VolumesApi) {}

  /**
   * Lists all available Volumes.
   *
   * @returns {Promise<Volume[]>} List of all Volumes accessible to the user
   *
   * @example
   * const daytona = new Daytona();
   * const volumes = await daytona.volume.list();
   * console.log(`Found ${volumes.length} volumes`);
   * volumes.forEach(vol => console.log(`${vol.name} (${vol.id})`));
   */
  async list(): Promise<Volume[]> {
    const response = await this.volumesApi.listVolumes()
    return response.data as Volume[]
  }

  /**
   * Gets a Volume by its name.
   *
   * @param {string} name - Name of the Volume to retrieve
   * @param {boolean} create - Whether to create the Volume if it does not exist
   * @returns {Promise<Volume>} The requested Volume
   * @throws {Error} If the Volume does not exist or cannot be accessed
   *
   * @example
   * const daytona = new Daytona();
   * const volume = await daytona.volume.get("volume-name", true);
   * console.log(`Volume ${volume.name} is in state ${volume.state}`);
   */
  async get(name: string, create = false): Promise<Volume> {
    try {
      const response = await this.volumesApi.getVolumeByName(name)
      return response.data as Volume
    } catch (error) {
      if (
        error instanceof DaytonaNotFoundError &&
        create &&
        error.message.match(/Volume with name ([\w-]+) not found/)
      ) {
        return await this.create(name)
      }
      throw error
    }
  }

  /**
   * Creates a new Volume with the specified name.
   *
   * @param {string} name - Name for the new Volume
   * @returns {Promise<Volume>} The newly created Volume
   * @throws {Error} If the Volume cannot be created
   *
   * @example
   * const daytona = new Daytona();
   * const volume = await daytona.volume.create("my-data-volume");
   * console.log(`Created volume ${volume.name} with ID ${volume.id}`);
   */
  async create(name: string): Promise<Volume> {
    const response = await this.volumesApi.createVolume({ name })
    return response.data as Volume
  }

  /**
   * Deletes a Volume.
   *
   * @param {Volume} volume - Volume to delete
   * @returns {Promise<void>}
   * @throws {Error} If the Volume does not exist or cannot be deleted
   *
   * @example
   * const daytona = new Daytona();
   * const volume = await daytona.volume.get("volume-name");
   * await daytona.volume.delete(volume);
   * console.log("Volume deleted successfully");
   */
  async delete(volume: Volume): Promise<void> {
    await this.volumesApi.deleteVolume(volume.id)
  }
}
