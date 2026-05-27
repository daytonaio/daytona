# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Tests for daytona.common.errors module."""

from __future__ import annotations

import pytest

from daytona.common.errors import (
    DaytonaBadGatewayError,
    DaytonaConnectionError,
    DaytonaConnectionTimeoutError,
    DaytonaError,
    DaytonaGoneError,
    DaytonaInternalServerError,
    DaytonaNotFoundError,
    DaytonaRateLimitError,
    DaytonaServiceUnavailableError,
    DaytonaTimeoutError,
    DaytonaUnprocessableEntityError,
    create_daytona_error,
    error_class_from_status_code,
)


class TestDaytonaError:
    def test_basic_error(self):
        err = DaytonaError("something went wrong")
        assert str(err) == "something went wrong"
        assert err.status_code is None
        assert err.headers == {}

    def test_with_status_code(self):
        err = DaytonaError("bad request", status_code=400)
        assert err.status_code == 400
        assert str(err) == "bad request"

    def test_with_headers(self):
        headers = {"X-RateLimit-Remaining": "0", "Retry-After": "60"}
        err = DaytonaError("rate limited", status_code=429, headers=headers)
        assert err.status_code == 429
        assert err.headers["X-RateLimit-Remaining"] == "0"
        assert err.headers["Retry-After"] == "60"

    def test_is_exception(self):
        err = DaytonaError("test")
        assert isinstance(err, Exception)

    def test_none_headers_becomes_empty_dict(self):
        err = DaytonaError("msg", headers=None)
        assert err.headers == {}


class TestDaytonaNotFoundError:
    def test_inherits_daytona_error(self):
        err = DaytonaNotFoundError("sandbox not found", status_code=404)
        assert isinstance(err, DaytonaError)
        assert isinstance(err, Exception)
        assert err.status_code == 404

    def test_message(self):
        err = DaytonaNotFoundError("not found")
        assert str(err) == "not found"


class TestDaytonaRateLimitError:
    def test_inherits_daytona_error(self):
        err = DaytonaRateLimitError("rate limit exceeded", status_code=429)
        assert isinstance(err, DaytonaError)
        assert err.status_code == 429

    def test_with_retry_header(self):
        err = DaytonaRateLimitError(
            "rate limit",
            status_code=429,
            headers={"Retry-After": "30"},
        )
        assert err.headers["Retry-After"] == "30"


class TestDaytonaTimeoutError:
    def test_inherits_daytona_error(self):
        err = DaytonaTimeoutError("operation timed out")
        assert isinstance(err, DaytonaError)
        assert str(err) == "operation timed out"

    def test_with_status_code(self):
        err = DaytonaTimeoutError("timeout", status_code=504)
        assert err.status_code == 504


class TestErrorHierarchy:
    def test_catch_all_with_base_class(self):
        errors = [
            DaytonaError("base"),
            DaytonaNotFoundError("not found"),
            DaytonaRateLimitError("rate limit"),
            DaytonaTimeoutError("timeout"),
        ]
        for err in errors:
            with pytest.raises(DaytonaError):
                raise err

    def test_specific_catch(self):
        with pytest.raises(DaytonaNotFoundError):
            raise DaytonaNotFoundError("not found")

        with pytest.raises(DaytonaRateLimitError):
            raise DaytonaRateLimitError("rate limit")

        with pytest.raises(DaytonaTimeoutError):
            raise DaytonaTimeoutError("timeout")


class TestErrorFactories:
    def test_error_class_from_status_code(self):
        assert error_class_from_status_code(404) is DaytonaNotFoundError
        assert error_class_from_status_code(None) is DaytonaError

    def test_create_daytona_error_uses_specific_subclass(self):
        error = create_daytona_error("missing", status_code=404, code="NOT_FOUND")

        assert isinstance(error, DaytonaNotFoundError)
        assert error.code == "NOT_FOUND"


class TestStatusCodeClassification:
    """Every HTTP status code that Daytona services actually emit has a typed
    DaytonaError subclass. The ``(source, code)`` catalog can override these
    when a specific code needs a different class.
    """

    @pytest.mark.parametrize(
        "status_code,expected_cls",
        [
            (408, DaytonaTimeoutError),
            (410, DaytonaGoneError),
            (422, DaytonaUnprocessableEntityError),
            (500, DaytonaInternalServerError),
            (502, DaytonaBadGatewayError),
            (503, DaytonaServiceUnavailableError),
            (504, DaytonaTimeoutError),
        ],
    )
    def test_status_code_maps_to_typed_class(self, status_code, expected_cls):
        assert error_class_from_status_code(status_code) is expected_cls

    def test_unknown_status_code_falls_back_to_base(self):
        assert error_class_from_status_code(418) is DaytonaError
        assert error_class_from_status_code(None) is DaytonaError


class TestTransportErrorMapping:
    """`intercept_errors` translates aiohttp / urllib3 transport drops into
    ``DaytonaConnectionError`` / ``DaytonaTimeoutError`` so users can branch on
    infra failures without parsing strings.
    """

    def test_aiohttp_payload_error_maps_to_connection_error(self):
        import aiohttp

        from daytona._utils.errors import intercept_errors

        @intercept_errors()
        def fn():
            raise aiohttp.ClientPayloadError("payload truncated")

        with pytest.raises(DaytonaConnectionError):
            fn()

    def test_aiohttp_server_disconnected_maps_to_connection_error(self):
        import aiohttp

        from daytona._utils.errors import intercept_errors

        @intercept_errors()
        def fn():
            raise aiohttp.ServerDisconnectedError("server hung up")

        with pytest.raises(DaytonaConnectionError):
            fn()

    def test_urllib3_protocol_error_maps_to_connection_error(self):
        import urllib3.exceptions

        from daytona._utils.errors import intercept_errors

        @intercept_errors()
        def fn():
            raise urllib3.exceptions.ProtocolError("connection broken")

        with pytest.raises(DaytonaConnectionError):
            fn()

    def test_urllib3_read_timeout_maps_to_connection_timeout_error(self):
        import urllib3.exceptions

        from daytona._utils.errors import intercept_errors

        @intercept_errors()
        def fn():
            raise urllib3.exceptions.ReadTimeoutError(None, "/", "read timed out")  # type: ignore[arg-type]

        with pytest.raises(DaytonaConnectionTimeoutError):
            fn()

    def test_connection_timeout_is_a_connection_error(self):
        # Subclass relationship lets callers ``except DaytonaConnectionError`` to
        # catch both can't-connect and connection-timeout cases together.
        err = DaytonaConnectionTimeoutError("read timed out")
        assert isinstance(err, DaytonaConnectionError)


class TestDomainCodeClassification:
    """Domain-specific (source, code) pairs resolve to precise subclasses that
    inherit from the matching HTTP-status class so callers can branch on either.
    """

    def test_daemon_code_resolves_to_precise_subclass(self):
        from daytona import DaytonaGitAuthFailedError
        from daytona.common.errors import CODE_TO_ERROR

        cls = CODE_TO_ERROR[("DAYTONA_DAEMON", "GIT_AUTH_FAILED")]
        assert cls is DaytonaGitAuthFailedError

    def test_proxy_code_resolves_to_precise_subclass(self):
        from daytona import DaytonaRunnerUnreachableError
        from daytona.common.errors import CODE_TO_ERROR

        cls = CODE_TO_ERROR[("DAYTONA_PROXY", "RUNNER_UNREACHABLE")]
        assert cls is DaytonaRunnerUnreachableError

    def test_precise_class_inherits_from_status_class(self):
        from daytona import (
            DaytonaA11yUnavailableError,
            DaytonaAuthenticationError,
            DaytonaCommandAlreadyCompletedError,
            DaytonaConflictError,
            DaytonaFileAccessDeniedError,
            DaytonaFileNotFoundError,
            DaytonaGitAuthFailedError,
            DaytonaGitBranchExistsError,
            DaytonaGitMergeConflictError,
            DaytonaGoneError,
            DaytonaLspServerNotInitializedError,
            DaytonaProcessExecutionTimeoutError,
            DaytonaRunnerUnreachableError,
            DaytonaSandboxNotFoundError,
            DaytonaServiceUnavailableError,
            DaytonaSessionEndedError,
        )
        from daytona.common.errors import (
            DaytonaAuthorizationError,
            DaytonaBadGatewayError,
            DaytonaNotFoundError,
            DaytonaTimeoutError,
            DaytonaValidationError,
        )

        assert issubclass(DaytonaGitAuthFailedError, DaytonaAuthenticationError)
        assert issubclass(DaytonaFileNotFoundError, DaytonaNotFoundError)
        assert issubclass(DaytonaFileAccessDeniedError, DaytonaAuthorizationError)
        assert issubclass(DaytonaGitBranchExistsError, DaytonaConflictError)
        assert issubclass(DaytonaGitMergeConflictError, DaytonaConflictError)
        assert issubclass(DaytonaLspServerNotInitializedError, DaytonaValidationError)
        assert issubclass(DaytonaProcessExecutionTimeoutError, DaytonaTimeoutError)
        assert issubclass(DaytonaSessionEndedError, DaytonaGoneError)
        assert issubclass(DaytonaCommandAlreadyCompletedError, DaytonaGoneError)
        assert issubclass(DaytonaA11yUnavailableError, DaytonaServiceUnavailableError)
        assert issubclass(DaytonaSandboxNotFoundError, DaytonaNotFoundError)
        assert issubclass(DaytonaRunnerUnreachableError, DaytonaBadGatewayError)
