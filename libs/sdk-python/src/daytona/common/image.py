# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import glob
import os
import re
import shlex
import sys
from pathlib import Path, PurePosixPath
from typing import List, Literal, Optional, Sequence, Union, get_args

import toml
from pydantic import BaseModel, PrivateAttr

from .._sync.object_storage import ObjectStorage
from .errors import DaytonaError

SupportedPythonSeries = Literal["3.9", "3.10", "3.11", "3.12", "3.13"]
SUPPORTED_PYTHON_SERIES = list(get_args(SupportedPythonSeries))
LATEST_PYTHON_MICRO_VERSIONS = ["3.9.22", "3.10.17", "3.11.12", "3.12.10", "3.13.3"]


class Context(BaseModel):
    """Context for an image.

    Attributes:
        source_path (str): The path to the source file or directory.
        archive_path (Optional[str]): The path inside the archive file in object storage.
    """

    source_path: str
    archive_path: Optional[str] = None


class Image(BaseModel):
    """Represents an image definition for a Daytona sandbox.
    Do not construct this class directly. Instead use one of its static factory methods,
    such as `Image.base()`, `Image.debian_slim()`, or `Image.from_dockerfile()`.
    """

    _dockerfile: Optional[str] = PrivateAttr(default=None)
    _context_list: List[Context] = PrivateAttr(default_factory=list)

    def dockerfile(self) -> str:
        """Returns a generated Dockerfile for the image."""
        return self._dockerfile

    def pip_install(
        self,
        *packages: Union[str, list[str]],
        find_links: Optional[list[str]] = None,
        index_url: Optional[str] = None,
        extra_index_urls: Optional[list[str]] = None,
        pre: bool = False,
        extra_options: str = "",
    ) -> "Image":
        """Adds commands to install packages using pip.

        Args:
            *packages: The packages to install.
            find_links: Optional[list[str]]: The find-links to use.
            index_url: Optional[str]: The index URL to use.
            extra_index_urls: Optional[list[str]]: The extra index URLs to use.
            pre: bool = False: Whether to install pre-release packages.
            extra_options: str = "": Additional options to pass to pip. Given string is passed
            directly to the pip install command.

        Returns:
            Image: The image with the pip install commands added.

        Example:
            ```python
            image = Image.debian_slim("3.12").pip_install("requests", "pandas")
            ```
        """
        pkgs = self.__flatten_str_args("pip_install", "packages", packages)
        if not pkgs:
            return self
        extra_args = self.__format_pip_install_args(find_links, index_url, extra_index_urls, pre, extra_options)
        self._dockerfile += f"RUN python -m pip install {shlex.join(sorted(pkgs))}{extra_args}\n"

        return self

    def pip_install_from_requirements(
        self,
        requirements_txt: str,  # Path to a requirements.txt file.
        find_links: Optional[list[str]] = None,
        index_url: Optional[str] = None,
        extra_index_urls: Optional[list[str]] = None,
        pre: bool = False,
        extra_options: str = "",
    ) -> "Image":
        """Installs dependencies from a requirements.txt file.

        Args:
            requirements_txt: str: The path to the requirements.txt file.
            find_links: Optional[list[str]]: The find-links to use.
            index_url: Optional[str]: The index URL to use.
            extra_index_urls: Optional[list[str]]: The extra index URLs to use.
            pre: bool = False: Whether to install pre-release packages.
            extra_options: str = "": Additional options to pass to pip.

        Returns:
            Image: The image with the pip install commands added.

        Example:
            ```python
            image = Image.debian_slim("3.12").pip_install_from_requirements("requirements.txt")
            ```
        """
        requirements_txt = os.path.expanduser(requirements_txt)
        if not Path(requirements_txt).exists():
            raise DaytonaError(f"Requirements file {requirements_txt} does not exist")

        extra_args = self.__format_pip_install_args(find_links, index_url, extra_index_urls, pre, extra_options)

        archive_path = ObjectStorage.compute_archive_base_path(requirements_txt)
        self._context_list.append(Context(source_path=requirements_txt, archive_path=archive_path))
        self._dockerfile += f"COPY {archive_path} /.requirements.txt\n"
        self._dockerfile += f"RUN python -m pip install -r /.requirements.txt{extra_args}\n"

        return self

    def pip_install_from_pyproject(
        self,
        pyproject_toml: str,
        optional_dependencies: list[str],
        find_links: Optional[str] = None,
        index_url: Optional[str] = None,
        extra_index_url: Optional[str] = None,
        pre: bool = False,
        extra_options: str = "",
    ) -> "Image":
        """Installs dependencies from a pyproject.toml file.

        Args:
            pyproject_toml: str: The path to the pyproject.toml file.
            optional_dependencies: list[str] = []: The optional dependencies to install from the pyproject.toml file.
            find_links: Optional[str] = None: The find-links to use.
            index_url: Optional[str] = None: The index URL to use.
            extra_index_url: Optional[str] = None: The extra index URL to use.
            pre: bool = False: Whether to install pre-release packages.
            extra_options: str = "": Additional options to pass to pip. Given string is passed
            directly to the pip install command.

        Returns:
            Image: The image with the pip install commands added.

        Example:
            ```python
            image = Image.debian_slim("3.12") \
                .pip_install_from_pyproject("pyproject.toml", optional_dependencies=["dev"])
            ```
        """
        toml_data = toml.load(os.path.expanduser(pyproject_toml))
        dependencies = []

        if "project" not in toml_data or "dependencies" not in toml_data["project"]:
            msg = (
                "No [project.dependencies] section in pyproject.toml file. "
                "See https://packaging.python.org/en/latest/guides/writing-pyproject-toml "
                "for further file format guidelines."
            )
            raise DaytonaError(msg)

        dependencies.extend(toml_data["project"]["dependencies"])
        if optional_dependencies:
            optionals = toml_data["project"]["optional-dependencies"]
            for dep_group_name in optional_dependencies:
                if dep_group_name in optionals:
                    dependencies.extend(optionals[dep_group_name])

        return self.pip_install(
            *dependencies,
            find_links=find_links,
            index_url=index_url,
            extra_index_urls=extra_index_url,
            pre=pre,
            extra_options=extra_options,
        )

    def add_local_file(self, local_path: Union[str, Path], remote_path: str) -> "Image":
        """Adds a local file to the image.

        Args:
            local_path: Union[str, Path]: The path to the local file.
            remote_path: str: The path to the file in the image.

        Returns:
            Image: The image with the local file added.

        Example:
            ```python
            image = Image.debian_slim("3.12").add_local_file("package.json", "/home/daytona/package.json")
            ```
        """
        if remote_path.endswith("/"):
            remote_path = remote_path + Path(local_path).name

        local_path = os.path.expanduser(local_path)
        archive_path = ObjectStorage.compute_archive_base_path(local_path)
        self._context_list.append(Context(source_path=local_path, archive_path=archive_path))
        self._dockerfile += f"COPY {archive_path} {remote_path}\n"

        return self

    def add_local_dir(self, local_path: Union[str, Path], remote_path: str) -> "Image":
        """Adds a local directory to the image.

        Args:
            local_path: Union[str, Path]: The path to the local directory.
            remote_path: str: The path to the directory in the image.

        Returns:
            Image: The image with the local directory added.

        Example:
            ```python
            image = Image.debian_slim("3.12").add_local_dir("src", "/home/daytona/src")
            ```
        """
        local_path = os.path.expanduser(local_path)
        archive_path = ObjectStorage.compute_archive_base_path(local_path)
        self._context_list.append(Context(source_path=local_path, archive_path=archive_path))
        self._dockerfile += f"COPY {archive_path} {remote_path}\n"

        return self

    def run_commands(self, *commands: Union[str, list[str]]) -> "Image":
        """Runs commands in the image.

        Args:
            *commands: The commands to run.

        Returns:
            Image: The image with the commands added.

        Example:
            ```python
            image = Image.debian_slim("3.12").run_commands(
                'echo "Hello, world!"',
                ['bash', '-c', 'echo Hello, world, again!']
            )
            ```
        """
        for command in commands:
            if isinstance(command, list):
                escaped = []
                for c in command:
                    c_escaped = c.replace('"', '\\\\\\"').replace("'", "\\'")
                    escaped.append(f'"{c_escaped}"')
                self._dockerfile += f"RUN {' '.join(escaped)}\n"
            else:
                self._dockerfile += f"RUN {command}\n"

        return self

    def env(self, env_vars: dict[str, str]) -> "Image":
        """Sets environment variables in the image.

        Args:
            env_vars: dict[str, str]: The environment variables to set.

        Returns:
            Image: The image with the environment variables added.

        Example:
            ```python
            image = Image.debian_slim("3.12").env({"PROJECT_ROOT": "/home/daytona"})
            ```
        """
        non_str_keys = [key for key, val in env_vars.items() if not isinstance(val, str)]
        if non_str_keys:
            raise DaytonaError(f"Image ENV variables must be strings. Invalid keys: {non_str_keys}")

        for key, val in env_vars.items():
            self._dockerfile += f"ENV {key}={shlex.quote(val)}\n"

        return self

    def workdir(self, path: Union[str, Path]) -> "Image":
        """Sets the working directory in the image.

        Args:
            path: Union[str, Path]: The path to the working directory.

        Returns:
            Image: The image with the working directory added.

        Example:
            ```python
            image = Image.debian_slim("3.12").workdir("/home/daytona")
            ```
        """
        self._dockerfile += f"WORKDIR {shlex.quote(str(path))}\n"
        return self

    def entrypoint(self, entrypoint_commands: list[str]) -> "Image":
        """Sets the entrypoint for the image.

        Args:
            entrypoint_commands: list[str]: The commands to set as the entrypoint.

        Returns:
            Image: The image with the entrypoint added.

        Example:
            ```python
            image = Image.debian_slim("3.12").entrypoint(["/bin/bash"])
            ```
        """
        if not isinstance(entrypoint_commands, list) or not all(isinstance(x, str) for x in entrypoint_commands):
            raise DaytonaError("entrypoint_commands must be a list of strings.")

        args_str = self.__flatten_str_args("entrypoint", "entrypoint_commands", entrypoint_commands)
        args_str = '"' + '", "'.join(args_str) + '"' if args_str else ""
        self._dockerfile += f"ENTRYPOINT [{args_str}]\n"

        return self

    def cmd(self, cmd: list[str]) -> "Image":
        """Sets the default command for the image.

        Args:
            cmd: list[str]: The commands to set as the default command.

        Returns:
            Image: The image with the default command added.

        Example:
            ```python
            image = Image.debian_slim("3.12").cmd(["/bin/bash"])
            ```
        """
        if not isinstance(cmd, list) or not all(isinstance(x, str) for x in cmd):
            raise DaytonaError("Image CMD must be a list of strings.")
        cmd_str = self.__flatten_str_args("cmd", "cmd", cmd)
        cmd_str = '"' + '", "'.join(cmd_str) + '"' if cmd_str else ""
        self._dockerfile += f"CMD [{cmd_str}]\n"
        return self

    def dockerfile_commands(
        self,
        dockerfile_commands: list[str],
        context_dir: Optional[Union[Path, str]] = None,
    ) -> "Image":
        """Adds arbitrary Dockerfile-like commands to the image.

        Args:
            *dockerfile_commands: The commands to add to the Dockerfile.
            context_dir: Optional[Union[Path, str]]: The path to the context directory.

        Returns:
            Image: The image with the Dockerfile commands added.

        Example:
            ```python
            image = Image.debian_slim("3.12").dockerfile_commands(["RUN echo 'Hello, world!'"])
            ```
        """
        if context_dir:
            context_dir = os.path.expanduser(context_dir)
            if not os.path.isdir(context_dir):
                raise DaytonaError(f"Context directory {context_dir} does not exist")

        for context_path, original_path in Image.__extract_copy_sources(
            "\n".join(dockerfile_commands), context_dir or ""
        ):
            archive_base_path = context_path
            if context_dir and not original_path.startswith(context_dir):
                archive_base_path = context_path.removeprefix(context_dir)
            self._context_list.append(Context(source_path=context_path, archive_path=archive_base_path))

        self._dockerfile += "\n".join(dockerfile_commands) + "\n"

        return self

    @staticmethod
    def from_dockerfile(path: Union[str, Path]) -> "Image":
        """Creates an Image from an existing Dockerfile.

        Args:
            path: Union[str, Path]: The path to the Dockerfile.

        Returns:
            Image: The image with the Dockerfile added.

        Example:
            ```python
            image = Image.from_dockerfile("Dockerfile")
            ```
        """
        path = Path(os.path.expanduser(path))
        dockerfile = path.read_text()
        img = Image()
        img._dockerfile = dockerfile  # pylint: disable=protected-access

        # remove dockerfile filename from path
        path_prefix = str(path).removesuffix(path.name)

        for context_path, original_path in Image.__extract_copy_sources(dockerfile, path_prefix):
            archive_base_path = context_path
            if not original_path.startswith(path_prefix):
                archive_base_path = context_path.removeprefix(path_prefix)
            # pylint: disable=protected-access
            img._context_list.append(Context(source_path=context_path, archive_path=archive_base_path))

        return img

    @staticmethod
    def base(image: str) -> "Image":
        """Creates an Image from an existing base image.

        Args:
            image: str: The base image to use.

        Returns:
            Image: The image with the base image added.

        Example:
            ```python
            image = Image.base("python:3.12-slim-bookworm")
            ```
        """
        img = Image()
        img._dockerfile = f"FROM {image}\n"  # pylint: disable=protected-access
        return img

    @staticmethod
    def debian_slim(python_version: Optional[SupportedPythonSeries] = None) -> "Image":
        """Creates a Debian slim image based on the official Python Docker image.

        Args:
            python_version: Optional[SupportedPythonSeries]: The Python version to use.

        Returns:
            Image: The image with the Debian slim image added.

        Example:
            ```python
            image = Image.debian_slim("3.12")
            ```
        """
        python_version = Image.__process_python_version(python_version)
        img = Image()
        commands = [
            f"FROM python:{python_version}-slim-bookworm",
            "RUN apt-get update",
            "RUN apt-get install -y gcc gfortran build-essential",
            "RUN pip install --upgrade pip",
            # Set debian front-end to non-interactive to avoid users getting stuck with input prompts.
            "RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections",
        ]
        img._dockerfile = "\n".join(commands) + "\n"  # pylint: disable=protected-access
        return img

    @staticmethod
    def __extract_copy_sources(dockerfile_content, path_prefix="") -> list[tuple[str, str]]:
        """Extracts source files from COPY commands in a Dockerfile.

        Args:
            dockerfile_content: str: The content of the Dockerfile.
            path_prefix: str: The path prefix to use for the sources.

        Returns:
            list[tuple[str, str]]: The list of the actual file path and its corresponding COPY-command source path.
        """
        sources = []
        # Split the Dockerfile into lines
        lines = dockerfile_content.split("\n")

        for line in lines:
            # Skip empty lines and comments
            if not line.strip() or line.strip().startswith("#"):
                continue

            # Check if the line contains a COPY command (at the beginning of the line)
            if re.match(r"^\s*COPY\s", line):
                # Extract the sources from the COPY command
                command_parts = Image.__parse_copy_command(line)

                if command_parts:
                    # Get source paths from the parsed command parts
                    for source in command_parts["sources"]:
                        # Handle absolute and relative paths differently
                        if PurePosixPath(source).is_absolute():
                            # Absolute path - use as is
                            full_path_pattern = source
                        else:
                            # Relative path - add prefix
                            full_path_pattern = os.path.join(path_prefix, source)

                        # Handle glob patterns
                        matching_files = glob.glob(full_path_pattern)

                        if matching_files:
                            sources.extend((matching_file, source) for matching_file in matching_files)
                        else:
                            # If no files match, include the pattern anyway
                            sources.append((full_path_pattern, source))

        return sources

    @staticmethod
    def __parse_copy_command(line):
        """Parses a COPY command to extract sources and destination.

        Args:
            line: str: The line to parse.

        Returns:
            dict[str, str]: A dictionary containing the sources and destination.
        """
        # Remove initial "COPY" and strip whitespace
        parts = line.strip()[4:].strip()

        # Handle JSON array format: COPY ["src1", "src2", "dest"]
        if parts.startswith("["):
            try:
                # Parse the JSON-like array format
                elements = shlex.split(parts.replace("[", "").replace("]", ""))
                if len(elements) < 2:
                    return None

                return {"sources": elements[:-1], "dest": elements[-1]}
            except:
                return None

        # Handle regular format with possible flags
        parts = shlex.split(parts)

        # Extract flags like --chown, --chmod, --from
        sources_start_idx = 0
        for i, part in enumerate(parts):
            if part.startswith("--"):
                # Skip the flag and its value if it has one
                if "=" not in part and i + 1 < len(parts) and not parts[i + 1].startswith("--"):
                    sources_start_idx = i + 2
                else:
                    sources_start_idx = i + 1
            else:
                break

        # After skipping flags, we need at least one source and one destination
        if len(parts) - sources_start_idx < 2:
            return None

        return {"sources": parts[sources_start_idx:-1], "dest": parts[-1]}

    @staticmethod
    def __flatten_str_args(function_name: str, arg_name: str, args: Sequence[Union[str, list[str]]]) -> list[str]:
        """Flattens a list of strings and lists of strings into a single list of strings.

        Args:
            function_name: str: The name of the function that is being called.
            arg_name: str: The name of the argument that is being passed.
            args: Sequence[Union[str, list[str]]]: The list of arguments to flatten.

        Returns:
            list[str]: A list of strings.
        """

        def is_str_list(x):
            return isinstance(x, list) and all(isinstance(y, str) for y in x)

        ret: list[str] = []
        for x in args:
            if isinstance(x, str):
                ret.append(x)
            elif is_str_list(x):
                ret.extend(x)
            else:
                raise DaytonaError(f"{function_name}: {arg_name} must only contain strings")
        return ret

    @staticmethod
    def __format_pip_install_args(
        find_links: Optional[list[str]] = None,
        index_url: Optional[str] = None,
        extra_index_urls: Optional[list[str]] = None,
        pre: bool = False,
        extra_options: str = "",
    ) -> str:
        """Formats the arguments in a single string.

        Args:
            find_links: Optional[list[str]]: The find-links to use.
            index_url: Optional[str]: The index URL to use.
            extra_index_urls: Optional[list[str]]: The extra index URLs to use.
            pre: bool = False: Whether to install pre-release packages.
            extra_options: str = "": Additional options to pass to pip.

        Returns:
            str: The formatted arguments.
        """
        extra_args = ""
        if find_links:
            for find_link in find_links:
                extra_args += f" --find-links {shlex.quote(find_link)}"
        if index_url:
            extra_args += f" --index-url {shlex.quote(index_url)}"
        if extra_index_urls:
            for extra_index_url in extra_index_urls:
                extra_args += f" --extra-index-url {shlex.quote(extra_index_url)}"
        if pre:
            extra_args += " --pre"
        if extra_options:
            extra_args += f" {extra_options.strip()}"

        return extra_args

    @staticmethod
    def __process_python_version(python_version: Optional[str], allow_micro_granularity: bool = True) -> str:
        """Processes the Python version.

        Args:
            python_version: Optional[str]: The Python version to process.
            allow_micro_granularity: bool = True: Whether to allow micro-level granularity.

        Returns:
            str: The processed Python version.

        Raises:
            DaytonaError: If the Python version is invalid.
        """
        if python_version is None:
            # If Python version is unspecified, match the local version, up to the minor component
            python_version = series_version = f"{sys.version_info.major}.{sys.version_info.minor}"
        elif not re.match(r"^3(?:\.\d{1,2}){1,2}(rc\d*)?$", python_version):
            raise DaytonaError(f"Invalid Python version: {python_version!r}")
        else:
            components = python_version.split(".")
            if len(components) == 3 and not allow_micro_granularity:
                raise DaytonaError(
                    "Python version must be specified as 'major.minor' for this interface;"
                    f" micro-level specification ({python_version!r}) is not valid."
                )
            series_version = f"{components[0]}.{components[1]}"

        if series_version not in SUPPORTED_PYTHON_SERIES:
            raise DaytonaError(
                f"Unsupported Python version: {python_version!r}."
                f" Daytona supports the following series: {SUPPORTED_PYTHON_SERIES!r}."
            )

        # If the python version is specified as a micro version, return it as is
        components = python_version.split(".")
        if len(components) > 2:
            return python_version

        # If the python version is specified as a series, return the latest micro version
        series_to_micro_version = dict(tuple(v.rsplit(".", 1)) for v in LATEST_PYTHON_MICRO_VERSIONS)
        python_series_requested = f"{components[0]}.{components[1]}"
        micro_version = series_to_micro_version[python_series_requested]
        return f"{python_series_requested}.{micro_version}"
