# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Guard against the common `multipart` vs `python-multipart` namespace conflict.
# Both packages install into the `multipart` namespace. If a user previously had the
# old `multipart` package installed and upgrades to a Daytona SDK version that depends
# on `python-multipart`, pip won't auto-remove `multipart`, causing import failures.

import importlib.metadata as _meta

try:
    _ = _meta.version("multipart")
except _meta.PackageNotFoundError:
    pass
else:
    raise ImportError(
        "The 'multipart' package conflicts with 'python-multipart' required by the "
        + "Daytona SDK. Both packages use the same 'multipart' namespace.\n"
        + "Please run:\n\n"
        + "  pip uninstall multipart && pip install --force-reinstall python-multipart\n"
    )
