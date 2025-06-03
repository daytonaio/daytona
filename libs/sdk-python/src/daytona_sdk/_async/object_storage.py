# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import asyncio
import hashlib
import os
import tarfile
import threading

import aioboto3

from .._utils.docs_ignore import docs_ignore


class ObjectStorage:
    """ObjectStorage class for interacting with object storage services.

    Attributes:
        endpoint_url (str): The endpoint URL for the object storage service.
        aws_access_key_id (str): The access key ID for the object storage service.
        aws_secret_access_key (str): The secret access key for the object storage service.
        aws_session_token (str): The session token for the object storage service. Used for temporary credentials.
        bucket_name (str): The name of the bucket to use.
    """

    def __init__(
        self,
        endpoint_url,
        aws_access_key_id,
        aws_secret_access_key,
        aws_session_token,
        bucket_name="daytona-volume-builds",
    ):
        self.bucket_name = bucket_name
        self.endpoint_url = endpoint_url
        self.s3_client = aioboto3.Session().client(
            service_name="s3",
            endpoint_url=self.endpoint_url,
            aws_access_key_id=aws_access_key_id,
            aws_secret_access_key=aws_secret_access_key,
            aws_session_token=aws_session_token,
        )
        # unasync: delete start
        self._client_ctx = None

    async def open_client(self):
        try:
            self._client_ctx = self.s3_client
            # pylint: disable=unnecessary-dunder-call
            self.s3_client = await self._client_ctx.__aenter__()
        except Exception as e:
            raise Exception(f"Error opening S3 client: {e}") from e

    async def close_client(self):
        try:
            await self._client_ctx.__aexit__(None, None, None)
            self.s3_client = None
            self._client_ctx = None
        except Exception as e:
            raise Exception(f"Error closing S3 client: {e}") from e

    # unasync: delete end

    async def upload(self, path, organization_id, archive_base_path=None) -> str:
        """Uploads a file to the object storage service.

        Args:
            path (str): The path to the file to upload.
            organization_id (str): The organization ID to use.
            archive_base_path (str): The base path to use for the archive.
        """
        if not os.path.exists(path):
            raise FileNotFoundError(f"Path does not exist: {path}")

        # Compute hash for the path
        path_hash = self._compute_hash_for_path_md5(path, archive_base_path)

        # Define the S3 prefix
        prefix = f"{organization_id}/{path_hash}/"
        s3_key = f"{prefix}context.tar"

        # Check if it already exists in S3
        if await self._folder_exists_in_s3(prefix):
            return path_hash

        # Upload to S3
        await self._upload_as_tar(s3_key, path, archive_base_path)

        return path_hash

    @staticmethod
    @docs_ignore
    def compute_archive_base_path(path_str) -> str:
        """Compute the base path for an archive. Returns normalized path without the root
        (drive letter or leading slash).

        Args:
            path_str (str): The path to compute the base path for.

        Returns:
            str: The base path for the given path.
        """
        path_str = os.path.normpath(path_str)
        # Remove drive letter for Windows paths (e.g., C:)
        _, path_without_drive = os.path.splitdrive(path_str)
        # Remove leading separators (both / and \)
        return path_without_drive.lstrip("/").lstrip("\\")

    def _compute_hash_for_path_md5(self, path_str, archive_base_path=None):
        """Computes the MD5 hash for a given path.

        Args:
            path_str (str): The path to compute the hash for.
            archive_base_path (str): The base path to use for the archive.

        Returns:
            str: The MD5 hash for the given path.
        """
        md5_hasher = hashlib.md5()
        abs_path_str = os.path.abspath(path_str)

        if archive_base_path is None:
            archive_base_path = self.compute_archive_base_path(path_str)
        md5_hasher.update(archive_base_path.encode("utf-8"))

        if os.path.isfile(abs_path_str):
            with open(abs_path_str, "rb") as f:
                for chunk in iter(lambda: f.read(8192), b""):
                    md5_hasher.update(chunk)
        else:
            for root, dirs, files in os.walk(abs_path_str):
                if not dirs and not files:
                    rel_dir = os.path.relpath(root, path_str)
                    md5_hasher.update(rel_dir.encode("utf-8"))
                for filename in files:
                    file_path = os.path.join(root, filename)
                    rel_path = os.path.relpath(file_path, abs_path_str)

                    # Incorporate the relative path
                    md5_hasher.update(rel_path.encode("utf-8"))

                    # Incorporate file contents
                    with open(file_path, "rb") as f:
                        for chunk in iter(lambda: f.read(8192), b""):
                            md5_hasher.update(chunk)

        return md5_hasher.hexdigest()

    async def _folder_exists_in_s3(self, prefix):
        """Checks if a folder exists in S3.

        Args:
            prefix (str): The prefix to check.

        Returns:
            bool: True if the folder exists, False otherwise.
        """
        resp = await self.s3_client.list_objects_v2(Bucket=self.bucket_name, Prefix=prefix)
        return "Contents" in resp

    async def _upload_as_tar(self, s3_key, source_path, archive_base_path=None):
        """Uploads a file to the object storage service as a tar.

        Args:
            s3_key (str): The key to upload the file to.
            source_path (str): The path to the file to upload.
            archive_base_path (str): The base path to use for the archive.
        """
        source_path = os.path.normpath(source_path)

        if archive_base_path is None:
            archive_base_path = self.compute_archive_base_path(source_path)

        read_fd, write_fd = os.pipe()
        read_file = os.fdopen(read_fd, "rb")
        write_file = os.fdopen(write_fd, "wb")

        def tar_worker():
            with tarfile.open(fileobj=write_file, mode="w|") as tar:
                tar.add(source_path, arcname=archive_base_path)
            write_file.close()

        thread = threading.Thread(target=tar_worker, daemon=True)
        thread.start()

        await self.s3_client.upload_fileobj(Fileobj=read_file, Bucket=self.bucket_name, Key=s3_key)

        read_file.close()
        await asyncio.to_thread(thread.join)
