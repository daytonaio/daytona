# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Tests for common types and utilities shared between sync/async packages."""

from __future__ import annotations

import warnings

import pytest

from daytona.common.daytona import (
    CodeLanguage,
    CreateSandboxBaseParams,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from daytona.common.filesystem import FileDownloadRequest, FileDownloadResponse, FileUpload
from daytona.common.git import GitCommitResponse
from daytona.common.image import Image
from daytona.common.lsp_server import LspCompletionPosition, LspLanguageId
from daytona.common.process import (
    CodeRunParams,
    ExecuteResponse,
    ExecutionArtifacts,
    SessionCommandLogsResponse,
    SessionExecuteRequest,
    SessionExecuteResponse,
    demux_log,
    parse_session_command_logs,
    STDOUT_PREFIX,
    STDERR_PREFIX,
)
from daytona.common.pty import PtyResult, PtySize
from daytona.common.sandbox import Resources
from daytona.common.snapshot import CreateSnapshotParams, Snapshot
from daytona.common.volume import Volume, VolumeMount
from daytona.common.charts import (
    Chart,
    ChartType,
    LineChart,
    ScatterChart,
    BarChart,
    PieChart,
    BoxAndWhiskerChart,
    CompositeChart,
    PointData,
    BarData,
    PieData,
    BoxAndWhiskerData,
    parse_chart,
)


# --- DaytonaConfig ---
class TestDaytonaConfig:
    def test_basic_config(self):
        config = DaytonaConfig(api_key="key123", api_url="https://api.test.io")
        assert config.api_key == "key123"
        assert config.api_url == "https://api.test.io"
        assert config.target is None

    def test_config_with_all_fields(self):
        config = DaytonaConfig(
            api_key="key",
            api_url="https://api.test.io",
            target="us",
            jwt_token="jwt",
            organization_id="org-1",
        )
        assert config.jwt_token == "jwt"
        assert config.organization_id == "org-1"
        assert config.target == "us"

    def test_deprecated_server_url(self):
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            config = DaytonaConfig(server_url="https://old.api.io")
            assert config.api_url == "https://old.api.io"
            assert any("deprecated" in str(warning.message).lower() for warning in w)

    def test_api_url_takes_precedence_over_server_url(self):
        with warnings.catch_warnings(record=True):
            warnings.simplefilter("always")
            config = DaytonaConfig(
                api_url="https://new.api.io",
                server_url="https://old.api.io",
            )
            assert config.api_url == "https://new.api.io"


# --- CodeLanguage ---
class TestCodeLanguage:
    def test_enum_values(self):
        assert CodeLanguage.PYTHON.value == "python"
        assert CodeLanguage.TYPESCRIPT.value == "typescript"
        assert CodeLanguage.JAVASCRIPT.value == "javascript"

    def test_str_representation(self):
        assert str(CodeLanguage.PYTHON) == "python"

    def test_equality_with_string(self):
        assert CodeLanguage.PYTHON == "python"
        assert CodeLanguage.TYPESCRIPT == "typescript"

    def test_inequality(self):
        assert CodeLanguage.PYTHON != "typescript"
        assert CodeLanguage.PYTHON != 42


# --- CreateSandboxParams ---
class TestCreateSandboxParams:
    def test_snapshot_params_defaults(self):
        params = CreateSandboxFromSnapshotParams()
        assert params.snapshot is None
        assert params.language is None
        assert params.auto_stop_interval is None

    def test_snapshot_params_with_values(self):
        params = CreateSandboxFromSnapshotParams(
            snapshot="my-snapshot",
            language="python",
            env_vars={"DEBUG": "1"},
            auto_stop_interval=30,
        )
        assert params.snapshot == "my-snapshot"
        assert params.env_vars == {"DEBUG": "1"}
        assert params.auto_stop_interval == 30

    def test_image_params(self):
        params = CreateSandboxFromImageParams(
            image="python:3.12",
            resources=Resources(cpu=2, memory=4),
        )
        assert params.image == "python:3.12"
        assert params.resources.cpu == 2
        assert params.resources.memory == 4

    def test_image_params_with_image_object(self):
        img = Image.base("python:3.12")
        params = CreateSandboxFromImageParams(image=img)
        assert isinstance(params.image, Image)

    def test_ephemeral_sets_auto_delete_to_zero(self):
        params = CreateSandboxFromSnapshotParams(ephemeral=True)
        assert params.auto_delete_interval == 0

    def test_ephemeral_with_auto_delete_warns(self):
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            params = CreateSandboxFromSnapshotParams(
                ephemeral=True, auto_delete_interval=60
            )
            assert params.auto_delete_interval == 0
            assert any("ephemeral" in str(warning.message).lower() for warning in w)

    def test_volume_mounts(self):
        mount = VolumeMount(volume_id="vol-1", mount_path="/data")
        params = CreateSandboxFromSnapshotParams(volumes=[mount])
        assert len(params.volumes) == 1
        assert params.volumes[0].volume_id == "vol-1"


# --- Resources ---
class TestResources:
    def test_default_values(self):
        r = Resources()
        assert r.cpu is None
        assert r.memory is None
        assert r.disk is None
        assert r.gpu is None

    def test_custom_values(self):
        r = Resources(cpu=4, memory=8, disk=30, gpu=1)
        assert r.cpu == 4
        assert r.memory == 8
        assert r.disk == 30
        assert r.gpu == 1


# --- PtySize / PtyResult ---
class TestPtyTypes:
    def test_pty_size(self):
        size = PtySize(rows=24, cols=80)
        assert size.rows == 24
        assert size.cols == 80

    def test_pty_result(self):
        result = PtyResult(exit_code=0, error=None)
        assert result.exit_code == 0
        assert result.error is None

    def test_pty_result_with_error(self):
        result = PtyResult(exit_code=1, error="command failed")
        assert result.exit_code == 1
        assert result.error == "command failed"


# --- FileSystem types ---
class TestFileSystemTypes:
    def test_file_upload(self):
        upload = FileUpload(source=b"hello", destination="/tmp/hello.txt")
        assert upload.source == b"hello"
        assert upload.destination == "/tmp/hello.txt"

    def test_file_upload_from_path(self):
        upload = FileUpload(source="/local/file.txt", destination="/remote/file.txt")
        assert upload.source == "/local/file.txt"

    def test_file_download_request(self):
        req = FileDownloadRequest(source="/remote/file.txt")
        assert req.source == "/remote/file.txt"
        assert req.destination is None

    def test_file_download_request_with_destination(self):
        req = FileDownloadRequest(source="/remote/file.txt", destination="/local/file.txt")
        assert req.destination == "/local/file.txt"

    def test_file_download_response(self):
        resp = FileDownloadResponse(source="/remote/file.txt", result=b"content")
        assert resp.source == "/remote/file.txt"
        assert resp.result == b"content"
        assert resp.error is None

    def test_file_download_response_with_error(self):
        resp = FileDownloadResponse(source="/remote/file.txt", error="not found")
        assert resp.error == "not found"
        assert resp.result is None


# --- Process types ---
class TestProcessTypes:
    def test_code_run_params(self):
        params = CodeRunParams(argv=["--verbose"], env={"DEBUG": "1"})
        assert params.argv == ["--verbose"]
        assert params.env == {"DEBUG": "1"}

    def test_code_run_params_defaults(self):
        params = CodeRunParams()
        assert params.argv is None
        assert params.env is None

    def test_execution_artifacts(self):
        artifacts = ExecutionArtifacts(stdout="hello\n", charts=[])
        assert artifacts.stdout == "hello\n"
        assert artifacts.charts == []

    def test_execute_response(self):
        resp = ExecuteResponse(exit_code=0, result="output")
        assert resp.exit_code == 0
        assert resp.result == "output"

    def test_session_command_logs_response(self):
        resp = SessionCommandLogsResponse(output="combined", stdout="out", stderr="err")
        assert resp.output == "combined"
        assert resp.stdout == "out"
        assert resp.stderr == "err"


class TestDemuxLog:
    def test_empty_data(self):
        stdout, stderr = demux_log(b"")
        assert stdout == b""
        assert stderr == b""

    def test_stdout_only(self):
        data = STDOUT_PREFIX + b"hello world"
        stdout, stderr = demux_log(data)
        assert stdout == b"hello world"
        assert stderr == b""

    def test_stderr_only(self):
        data = STDERR_PREFIX + b"error message"
        stdout, stderr = demux_log(data)
        assert stdout == b""
        assert stderr == b"error message"

    def test_mixed_streams(self):
        data = STDOUT_PREFIX + b"out1" + STDERR_PREFIX + b"err1" + STDOUT_PREFIX + b"out2"
        stdout, stderr = demux_log(data)
        assert stdout == b"out1out2"
        assert stderr == b"err1"

    def test_parse_session_command_logs(self):
        data = STDOUT_PREFIX + b"stdout data" + STDERR_PREFIX + b"stderr data"
        result = parse_session_command_logs(data)
        assert result.stdout == "stdout data"
        assert result.stderr == "stderr data"
        assert result.output is not None


# --- Git types ---
class TestGitTypes:
    def test_git_commit_response(self):
        resp = GitCommitResponse(sha="abc123def456")
        assert resp.sha == "abc123def456"


# --- LSP types ---
class TestLspTypes:
    def test_lsp_language_id(self):
        assert LspLanguageId.PYTHON.value == "python"
        assert LspLanguageId.TYPESCRIPT.value == "typescript"
        assert str(LspLanguageId.PYTHON) == "python"

    def test_lsp_completion_position(self):
        pos = LspCompletionPosition(line=10, character=15)
        assert pos.line == 10
        assert pos.character == 15


# --- VolumeMount ---
class TestVolumeMount:
    def test_volume_mount(self):
        mount = VolumeMount(volume_id="vol-1", mount_path="/data")
        assert mount.volume_id == "vol-1"
        assert mount.mount_path == "/data"

    def test_volume_mount_with_subpath(self):
        mount = VolumeMount(volume_id="vol-1", mount_path="/data", subpath="prefix/")
        assert mount.subpath == "prefix/"


# --- CreateSnapshotParams ---
class TestCreateSnapshotParams:
    def test_basic_params(self):
        params = CreateSnapshotParams(name="my-snapshot", image="python:3.12")
        assert params.name == "my-snapshot"
        assert params.image == "python:3.12"
        assert params.resources is None
        assert params.entrypoint is None

    def test_with_resources(self):
        params = CreateSnapshotParams(
            name="custom",
            image=Image.base("python:3.12"),
            resources=Resources(cpu=2, memory=4),
            entrypoint=["/bin/bash"],
        )
        assert isinstance(params.image, Image)
        assert params.resources.cpu == 2
        assert params.entrypoint == ["/bin/bash"]


# --- Charts ---
class TestCharts:
    def test_chart_type_enum(self):
        assert ChartType.LINE.value == "line"
        assert ChartType.SCATTER.value == "scatter"
        assert ChartType.BAR.value == "bar"
        assert ChartType.PIE.value == "pie"

    def test_basic_chart(self):
        chart = Chart(type=ChartType.UNKNOWN, title="Test Chart")
        assert chart.type == ChartType.UNKNOWN
        assert chart.title == "Test Chart"

    def test_chart_to_dict(self):
        chart = Chart(type=ChartType.LINE, title="My Chart")
        d = chart.to_dict()
        assert d["type"] == ChartType.LINE
        assert d["title"] == "My Chart"

    def test_line_chart(self):
        chart = LineChart(
            title="Line",
            x_label="X",
            y_label="Y",
            elements=[{"label": "series1", "points": [(0, 1), (1, 2)]}],
        )
        assert chart.type == ChartType.LINE
        assert len(chart.elements) == 1
        assert chart.elements[0].label == "series1"

    def test_scatter_chart(self):
        chart = ScatterChart(
            title="Scatter",
            elements=[{"label": "data", "points": [(1, 2), (3, 4)]}],
        )
        assert chart.type == ChartType.SCATTER

    def test_bar_chart(self):
        chart = BarChart(
            title="Bars",
            elements=[{"label": "A", "group": "G1", "value": 10}],
        )
        assert chart.type == ChartType.BAR
        assert chart.elements[0].value == 10

    def test_pie_chart(self):
        chart = PieChart(
            title="Pie",
            elements=[{"label": "slice1", "angle": 90.0, "radius": 1.0}],
        )
        assert chart.type == ChartType.PIE
        assert chart.elements[0].angle == 90.0

    def test_box_and_whisker_chart(self):
        chart = BoxAndWhiskerChart(
            title="Box",
            elements=[{"label": "data", "min": 1.0, "max": 10.0, "median": 5.0}],
        )
        assert chart.type == ChartType.BOX_AND_WHISKER

    def test_parse_chart_line(self):
        chart = parse_chart(type="line", title="Test", elements=[])
        assert isinstance(chart, LineChart)
        assert chart.title == "Test"

    def test_parse_chart_bar(self):
        chart = parse_chart(type="bar", title="Bar Test", elements=[])
        assert isinstance(chart, BarChart)

    def test_parse_chart_unknown_type(self):
        chart = parse_chart(type="unknown", title="Unknown")
        assert isinstance(chart, Chart)

    def test_parse_chart_empty(self):
        assert parse_chart() is None

    def test_point_data(self):
        pd = PointData(label="series", points=[(1, 2), (3, 4)])
        assert pd.label == "series"
        assert len(pd.points) == 2

    def test_bar_data(self):
        bd = BarData(label="A", group="G1", value=42)
        assert bd.label == "A"
        assert bd.value == 42

    def test_pie_data(self):
        pd = PieData(label="slice", angle=180.0)
        assert pd.angle == 180.0

    def test_box_whisker_data(self):
        bwd = BoxAndWhiskerData(
            label="data", min=1.0, first_quartile=2.0, median=5.0,
            third_quartile=8.0, max=10.0, outliers=[0.5, 11.0],
        )
        assert bwd.median == 5.0
        assert len(bwd.outliers) == 2
