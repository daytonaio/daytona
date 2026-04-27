# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from unittest.mock import patch

import pytest

from daytona._utils.env import DaytonaEnvReader


class TestDaytonaEnvReader:
    def test_get_rejects_non_daytona_variable_names(self):
        reader = DaytonaEnvReader()

        with pytest.raises(ValueError, match="must start with 'DAYTONA_'"):
            reader.get("OTHER_VAR")

    def test_runtime_env_takes_precedence(self, monkeypatch):
        monkeypatch.setenv("DAYTONA_API_KEY", "runtime")

        with patch.object(
            DaytonaEnvReader, "_load", side_effect=[{"DAYTONA_API_KEY": "local"}, {"DAYTONA_API_KEY": "env"}]
        ):
            reader = DaytonaEnvReader()

        assert reader.get("DAYTONA_API_KEY") == "runtime"

    def test_env_local_takes_precedence_over_env_file(self, monkeypatch):
        monkeypatch.delenv("DAYTONA_API_KEY", raising=False)

        with patch.object(
            DaytonaEnvReader, "_load", side_effect=[{"DAYTONA_API_KEY": "local"}, {"DAYTONA_API_KEY": "env"}]
        ):
            reader = DaytonaEnvReader()

        assert reader.get("DAYTONA_API_KEY") == "local"

    def test_get_returns_none_for_missing_variable(self, monkeypatch):
        monkeypatch.delenv("DAYTONA_API_KEY", raising=False)

        with patch.object(DaytonaEnvReader, "_load", side_effect=[{}, {}]):
            reader = DaytonaEnvReader()

        assert reader.get("DAYTONA_API_KEY") is None

    def test_load_filters_non_daytona_and_none_values(self):
        with patch(
            "daytona._utils.env.dotenv_values",
            return_value={"DAYTONA_API_KEY": "key", "OTHER": "nope", "DAYTONA_TARGET": None},
        ):
            assert DaytonaEnvReader._load(".env") == {"DAYTONA_API_KEY": "key"}
