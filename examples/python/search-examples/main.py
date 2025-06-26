#!/usr/bin/env python3
"""
Comprehensive examples demonstrating Daytona's enhanced search functionality.
"""

import time

from daytona import Daytona, FileUpload
from daytona._sync.filesystem import SearchRequest


def main():
    daytona = Daytona()

    # Create a sandbox for testing search functionality
    sandbox = daytona.create()

    try:
        print(f"Created sandbox with ID: {sandbox.id}")

        # Create some sample files to search through
        print("Setting up sample files for search demonstration...")

        # Create a project structure
        sandbox.fs.create_folder("~/search-demo", "755")
        sandbox.fs.create_folder("~/search-demo/src", "755")
        sandbox.fs.create_folder("~/search-demo/tests", "755")
        sandbox.fs.create_folder("~/search-demo/docs", "755")

        # Upload sample files with different content
        sandbox.fs.upload_files(
            [
                FileUpload(
                    source=b"""# Main application module
import logging
import json
from typing import Dict, Any

class Application:
    \"\"\"Main application class.\"\"\"
    
    def __init__(self, config_path: str = "config.json"):
        self.logger = logging.getLogger(__name__)
        self.config = self._load_config(config_path)
        # TODO: Add error handling for config loading
    
    def _load_config(self, path: str) -> Dict[str, Any]:
        \"\"\"Load configuration from JSON file.\"\"\"
        try:
            with open(path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            self.logger.error(f"Config file not found: {path}")
            return {}
    
    async def start(self) -> None:
        \"\"\"Start the application.\"\"\"
        self.logger.info("Starting application...")
        # TODO: Implement startup logic
    
    async def stop(self) -> None:
        \"\"\"Stop the application.\"\"\"
        self.logger.info("Stopping application...")

if __name__ == "__main__":
    app = Application()
""",
                    destination="~/search-demo/src/app.py",
                ),
                FileUpload(
                    source=b"""# Search Demo Project

This is a sample project for demonstrating search functionality.

## Features

- Advanced search capabilities with ripgrep
- File type filtering
- Pattern matching with regex support
- Context-aware results
- Performance optimizations

## TODO Items

- [ ] Add more comprehensive examples
- [ ] Improve documentation with screenshots
- [ ] Add unit tests for all search features
- [ ] Benchmark performance against other tools

## Configuration

The application uses a JSON configuration file for settings.
See `config.json` for available options.

## Usage

```python
from daytona._sync.filesystem import SearchRequest

# Basic search
results = sandbox.fs.search(SearchRequest(
    query="TODO",
    path=".",
    case_sensitive=False
))
```
""",
                    destination="~/search-demo/docs/README.md",
                ),
                FileUpload(
                    source=b"""{
  "name": "search-demo",
  "version": "1.0.0",
  "description": "Demo project for search functionality",
  "main": "src/app.py",
  "scripts": {
    "start": "python src/app.py",
    "test": "pytest tests/",
    "lint": "flake8 src/ tests/"
  },
  "dependencies": {
    "requests": "^2.28.0",
    "click": "^8.1.0",
    "pydantic": "^1.10.0"
  },
  "devDependencies": {
    "pytest": "^7.2.0",
    "flake8": "^5.0.0",
    "black": "^22.10.0",
    "mypy": "^0.991"
  },
  "keywords": ["search", "demo", "python", "daytona"],
  "author": "Daytona Team",
  "license": "MIT"
}
""",
                    destination="~/search-demo/config.json",
                ),
                FileUpload(
                    source=b"""import pytest
from unittest.mock import Mock, patch
from src.app import Application


class TestApplication:
    \"\"\"Test cases for the Application class.\"\"\"
    
    def setup_method(self):
        \"\"\"Set up test fixtures.\"\"\"
        self.app = Application()
    
    def test_init(self):
        \"\"\"Test application initialization.\"\"\"
        assert self.app is not None
        assert hasattr(self.app, 'logger')
        assert hasattr(self.app, 'config')
    
    @pytest.mark.asyncio
    async def test_start(self):
        \"\"\"Test application start method.\"\"\"
        # TODO: Add proper test implementation
        await self.app.start()
        # Add assertions here
    
    @pytest.mark.asyncio  
    async def test_stop(self):
        \"\"\"Test application stop method.\"\"\"
        # TODO: Add proper test implementation
        await self.app.stop()
        # Add assertions here
    
    def test_load_config_file_not_found(self):
        \"\"\"Test config loading with missing file.\"\"\"
        with patch('builtins.open', side_effect=FileNotFoundError):
            config = self.app._load_config("nonexistent.json")
            assert config == {}
""",
                    destination="~/search-demo/tests/test_app.py",
                ),
                FileUpload(
                    source=b"""#!/bin/bash

# Build script for the Python project
echo "Building Python project..."

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    python -m venv venv
    echo "Virtual environment created"
fi

# Activate virtual environment
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run linting
echo "Running linting..."
flake8 src/ tests/

# Run type checking
echo "Running type checking..."
mypy src/

# Run tests
echo "Running tests..."
pytest tests/ -v

echo "Build completed successfully!"
""",
                    destination="~/search-demo/build.sh",
                ),
            ]
        )

        print("Sample files created. Starting search demonstrations...\n")

        # === SEARCH DEMONSTRATIONS ===

        # 1. Basic text search
        print("=== 1. Basic Text Search ===")
        basic_search = sandbox.fs.search(SearchRequest(query="TODO", path="~/search-demo"))
        print(f"Found {basic_search.total_matches} TODO items in {basic_search.total_files} files:")
        for match in basic_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.line.strip()}")

        # 2. Case-insensitive search
        print("\n=== 2. Case-Insensitive Search ===")
        case_insensitive_search = sandbox.fs.search(
            SearchRequest(query="application", path="~/search-demo", case_sensitive=False)
        )
        print(f"Found {case_insensitive_search.total_matches} matches for 'application' (case-insensitive):")
        for match in case_insensitive_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.match}")

        # 3. File type filtering
        print("\n=== 3. File Type Filtering ===")
        python_search = sandbox.fs.search(
            SearchRequest(query="class|def|import", path="~/search-demo", file_types=["py"], max_results=10)
        )
        print(f"Found {python_search.total_matches} Python definitions:")
        for match in python_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.line.strip()}")

        # 4. Search with context
        print("\n=== 4. Search with Context ===")
        context_search = sandbox.fs.search(
            SearchRequest(query="class Application", path="~/search-demo", context=2, max_results=3)
        )
        print("Class definitions with context:")
        for match in context_search.matches:
            print(f"\n  {match.file}:{match.line_number}:")
            if match.context_before:
                for i, line in enumerate(match.context_before):
                    print(f"    {match.line_number - len(match.context_before) + i}: {line}")
            print(f"  > {match.line_number}: {match.line}")
            if match.context_after:
                for i, line in enumerate(match.context_after):
                    print(f"    {match.line_number + i + 1}: {line}")

        # 5. Include/Exclude patterns
        print("\n=== 5. Include/Exclude Patterns ===")
        pattern_search = sandbox.fs.search(
            SearchRequest(
                query="test",
                path="~/search-demo",
                include_globs=["*.py", "*.md"],
                exclude_globs=["*test*"],
                case_sensitive=False,
            )
        )
        print(f"Found {pattern_search.total_matches} 'test' matches in source files (excluding test files):")
        for match in pattern_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.line.strip()}")

        # 6. Count-only search
        print("\n=== 6. Count-Only Search ===")
        count_search = sandbox.fs.search(
            SearchRequest(query="import|from.*import", path="~/search-demo", count_only=True)
        )
        print(f"Total import statements: {count_search.total_matches} in {count_search.total_files} files")

        # 7. Filenames-only search
        print("\n=== 7. Filenames-Only Search ===")
        filenames_search = sandbox.fs.search(SearchRequest(query="app", path="~/search-demo", filenames_only=True))
        print(f"Files containing 'app': {', '.join(filenames_search.files or [])}")

        # 8. Advanced regex search
        print("\n=== 8. Advanced Regex Search ===")
        regex_search = sandbox.fs.search(
            SearchRequest(
                query=r'"[^"]*":\s*"[^"]*"',  # JSON key-value pairs
                path="~/search-demo",
                file_types=["json"],
                max_results=5,
            )
        )
        print(f"Found {regex_search.total_matches} JSON key-value pairs:")
        for match in regex_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.match}")

        # 9. Multiline search
        print("\n=== 9. Multiline Search ===")
        multiline_search = sandbox.fs.search(
            SearchRequest(query=r"class.*:\s*\n.*\"\"\".*\"\"\"", path="~/search-demo", multiline=True, max_results=3)
        )
        print(f"Found {multiline_search.total_matches} docstring patterns:")
        for match in multiline_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.match[:50]}...")

        # 10. Performance comparison
        print("\n=== 10. Performance Comparison ===")
        start_time = time.time()
        performance_search = sandbox.fs.search(
            # Match any character (lots of matches)
            SearchRequest(query=".", path="~/search-demo", max_results=1000)
        )
        end_time = time.time()
        print(
            f"Performance test: Found {performance_search.total_matches} matches in "
            f"{(end_time - start_time)*1000:.1f}ms"
        )

        print("\n✅ All search demonstrations completed successfully!")

    except Exception as error:
        print(f"❌ Error during search demonstration: {error}")
    finally:
        # Cleanup
        daytona.delete(sandbox)
        print("Sandbox cleaned up")


if __name__ == "__main__":
    main()
