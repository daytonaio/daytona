# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings

import pytest

from daytona.common.charts import (
    BarChart,
    BoxAndWhiskerChart,
    Chart,
    ChartType,
    CompositeChart,
    LineChart,
    PieChart,
    PointData,
    ScatterChart,
    parse_chart,
)
from daytona.common.computer_use import ScreenshotOptions
from daytona.common.daytona import (
    CodeLanguage,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from daytona.common.errors import DaytonaNotFoundError, DaytonaValidationError
from daytona.common.filesystem import (
    FileDownloadErrorDetails,
    FileDownloadRequest,
    FileDownloadResponse,
    FileUpload,
    create_file_download_error,
    parse_file_download_error_payload,
)
from daytona.common.git import GitCommitResponse
from daytona.common.image import Image
from daytona.common.lsp_server import LspCompletionPosition, LspLanguageId
from daytona.common.process import (
    STDERR_PREFIX,
    STDOUT_PREFIX,
    CodeRunParams,
    ExecuteResponse,
    ExecutionArtifacts,
    SessionCommandLogsResponse,
    SessionExecuteRequest,
    SessionExecuteResponse,
)
from daytona.common.pty import PtyResult, PtySize
from daytona.common.sandbox import Resources
from daytona.common.snapshot import CreateSnapshotParams
from daytona.common.volume import Volume, VolumeMount
from daytona_toolbox_api_client import Chart as GeneratedChart
from daytona_toolbox_api_client import ChartElement as GeneratedChartElement


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
            connection_pool_maxsize=None,
        )
        assert config.jwt_token == "jwt"
        assert config.organization_id == "org-1"
        assert config.target == "us"
        assert config.connection_pool_maxsize is None

    def test_deprecated_server_url_sets_api_url(self):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            config = DaytonaConfig(server_url="https://old.api.io")
        assert config.api_url == "https://old.api.io"
        assert any("deprecated" in str(w.message).lower() for w in caught)

    def test_api_url_takes_precedence_over_server_url(self):
        with warnings.catch_warnings(record=True):
            warnings.simplefilter("always")
            config = DaytonaConfig(api_url="https://new.api.io", server_url="https://old.api.io")
        assert config.api_url == "https://new.api.io"


class TestCodeLanguage:
    @pytest.mark.parametrize(
        ("language", "expected"),
        [
            (CodeLanguage.PYTHON, "python"),
            (CodeLanguage.TYPESCRIPT, "typescript"),
            (CodeLanguage.JAVASCRIPT, "javascript"),
        ],
    )
    def test_enum_values_and_str(self, language, expected):
        assert language.value == expected
        assert str(language) == expected
        assert language == expected

    def test_inequality(self):
        assert CodeLanguage.PYTHON != "typescript"
        assert CodeLanguage.PYTHON != 42


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
        params = CreateSandboxFromImageParams(image="python:3.12", resources=Resources(cpu=2, memory=4))
        assert params.image == "python:3.12"
        assert params.resources is not None
        assert params.resources.cpu == 2
        assert params.resources.memory == 4

    def test_image_params_with_image_object(self):
        params = CreateSandboxFromImageParams(image=Image.base("python:3.12"))
        assert isinstance(params.image, Image)

    def test_ephemeral_sets_auto_delete_to_zero(self):
        params = CreateSandboxFromSnapshotParams(ephemeral=True)
        assert params.auto_delete_interval == 0

    def test_ephemeral_with_auto_delete_warns(self):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            params = CreateSandboxFromSnapshotParams(ephemeral=True, auto_delete_interval=60)
        assert params.auto_delete_interval == 0
        assert any("ephemeral" in str(w.message).lower() for w in caught)

    def test_volume_mounts(self):
        mount = VolumeMount(volume_id="vol-1", mount_path="/data")
        params = CreateSandboxFromSnapshotParams(volumes=[mount])
        volumes = list(params.volumes or [])
        assert len(volumes) == 1
        assert volumes[0].volume_id == "vol-1"


class TestResourcesAndPtyTypes:
    def test_resources_defaults(self):
        assert Resources() == Resources(cpu=None, memory=None, disk=None, gpu=None)

    def test_resources_custom_values(self):
        resources = Resources(cpu=4, memory=8, disk=30, gpu=1)
        assert resources.cpu == 4
        assert resources.memory == 8
        assert resources.disk == 30
        assert resources.gpu == 1

    def test_pty_size(self):
        size = PtySize(rows=24, cols=80)
        assert size.rows == 24
        assert size.cols == 80

    def test_pty_result(self):
        result = PtyResult(exit_code=1, error="failed")
        assert result.exit_code == 1
        assert result.error == "failed"


class TestFilesystemTypes:
    def test_file_upload(self):
        upload = FileUpload(source=b"hello", destination="/tmp/hello.txt")
        assert upload.source == b"hello"
        assert upload.destination == "/tmp/hello.txt"

    def test_file_download_request(self):
        request = FileDownloadRequest(source="/remote/file.txt", destination="/local/file.txt")
        assert request.source == "/remote/file.txt"
        assert request.destination == "/local/file.txt"

    def test_file_download_response(self):
        response = FileDownloadResponse(source="/remote/file.txt", result=b"content")
        assert response.source == "/remote/file.txt"
        assert response.result == b"content"
        assert response.error is None

    def test_create_file_download_error_from_details(self):
        response = FileDownloadResponse(
            source="/remote/file.txt",
            error="not found",
            error_details=FileDownloadErrorDetails(message="missing", status_code=404, error_code="NOT_FOUND"),
        )
        error = create_file_download_error(response)
        assert error.message == "missing"
        assert error.status_code == 404
        assert error.error_code == "NOT_FOUND"

    def test_parse_file_download_error_payload_json(self):
        message, details = parse_file_download_error_payload(
            b'{"message":"missing","statusCode":404,"code":"NOT_FOUND"}',
            "application/json",
        )
        assert message == "missing"
        assert details == FileDownloadErrorDetails(message="missing", status_code=404, error_code="NOT_FOUND")

    def test_parse_file_download_error_payload_plain_text(self):
        message, details = parse_file_download_error_payload(b"plain failure", "text/plain")
        assert message == "plain failure"
        assert details is None

    def test_create_file_download_error_requires_error_message(self):
        with pytest.raises(DaytonaValidationError, match="must not be None"):
            create_file_download_error(FileDownloadResponse(source="/tmp/file"))

    def test_parse_file_download_error_payload_supports_snake_case_keys(self):
        message, details = parse_file_download_error_payload(
            b'{"message":"missing","status_code":404,"error_code":"NOT_FOUND"}',
            "application/json",
        )
        assert message == "missing"
        assert details == FileDownloadErrorDetails(message="missing", status_code=404, error_code="NOT_FOUND")

    def test_create_file_download_error_maps_structured_status_code(self):
        error = create_file_download_error(
            FileDownloadResponse(
                source="/tmp/file",
                error="not found",
                error_details=FileDownloadErrorDetails(message="missing", status_code=404, error_code="NOT_FOUND"),
            )
        )
        assert isinstance(error, DaytonaNotFoundError)


class TestProcessTypes:
    def test_code_run_params(self):
        params = CodeRunParams(argv=["--verbose"], env={"DEBUG": "1"})
        assert params.argv == ["--verbose"]
        assert params.env == {"DEBUG": "1"}

    def test_execution_artifacts(self):
        artifacts = ExecutionArtifacts(stdout="hello\n", charts=[])
        assert artifacts.stdout == "hello\n"
        assert artifacts.charts == []

    def test_execute_response(self):
        response = ExecuteResponse(exit_code=0, result="output")
        assert response.exit_code == 0
        assert response.result == "output"

    def test_session_execute_request_var_async_maps_to_run_async(self):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            request = SessionExecuteRequest(command="echo hi", var_async=True)
        assert request.run_async is True
        assert any("deprecated" in str(w.message).lower() for w in caught)

    def test_session_execute_request_preserves_existing_run_async(self):
        request = SessionExecuteRequest(command="echo hi", run_async=False)
        assert request.run_async is False

    def test_session_command_logs_response(self):
        response = SessionCommandLogsResponse(output="combined", stdout="out", stderr="err")
        assert response.output == "combined"
        assert response.stdout == "out"
        assert response.stderr == "err"

    def test_session_execute_response(self):
        response = SessionExecuteResponse(cmd_id="cmd-1", stdout="out", stderr="err", output="all", exit_code=0)
        assert response.cmd_id == "cmd-1"
        assert response.exit_code == 0

    def test_stream_prefix_constants_are_distinct(self):
        assert STDOUT_PREFIX != STDERR_PREFIX
        assert len(STDOUT_PREFIX) == 3
        assert len(STDERR_PREFIX) == 3


class TestGitAndLspTypes:
    def test_git_commit_response(self):
        response = GitCommitResponse(sha="abc123def456")
        assert response.sha == "abc123def456"

    @pytest.mark.parametrize(
        ("language", "expected"),
        [
            (LspLanguageId.PYTHON, "python"),
            (LspLanguageId.TYPESCRIPT, "typescript"),
            (LspLanguageId.JAVASCRIPT, "javascript"),
        ],
    )
    def test_lsp_language_id(self, language, expected):
        assert language.value == expected
        assert str(language) == expected

    def test_lsp_completion_position(self):
        position = LspCompletionPosition(line=10, character=15)
        assert position.line == 10
        assert position.character == 15

    def test_screenshot_options_validation(self):
        options = ScreenshotOptions(show_cursor=True, fmt="jpeg", quality=90, scale=0.5)
        assert options.quality == 90
        assert options.scale == 0.5


class TestSnapshotAndVolumeTypes:
    def test_volume_mount_with_subpath(self):
        mount = VolumeMount(volume_id="vol-1", mount_path="/data", subpath="prefix/")
        assert mount.subpath == "prefix/"

    def test_volume_from_dto(self):
        volume = Volume.model_validate(
            {
                "id": "vol-1",
                "name": "test-vol",
                "organization_id": "org-1",
                "state": "ready",
                "error_reason": None,
                "created_at": "2025-01-01T00:00:00Z",
                "updated_at": "2025-01-01T00:00:00Z",
                "last_used_at": "2025-01-01T00:00:00Z",
            }
        )
        assert volume.name == "test-vol"

    def test_create_snapshot_params(self):
        params = CreateSnapshotParams(
            name="my-snapshot",
            image=Image.base("python:3.12"),
            resources=Resources(cpu=2, memory=4),
            entrypoint=["/bin/bash"],
        )
        assert params.name == "my-snapshot"
        assert isinstance(params.image, Image)
        assert params.resources is not None
        assert params.resources.cpu == 2
        assert params.entrypoint == ["/bin/bash"]


class TestCharts:
    @pytest.mark.parametrize(
        ("chart_type", "expected_class"),
        [
            (ChartType.LINE.value, LineChart),
            (ChartType.SCATTER.value, ScatterChart),
            (ChartType.BAR.value, BarChart),
            (ChartType.PIE.value, PieChart),
            (ChartType.BOX_AND_WHISKER.value, BoxAndWhiskerChart),
            (ChartType.COMPOSITE_CHART.value, CompositeChart),
            (ChartType.UNKNOWN.value, Chart),
        ],
    )
    def test_parse_chart_returns_expected_subclass(self, chart_type, expected_class):
        chart = GeneratedChart(type=chart_type, title="Test", elements=[])
        parsed = parse_chart(chart)
        assert isinstance(parsed, expected_class)
        assert parsed.title == "Test"

    def test_chart_model_validate_resolves_subclass(self):
        chart = Chart.model_validate({"type": ChartType.LINE.value, "title": "Line", "elements": []})
        assert isinstance(chart, LineChart)

    def test_chart_model_dump(self):
        chart = Chart.model_validate({"type": ChartType.UNKNOWN.value, "title": "Chart", "elements": []})
        dumped = chart.model_dump()
        assert dumped["type"] == ChartType.UNKNOWN.value
        assert dumped["title"] == "Chart"

    def test_line_chart_with_elements(self):
        chart = LineChart.model_validate(
            {
                "type": ChartType.LINE.value,
                "title": "Line",
                "elements": [GeneratedChartElement(label="series1", points=[[0, 1], [1, 2]])],
            }
        )
        assert chart.elements[0].label == "series1"

    def test_point_data(self):
        point = PointData(label="series", points=[[1, 2], [3, 4]])
        assert point.label == "series"
        assert point.points is not None
        assert len(point.points) == 2
