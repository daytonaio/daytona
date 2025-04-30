/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  S3Client,
  ListObjectsV2Command,
  DeleteObjectsCommand,
  ListObjectVersionsCommand,
  DeleteBucketCommand,
} from '@aws-sdk/client-s3'

export async function deleteS3Bucket(s3: S3Client, bucket: string): Promise<void> {
  // First delete all object versions & delete markers (if any exist)
  let keyMarker: string | undefined
  let versionIdMarker: string | undefined
  do {
    const versions = await s3.send(
      new ListObjectVersionsCommand({
        Bucket: bucket,
        KeyMarker: keyMarker,
        VersionIdMarker: versionIdMarker,
      }),
    )
    const items = [
      ...(versions.Versions || []).map((v) => ({ Key: v.Key, VersionId: v.VersionId })),
      ...(versions.DeleteMarkers || []).map((d) => ({ Key: d.Key, VersionId: d.VersionId })),
    ]
    if (items.length) {
      await s3.send(
        new DeleteObjectsCommand({
          Bucket: bucket,
          Delete: { Objects: items, Quiet: true },
        }),
      )
    }
    keyMarker = versions.NextKeyMarker
    versionIdMarker = versions.NextVersionIdMarker
  } while (keyMarker || versionIdMarker)

  // Then delete any remaining live objects (for unversioned buckets)
  let continuationToken: string | undefined
  do {
    const list = await s3.send(
      new ListObjectsV2Command({
        Bucket: bucket,
        ContinuationToken: continuationToken,
      }),
    )
    if (list.Contents && list.Contents.length) {
      await s3.send(
        new DeleteObjectsCommand({
          Bucket: bucket,
          Delete: {
            Objects: list.Contents.map((o) => ({ Key: o.Key })),
            Quiet: true,
          },
        }),
      )
    }
    continuationToken = list.NextContinuationToken
  } while (continuationToken)

  // Finally delete the (now-empty) bucket
  await s3.send(new DeleteBucketCommand({ Bucket: bucket }))
}
