/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DiskDto, DisksApi } from '@daytonaio/api-client'
import { DaytonaNotFoundError } from './errors/DaytonaError'

/**
 * Represents a Daytona Disk which is persistent storage for Sandboxes.
 *
 * @property {string} id - Unique identifier for the Disk
 * @property {string} name - Name of the Disk
 * @property {string} organizationId - Organization ID that owns the Disk
 * @property {number} size - Disk size in GB
 * @property {string} state - Current state of the Disk
 * @property {string} runnerId - Runner ID where the Disk is located
 * @property {string} errorReason - Error reason if Disk is in error state
 * @property {string} createdAt - Date and time when the Disk was created
 * @property {string} updatedAt - Date and time when the Disk was last updated
 */
export type Disk = DiskDto & { __brand: 'Disk' }

/**
 * Service for managing Daytona Disks.
 *
 * This service provides methods to list, get, create, and delete Disks.
 *
 * @class
 */
export class DiskService {
  constructor(private disksApi: DisksApi) {}

  /**
   * Lists all available Disks.
   *
   * @returns {Promise<Disk[]>} List of all Disks accessible to the user
   *
   * @example
   * const daytona = new Daytona();
   * const disks = await daytona.disk.list();
   * console.log(`Found ${disks.length} disks`);
   * disks.forEach(disk => console.log(`${disk.name} (${disk.id}) - ${disk.size}GB`));
   */
  async list(): Promise<Disk[]> {
    const response = await this.disksApi.listDisks()
    return response.data as Disk[]
  }

  /**
   * Gets a Disk by its ID.
   *
   * @param {string} diskId - ID of the Disk to retrieve
   * @returns {Promise<Disk>} The requested Disk
   * @throws {Error} If the Disk does not exist or cannot be accessed
   *
   * @example
   * const daytona = new Daytona();
   * const disk = await daytona.disk.get("disk-id");
   * console.log(`Disk ${disk.name} is in state ${disk.state}`);
   */
  async get(diskId: string): Promise<Disk> {
    const response = await this.disksApi.getDisk(diskId)
    return response.data as Disk
  }

  /**
   * Creates a new Disk with the specified name and size.
   *
   * @param {string} name - Name for the new Disk
   * @param {number} size - Size of the Disk in GB
   * @returns {Promise<Disk>} The newly created Disk
   * @throws {Error} If the Disk cannot be created
   *
   * @example
   * const daytona = new Daytona();
   * const disk = await daytona.disk.create("my-data-disk", 50);
   * console.log(`Created disk ${disk.name} with ID ${disk.id} (${disk.size}GB)`);
   */
  async create(name: string, size: number): Promise<Disk> {
    const response = await this.disksApi.createDisk({ name, size })
    return response.data as Disk
  }

  /**
   * Deletes a Disk.
   *
   * @param {Disk} disk - Disk to delete
   * @returns {Promise<void>}
   * @throws {Error} If the Disk does not exist or cannot be deleted
   *
   * @example
   * const daytona = new Daytona();
   * const disk = await daytona.disk.get("disk-name");
   * await daytona.disk.delete(disk);
   * console.log("Disk deleted successfully");
   */
  async delete(disk: Disk): Promise<void> {
    await this.disksApi.deleteDisk(disk.id)
  }
}
