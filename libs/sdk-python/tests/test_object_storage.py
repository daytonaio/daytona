# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest


def _make_storage():
    from daytona._sync.object_storage import ObjectStorage

    with patch("daytona._sync.object_storage.S3Store") as mock_store_cls:
        mock_store = MagicMock()
        mock_store_cls.return_value = mock_store
        storage = ObjectStorage("https://s3.example", "key", "secret", "token", bucket_name="bucket")
    return storage, mock_store


class TestObjectStorage:
    def test_upload_missing_path_raises(self):
        storage, _store = _make_storage()

        with pytest.raises(FileNotFoundError, match="Path does not exist"):
            storage.upload("/missing/path", "org-1")

    def test_compute_archive_base_path_trims_root_prefixes(self):
        from daytona._sync.object_storage import ObjectStorage

        assert ObjectStorage.compute_archive_base_path("/workspace/project") == "workspace/project"
        windows_style = ObjectStorage.compute_archive_base_path(r"\\workspace\\project")
        assert windows_style.startswith("workspace")
        assert not windows_style.startswith("\\")

    def test_file_exists_in_s3_true_when_head_succeeds(self):
        storage, store = _make_storage()

        assert storage._file_exists_in_s3("org/hash/context.tar") is True
        store.head.assert_called_once_with("org/hash/context.tar")

    def test_file_exists_in_s3_false_when_head_raises_not_found(self):
        storage, store = _make_storage()
        store.head.side_effect = FileNotFoundError()

        assert storage._file_exists_in_s3("org/hash/context.tar") is False

    def test_compute_hash_for_file_changes_with_contents(self, tmp_path: Path):
        storage, _store = _make_storage()
        file_path = tmp_path / "file.txt"
        file_path.write_text("one", encoding="utf-8")

        first_hash = storage._compute_hash_for_path_md5(str(file_path))
        file_path.write_text("two", encoding="utf-8")
        second_hash = storage._compute_hash_for_path_md5(str(file_path))

        assert first_hash != second_hash

    def test_compute_hash_for_directory_changes_with_new_file(self, tmp_path: Path):
        storage, _store = _make_storage()
        directory = tmp_path / "dir"
        directory.mkdir()
        (directory / "a.txt").write_text("a", encoding="utf-8")

        first_hash = storage._compute_hash_for_path_md5(str(directory))
        (directory / "b.txt").write_text("b", encoding="utf-8")
        second_hash = storage._compute_hash_for_path_md5(str(directory))

        assert first_hash != second_hash

    def test_upload_returns_existing_hash_without_uploading(self, tmp_path: Path):
        storage, _store = _make_storage()
        file_path = tmp_path / "file.txt"
        file_path.write_text("hello", encoding="utf-8")
        storage._compute_hash_for_path_md5 = MagicMock(return_value="hash123")
        storage._file_exists_in_s3 = MagicMock(return_value=True)
        storage._upload_as_tar = MagicMock()

        result = storage.upload(str(file_path), "org-1")

        assert result == "hash123"
        storage._file_exists_in_s3.assert_called_once_with("org-1/hash123/context.tar")
        storage._upload_as_tar.assert_not_called()

    def test_upload_uses_archive_base_path_when_uploading(self, tmp_path: Path):
        storage, _store = _make_storage()
        file_path = tmp_path / "file.txt"
        file_path.write_text("hello", encoding="utf-8")
        storage._compute_hash_for_path_md5 = MagicMock(return_value="hash123")
        storage._file_exists_in_s3 = MagicMock(return_value=False)
        storage._upload_as_tar = MagicMock()

        result = storage.upload(str(file_path), "org-1", archive_base_path="custom/base")

        assert result == "hash123"
        storage._compute_hash_for_path_md5.assert_called_once_with(str(file_path), "custom/base")
        storage._upload_as_tar.assert_called_once_with("org-1/hash123/context.tar", str(file_path), "custom/base")

    def test_upload_as_tar_streams_content_to_store(self, tmp_path: Path):
        storage, store = _make_storage()
        file_path = tmp_path / "file.txt"
        file_path.write_text("hello world", encoding="utf-8")

        captured: dict[str, bytes] = {}

        def consume_stream(key: str, chunks):
            captured[key] = b"".join(chunks)

        store.put.side_effect = consume_stream

        storage._upload_as_tar("org/hash/context.tar", str(file_path), "base/file.txt")

        assert "org/hash/context.tar" in captured
        assert len(captured["org/hash/context.tar"]) > 0
