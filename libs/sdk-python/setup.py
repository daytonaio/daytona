# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from setuptools import find_packages, setup

setup(
    name="daytona",
    version="0.1.4",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    install_requires=[
        # Add dependencies here
    ],
    # Include both packages
    package_data={"daytona": ["*"]},
)
