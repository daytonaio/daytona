from __future__ import annotations

import os
import tempfile

import pytest

from daytona.common.errors import DaytonaError
from daytona.common.image import Image, SUPPORTED_PYTHON_SERIES


class TestImageBase:
    def test_base_image(self):
        img = Image.base("python:3.12-slim")
        assert img.dockerfile() == "FROM python:3.12-slim\n"

    def test_base_with_tag(self):
        img = Image.base("ubuntu:22.04")
        assert "FROM ubuntu:22.04" in img.dockerfile()


class TestImageDebianSlim:
    def test_default_version(self):
        img = Image.debian_slim()
        df = img.dockerfile()
        assert "FROM python:" in df
        assert "slim-bookworm" in df
        assert "apt-get update" in df
        assert "gcc" in df
        assert "pip install --upgrade pip" in df

    def test_specific_version(self):
        img = Image.debian_slim("3.12")
        df = img.dockerfile()
        assert "FROM python:3.12" in df
        assert "slim-bookworm" in df

    def test_unsupported_version(self):
        with pytest.raises(DaytonaError, match="Invalid Python version"):
            Image.debian_slim("2.7")

    def test_invalid_version_format(self):
        with pytest.raises(DaytonaError, match="Invalid Python version"):
            Image.debian_slim("abc")

    def test_all_supported_versions(self):
        for version in SUPPORTED_PYTHON_SERIES:
            img = Image.debian_slim(version)
            assert f"FROM python:{version}" in img.dockerfile()


class TestImagePipInstall:
    def test_single_package(self):
        img = Image.base("python:3.12").pip_install("requests")
        assert "RUN python -m pip install requests" in img.dockerfile()

    def test_multiple_packages(self):
        img = Image.base("python:3.12").pip_install("requests", "pandas")
        df = img.dockerfile()
        assert "pandas" in df
        assert "requests" in df

    def test_list_of_packages(self):
        img = Image.base("python:3.12").pip_install(["numpy", "scipy"])
        df = img.dockerfile()
        assert "numpy" in df
        assert "scipy" in df

    def test_empty_packages(self):
        img = Image.base("python:3.12")
        original_df = img.dockerfile()
        img.pip_install()
        assert img.dockerfile() == original_df

    def test_with_index_url(self):
        img = Image.base("python:3.12").pip_install(
            "private-pkg", index_url="https://my-pypi.example.com/simple"
        )
        assert "--index-url" in img.dockerfile()

    def test_with_extra_index_urls(self):
        img = Image.base("python:3.12").pip_install(
            "pkg", extra_index_urls=["https://extra1.example.com/simple"]
        )
        assert "--extra-index-url" in img.dockerfile()

    def test_with_find_links(self):
        img = Image.base("python:3.12").pip_install(
            "pkg", find_links=["https://download.example.com/wheels"]
        )
        assert "--find-links" in img.dockerfile()

    def test_with_pre(self):
        img = Image.base("python:3.12").pip_install("pkg", pre=True)
        assert "--pre" in img.dockerfile()

    def test_with_extra_options(self):
        img = Image.base("python:3.12").pip_install("pkg", extra_options="--no-deps")
        assert "--no-deps" in img.dockerfile()


class TestImagePipInstallFromRequirements:
    def test_from_requirements(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write("requests==2.31.0\npandas>=1.5.0\n")
            f.flush()
            try:
                img = Image.base("python:3.12").pip_install_from_requirements(f.name)
                df = img.dockerfile()
                assert "COPY" in df
                assert "pip install -r" in df
            finally:
                os.unlink(f.name)

    def test_nonexistent_requirements(self):
        with pytest.raises(DaytonaError, match="does not exist"):
            Image.base("python:3.12").pip_install_from_requirements("/nonexistent/requirements.txt")


class TestImageRunCommands:
    def test_single_command(self):
        img = Image.base("ubuntu:22.04").run_commands("apt-get update")
        assert "RUN apt-get update" in img.dockerfile()

    def test_multiple_commands(self):
        img = Image.base("ubuntu:22.04").run_commands(
            "apt-get update",
            "apt-get install -y curl",
        )
        df = img.dockerfile()
        assert "RUN apt-get update" in df
        assert "RUN apt-get install -y curl" in df

    def test_list_command(self):
        img = Image.base("ubuntu:22.04").run_commands(["bash", "-c", "echo hello"])
        df = img.dockerfile()
        assert "RUN" in df
        assert "bash" in df


class TestImageEnv:
    def test_set_env_vars(self):
        img = Image.base("python:3.12").env({"MY_VAR": "my_value", "DEBUG": "1"})
        df = img.dockerfile()
        assert "ENV MY_VAR=" in df
        assert "ENV DEBUG=" in df


class TestImageWorkdir:
    def test_set_workdir(self):
        img = Image.base("python:3.12").workdir("/home/daytona")
        assert "WORKDIR" in img.dockerfile()
        assert "/home/daytona" in img.dockerfile()


class TestImageEntrypoint:
    def test_set_entrypoint(self):
        img = Image.base("python:3.12").entrypoint(["/bin/bash"])
        assert 'ENTRYPOINT ["/bin/bash"]' in img.dockerfile()

    def test_multiple_args(self):
        img = Image.base("python:3.12").entrypoint(["python", "-u", "app.py"])
        df = img.dockerfile()
        assert "ENTRYPOINT" in df
        assert "python" in df
        assert "app.py" in df


class TestImageCmd:
    def test_set_cmd(self):
        img = Image.base("python:3.12").cmd(["/bin/bash"])
        assert 'CMD ["/bin/bash"]' in img.dockerfile()

    def test_multiple_args(self):
        img = Image.base("python:3.12").cmd(["python", "main.py"])
        df = img.dockerfile()
        assert "CMD" in df
        assert "python" in df


class TestImageFromDockerfile:
    def test_from_dockerfile(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix="Dockerfile", delete=False) as f:
            f.write("FROM python:3.12\nRUN pip install flask\n")
            f.flush()
            try:
                img = Image.from_dockerfile(f.name)
                df = img.dockerfile()
                assert "FROM python:3.12" in df
                assert "pip install flask" in df
            finally:
                os.unlink(f.name)


class TestImageDockerfileCommands:
    def test_add_dockerfile_commands(self):
        img = Image.base("python:3.12").dockerfile_commands(
            ["RUN echo 'Hello, world!'", "EXPOSE 8080"]
        )
        df = img.dockerfile()
        assert "RUN echo 'Hello, world!'" in df
        assert "EXPOSE 8080" in df


class TestImageChaining:
    def test_method_chaining(self):
        img = (
            Image.debian_slim("3.12")
            .pip_install("requests", "flask")
            .env({"APP_ENV": "production"})
            .workdir("/app")
            .run_commands("echo setup done")
        )
        df = img.dockerfile()
        assert "FROM python:" in df
        assert "pip install" in df
        assert "ENV APP_ENV=" in df
        assert "WORKDIR" in df
        assert "RUN echo setup done" in df

    def test_returns_image_instance(self):
        img = Image.base("python:3.12")
        result = img.pip_install("requests")
        assert result is img

        result = img.env({"X": "1"})
        assert result is img

        result = img.workdir("/app")
        assert result is img

        result = img.run_commands("echo hi")
        assert result is img

        result = img.entrypoint(["python"])
        assert result is img

        result = img.cmd(["main.py"])
        assert result is img


class TestImageAddLocal:
    def test_add_local_file(self):
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write("test content")
            f.flush()
            try:
                img = Image.base("python:3.12").add_local_file(f.name, "/app/config.txt")
                df = img.dockerfile()
                assert "COPY" in df
                assert "/app/config.txt" in df
            finally:
                os.unlink(f.name)

    def test_add_local_dir(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            img = Image.base("python:3.12").add_local_dir(tmpdir, "/app/src")
            df = img.dockerfile()
            assert "COPY" in df
            assert "/app/src" in df
