# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Tests for daytona.common.errors module."""

from __future__ import annotations

import pytest

from daytona.common.errors import (
    DaytonaError,
    DaytonaNotFoundError,
    DaytonaRateLimitError,
    DaytonaTimeoutError,
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
