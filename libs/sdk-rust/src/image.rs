// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

/// Represents a Docker image built from a series of Dockerfile instructions.
#[derive(Debug, Clone)]
pub struct DockerImage {
    instructions: Vec<String>,
    contexts: Vec<DockerImageContext>,
}

/// Represents a file or directory to include in the Docker build context.
#[derive(Debug, Clone)]
pub struct DockerImageContext {
    pub source_path: String,
    pub archive_path: String,
}

impl DockerImage {
    /// Create from a base image.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04");
    /// assert_eq!(image.dockerfile(), "FROM ubuntu:22.04");
    /// ```
    pub fn base(image: &str) -> Self {
        DockerImage {
            instructions: vec![format!("FROM {}", image)],
            contexts: Vec::new(),
        }
    }

    /// Create from debian-slim with optional python version.
    ///
    /// If `python_version` is provided, uses `python:{version}-slim` as base.
    /// Otherwise, uses `debian:12-slim`.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::debian_slim(Some("3.11"));
    /// assert_eq!(image.dockerfile(), "FROM python:3.11-slim");
    ///
    /// let image = DockerImage::debian_slim(None);
    /// assert_eq!(image.dockerfile(), "FROM debian:12-slim");
    /// ```
    pub fn debian_slim(python_version: Option<&str>) -> Self {
        let base = match python_version {
            Some(version) => format!("FROM python:{}-slim", version),
            None => "FROM debian:12-slim".to_string(),
        };
        DockerImage {
            instructions: vec![base],
            contexts: Vec::new(),
        }
    }

    /// Create from existing Dockerfile content.
    ///
    /// Parses the content and splits into individual instructions.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::from_dockerfile("FROM ubuntu:22.04\nRUN apt-get update");
    /// assert_eq!(image.dockerfile(), "FROM ubuntu:22.04\nRUN apt-get update");
    /// ```
    pub fn from_dockerfile(dockerfile: &str) -> Self {
        let instructions: Vec<String> = dockerfile.lines().map(|line| line.to_string()).collect();
        DockerImage {
            instructions,
            contexts: Vec::new(),
        }
    }

    /// Add a RUN instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .run("echo hello");
    /// assert!(image.dockerfile().contains("RUN echo hello"));
    /// ```
    pub fn run(mut self, command: &str) -> Self {
        self.instructions.push(format!("RUN {}", command));
        self
    }

    /// Add a pip install instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::debian_slim(Some("3.11"))
    ///     .pip_install(&["numpy", "pandas"]);
    /// assert!(image.dockerfile().contains("RUN pip install numpy pandas"));
    /// ```
    pub fn pip_install(mut self, packages: &[&str]) -> Self {
        let pkgs = packages.join(" ");
        self.instructions.push(format!("RUN pip install {}", pkgs));
        self
    }

    /// Add an apt-get install instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("debian:12-slim")
    ///     .apt_get(&["curl", "wget"]);
    /// assert!(image.dockerfile().contains("RUN apt-get update && apt-get install -y curl wget"));
    /// ```
    pub fn apt_get(mut self, packages: &[&str]) -> Self {
        let pkgs = packages.join(" ");
        self.instructions
            .push(format!("RUN apt-get update && apt-get install -y {}", pkgs));
        self
    }

    /// Add an ENV instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .env("MY_VAR", "value");
    /// assert!(image.dockerfile().contains("ENV MY_VAR=\"value\""));
    /// ```
    pub fn env(mut self, key: &str, value: &str) -> Self {
        self.instructions.push(format!("ENV {}=\"{}\"", key, value));
        self
    }

    /// Add a WORKDIR instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .workdir("/app");
    /// assert!(image.dockerfile().contains("WORKDIR /app"));
    /// ```
    pub fn workdir(mut self, path: &str) -> Self {
        self.instructions.push(format!("WORKDIR {}", path));
        self
    }

    /// Add an ENTRYPOINT instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .entrypoint(&["python3", "-m", "http.server"]);
    /// assert!(image.dockerfile().contains(r#"ENTRYPOINT ["python3", "-m", "http.server"]"#));
    /// ```
    pub fn entrypoint(mut self, cmd: &[&str]) -> Self {
        let args = cmd
            .iter()
            .map(|s| format!("\"{}\"", s))
            .collect::<Vec<_>>()
            .join(", ");
        self.instructions.push(format!("ENTRYPOINT [{}]", args));
        self
    }

    /// Add a CMD instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .cmd(&["echo", "hello"]);
    /// assert!(image.dockerfile().contains(r#"CMD ["echo", "hello"]"#));
    /// ```
    pub fn cmd(mut self, cmd: &[&str]) -> Self {
        let args = cmd
            .iter()
            .map(|s| format!("\"{}\"", s))
            .collect::<Vec<_>>()
            .join(", ");
        self.instructions.push(format!("CMD [{}]", args));
        self
    }

    /// Add a USER instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .user("nonroot");
    /// assert!(image.dockerfile().contains("USER nonroot"));
    /// ```
    pub fn user(mut self, username: &str) -> Self {
        self.instructions.push(format!("USER {}", username));
        self
    }

    /// Add a COPY instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .copy("./src", "/app/src");
    /// assert!(image.dockerfile().contains("COPY ./src /app/src"));
    /// ```
    pub fn copy(mut self, src: &str, dst: &str) -> Self {
        self.instructions.push(format!("COPY {} {}", src, dst));
        self
    }

    /// Add an ADD instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .add("./archive.tar.gz", "/app");
    /// assert!(image.dockerfile().contains("ADD ./archive.tar.gz /app"));
    /// ```
    pub fn add(mut self, src: &str, dst: &str) -> Self {
        self.instructions.push(format!("ADD {} {}", src, dst));
        self
    }

    /// Add EXPOSE instructions for the given ports.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .expose(&[8080, 443]);
    /// let df = image.dockerfile();
    /// assert!(df.contains("EXPOSE 8080"));
    /// assert!(df.contains("EXPOSE 443"));
    /// ```
    pub fn expose(mut self, ports: &[i32]) -> Self {
        for port in ports {
            self.instructions.push(format!("EXPOSE {}", port));
        }
        self
    }

    /// Add a LABEL instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .label("maintainer", "daytona");
    /// assert!(image.dockerfile().contains("LABEL maintainer=\"daytona\""));
    /// ```
    pub fn label(mut self, key: &str, value: &str) -> Self {
        self.instructions
            .push(format!("LABEL {}=\"{}\"", key, value));
        self
    }

    /// Add a VOLUME instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .volume(&["/data", "/logs"]);
    /// assert!(image.dockerfile().contains(r#"VOLUME ["/data", "/logs"]"#));
    /// ```
    pub fn volume(mut self, paths: &[&str]) -> Self {
        let vols = paths
            .iter()
            .map(|s| format!("\"{}\"", s))
            .collect::<Vec<_>>()
            .join(", ");
        self.instructions.push(format!("VOLUME [{}]", vols));
        self
    }

    /// Add a local file to the build context and generate an ADD instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .add_local_file("./config.toml", "/etc/app/config.toml");
    /// assert_eq!(image.contexts().len(), 1);
    /// assert!(image.dockerfile().contains("ADD ./config.toml /etc/app/config.toml"));
    /// ```
    pub fn add_local_file(mut self, local: &str, remote: &str) -> Self {
        self.contexts.push(DockerImageContext {
            source_path: local.to_string(),
            archive_path: remote.to_string(),
        });
        self.add(local, remote)
    }

    /// Add a local directory to the build context and generate an ADD instruction.
    ///
    /// # Example
    /// ```
    /// use daytona::DockerImage;
    /// let image = DockerImage::base("ubuntu:22.04")
    ///     .add_local_dir("./src", "/app/src");
    /// assert_eq!(image.contexts().len(), 1);
    /// assert!(image.dockerfile().contains("ADD ./src /app/src"));
    /// ```
    pub fn add_local_dir(mut self, local: &str, remote: &str) -> Self {
        self.contexts.push(DockerImageContext {
            source_path: local.to_string(),
            archive_path: remote.to_string(),
        });
        self.add(local, remote)
    }

    /// Generate Dockerfile content from all instructions.
    pub fn dockerfile(&self) -> String {
        self.instructions.join("\n")
    }

    /// Get contexts for object storage upload.
    pub fn contexts(&self) -> &[DockerImageContext] {
        &self.contexts
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_base_image() {
        let image = DockerImage::base("ubuntu:22.04");
        assert_eq!(image.dockerfile(), "FROM ubuntu:22.04");
    }

    #[test]
    fn test_debian_slim_with_python() {
        let image = DockerImage::debian_slim(Some("3.11"));
        assert_eq!(image.dockerfile(), "FROM python:3.11-slim");
    }

    #[test]
    fn test_debian_slim_without_python() {
        let image = DockerImage::debian_slim(None);
        assert_eq!(image.dockerfile(), "FROM debian:12-slim");
    }

    #[test]
    fn test_from_dockerfile() {
        let content = "FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl";
        let image = DockerImage::from_dockerfile(content);
        assert_eq!(image.dockerfile(), content);
    }

    #[test]
    fn test_builder_chain() {
        let image = DockerImage::debian_slim(Some("3.11"))
            .env("APP_ENV", "production")
            .workdir("/app")
            .apt_get(&["curl", "git"])
            .pip_install(&["flask", "gunicorn"])
            .copy(".", "/app")
            .expose(&[8080])
            .user("appuser")
            .entrypoint(&["python3"])
            .cmd(&["-m", "flask", "run"]);

        let df = image.dockerfile();
        assert!(df.starts_with("FROM python:3.11-slim"));
        assert!(df.contains("ENV APP_ENV=\"production\""));
        assert!(df.contains("WORKDIR /app"));
        assert!(df.contains("RUN apt-get update && apt-get install -y curl git"));
        assert!(df.contains("RUN pip install flask gunicorn"));
        assert!(df.contains("COPY . /app"));
        assert!(df.contains("EXPOSE 8080"));
        assert!(df.contains("USER appuser"));
        assert!(df.contains("ENTRYPOINT [\"python3\"]"));
        assert!(df.contains("CMD [\"-m\", \"flask\", \"run\"]"));
    }

    #[test]
    fn test_add_local_file_creates_context() {
        let image =
            DockerImage::base("ubuntu:22.04").add_local_file("./config.toml", "/etc/config.toml");

        assert_eq!(image.contexts().len(), 1);
        assert_eq!(image.contexts()[0].source_path, "./config.toml");
        assert_eq!(image.contexts()[0].archive_path, "/etc/config.toml");
        assert!(image
            .dockerfile()
            .contains("ADD ./config.toml /etc/config.toml"));
    }

    #[test]
    fn test_add_local_dir_creates_context() {
        let image = DockerImage::base("ubuntu:22.04").add_local_dir("./src", "/app/src");

        assert_eq!(image.contexts().len(), 1);
        assert_eq!(image.contexts()[0].source_path, "./src");
        assert_eq!(image.contexts()[0].archive_path, "/app/src");
    }

    #[test]
    fn test_labels_and_volumes() {
        let image = DockerImage::base("ubuntu:22.04")
            .label("version", "1.0")
            .volume(&["/data", "/logs"]);

        let df = image.dockerfile();
        assert!(df.contains("LABEL version=\"1.0\""));
        assert!(df.contains("VOLUME [\"/data\", \"/logs\"]"));
    }
}
