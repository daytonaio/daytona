

import json
import logging
import os
import tempfile
from contextlib import contextmanager
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from daytona_sdk import CreateSandboxParams, Daytona, FileUpload


class DaytonaManager:
    """Manages Daytona sandbox operations with proper error handling and cleanup."""
    
    def __init__(self, language: str = "python"):
        """Initialize the Daytona manager.
        
        Args:
            language: Programming language for the sandbox environment
        """
        self.daytona = Daytona()
        self.language = language
        self.sandbox = None
        self.logger = self._setup_logging()
        
    def _setup_logging(self) -> logging.Logger:
        """Set up structured logging for better debugging and monitoring."""
        logger = logging.getLogger(__name__)
        logger.setLevel(logging.INFO)
        
        if not logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            handler.setFormatter(formatter)
            logger.addHandler(handler)
            
        return logger
    
    @contextmanager
    def sandbox_context(self):
        """Context manager for automatic sandbox cleanup."""
        try:
            params = CreateSandboxParams(language=self.language)
            self.sandbox = self.daytona.create(params)
            self.logger.info(f"Created sandbox with ID: {self.sandbox.id}")
            yield self.sandbox
        except Exception as e:
            self.logger.error(f"Failed to create sandbox: {e}")
            raise
        finally:
            if self.sandbox:
                try:
                    self.daytona.delete(self.sandbox)
                    self.logger.info(f"Deleted sandbox {self.sandbox.id}")
                except Exception as e:
                    self.logger.error(f"Failed to delete sandbox: {e}")
    
    def create_project_structure(self, base_dir: str) -> Dict[str, str]:
        """Create a structured project directory layout.
        
        Args:
            base_dir: Base directory path for the project
            
        Returns:
            Dictionary mapping directory purposes to their paths
        """
        directories = {
            "base": base_dir,
            "config": os.path.join(base_dir, "config"),
            "scripts": os.path.join(base_dir, "scripts"),
            "data": os.path.join(base_dir, "data"),
            "logs": os.path.join(base_dir, "logs")
        }
        
        for name, path in directories.items():
            try:
                self.sandbox.fs.create_folder(path, "755")
                self.logger.info(f"Created {name} directory: {path}")
            except Exception as e:
                self.logger.error(f"Failed to create {name} directory {path}: {e}")
                raise
                
        return directories
    
    def generate_config_file(self, environment: str = "development") -> str:
        """Generate a comprehensive configuration file.
        
        Args:
            environment: Target environment (development, staging, production)
            
        Returns:
            JSON string containing the configuration
        """
        config = {
            "project": {
                "name": "daytona-demo-project",
                "version": "2.0.0",
                "description": "Demonstration project for Daytona SDK capabilities",
                "created_at": datetime.now().isoformat(),
                "environment": environment
            },
            "settings": {
                "debug": environment == "development",
                "max_connections": 10 if environment == "development" else 100,
                "timeout": 30,
                "retry_attempts": 3,
                "enable_metrics": True,
                "log_level": "DEBUG" if environment == "development" else "INFO"
            },
            "features": {
                "file_upload": True,
                "command_execution": True,
                "search_functionality": True,
                "content_replacement": True
            },
            "paths": {
                "temp": "/tmp",
                "logs": "./logs",
                "data": "./data"
            }
        }
        
        return json.dumps(config, indent=2, sort_keys=True)
    
    def create_demo_script(self) -> str:
        """Create a demonstration shell script with error handling."""
        script_content = """#!/bin/bash
# Daytona Demo Script
# This script demonstrates basic system operations

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "=== Daytona Demo Script ==="
echo "Timestamp: $(date)"
echo "Current user: $(whoami)"
echo "Working directory: $(pwd)"
echo "System info: $(uname -a)"

# Create a simple log entry
echo "Script executed successfully at $(date)" >> ./logs/execution.log

# Test Python availability
if command -v python3 &> /dev/null; then
    echo "Python version: $(python3 --version)"
    python3 -c "import sys; print(f'Python path: {sys.executable}')"
else
    echo "Python3 not found"
fi

echo "=== Demo completed successfully ==="
exit 0
"""
        return script_content
    
    def upload_project_files(self, directories: Dict[str, str]) -> List[FileUpload]:
        """Upload all project files to the sandbox.
        
        Args:
            directories: Dictionary of directory paths
            
        Returns:
            List of FileUpload objects that were uploaded
        """
        uploads = []
        
        try:
            # Generate dynamic content
            config_content = self.generate_config_file()
            script_content = self.create_demo_script()
            
            # Create sample data file
            sample_data = {
                "users": [
                    {"id": 1, "name": "Alice", "role": "admin"},
                    {"id": 2, "name": "Bob", "role": "user"},
                    {"id": 3, "name": "Charlie", "role": "moderator"}
                ],
                "metadata": {
                    "total_users": 3,
                    "created_at": datetime.now().isoformat(),
                    "version": "1.0"
                }
            }
            
            # Create README file
            readme_content = """# Daytona Demo Project

This project demonstrates the capabilities of the Daytona SDK for file management
and sandbox operations.

## Structure

- `config/`: Configuration files
- `scripts/`: Executable scripts
- `data/`: Sample data files
- `logs/`: Application logs

## Usage

Run the main script to see the demo in action:

```bash
./scripts/demo.sh
```

## Features

- Automated file uploads
- Command execution
- File search and replacement
- Structured logging
- Error handling
"""
            
            # Define file uploads
            file_uploads = [
                FileUpload(
                    source=config_content.encode("utf-8"),
                    destination=os.path.join(directories["config"], "app.json")
                ),
                FileUpload(
                    source=script_content.encode("utf-8"),
                    destination=os.path.join(directories["scripts"], "demo.sh")
                ),
                FileUpload(
                    source=json.dumps(sample_data, indent=2).encode("utf-8"),
                    destination=os.path.join(directories["data"], "sample.json")
                ),
                FileUpload(
                    source=readme_content.encode("utf-8"),
                    destination=os.path.join(directories["base"], "README.md")
                ),
                FileUpload(
                    source=b"# Execution Log\n",
                    destination=os.path.join(directories["logs"], "execution.log")
                )
            ]
            
            # Upload all files
            self.sandbox.fs.upload_files(file_uploads)
            self.logger.info(f"Successfully uploaded {len(file_uploads)} files")
            
            return file_uploads
            
        except Exception as e:
            self.logger.error(f"Failed to upload files: {e}")
            raise
    
    def execute_commands(self, directories: Dict[str, str]) -> Dict[str, any]:
        """Execute various commands to demonstrate sandbox capabilities.
        
        Args:
            directories: Dictionary of directory paths
            
        Returns:
            Dictionary containing execution results
        """
        results = {}
        
        commands = [
            ("list_files", f"find {directories['base']} -type f -name '*' | head -20"),
            ("make_executable", f"chmod +x {os.path.join(directories['scripts'], 'demo.sh')}"),
            ("run_script", os.path.join(directories["scripts"], "demo.sh")),
            ("check_permissions", f"ls -la {directories['scripts']}/"),
            ("disk_usage", f"du -sh {directories['base']}"),
        ]
        
        for name, command in commands:
            try:
                self.logger.info(f"Executing: {command}")
                result = self.sandbox.process.exec(command)
                results[name] = {
                    "command": command,
                    "exit_code": result.exit_code,
                    "output": result.result,
                    "success": result.exit_code == 0
                }
                
                if result.exit_code == 0:
                    self.logger.info(f"✓ {name} completed successfully")
                else:
                    self.logger.warning(f"✗ {name} failed with exit code {result.exit_code}")
                    
            except Exception as e:
                self.logger.error(f"Failed to execute {name}: {e}")
                results[name] = {
                    "command": command,
                    "error": str(e),
                    "success": False
                }
        
        return results
    
    def perform_file_operations(self, directories: Dict[str, str]) -> Dict[str, any]:
        """Demonstrate file search and modification operations.
        
        Args:
            directories: Dictionary of directory paths
            
        Returns:
            Dictionary containing operation results
        """
        operations = {}
        
        try:
            # Search for JSON files
            json_matches = self.sandbox.fs.search_files(directories["base"], "*.json")
            operations["json_search"] = {
                "pattern": "*.json",
                "matches": len(json_matches.files),
                "files": json_matches.files
            }
            self.logger.info(f"Found {len(json_matches.files)} JSON files")
            
            # Search for shell scripts
            script_matches = self.sandbox.fs.search_files(directories["base"], "*.sh")
            operations["script_search"] = {
                "pattern": "*.sh", 
                "matches": len(script_matches.files),
                "files": script_matches.files
            }
            self.logger.info(f"Found {len(script_matches.files)} shell scripts")
            
            # Modify configuration file
            config_file = os.path.join(directories["config"], "app.json")
            self.sandbox.fs.replace_in_files(
                [config_file], 
                '"debug": true', 
                '"debug": false'
            )
            operations["config_modification"] = {
                "file": config_file,
                "change": "debug: true -> false",
                "success": True
            }
            self.logger.info("Modified configuration file")
            
            # Download and verify the modified file
            modified_content = self.sandbox.fs.download_file(config_file)
            operations["file_download"] = {
                "file": config_file,
                "size": len(modified_content),
                "contains_debug_false": b'"debug": false' in modified_content
            }
            
        except Exception as e:
            self.logger.error(f"File operation failed: {e}")
            operations["error"] = str(e)
            
        return operations
    
    def generate_comprehensive_report(self, 
                                   directories: Dict[str, str],
                                   command_results: Dict[str, any],
                                   file_operations: Dict[str, any]) -> str:
        """Generate a detailed report of all operations performed.
        
        Args:
            directories: Dictionary of directory paths
            command_results: Results from command execution
            file_operations: Results from file operations
            
        Returns:
            Formatted report string
        """
        successful_commands = sum(1 for r in command_results.values() 
                                if isinstance(r, dict) and r.get("success", False))
        
        report = f"""
# Daytona SDK Demo Report
========================

**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
**Environment:** {self.language}
**Sandbox ID:** {self.sandbox.id if self.sandbox else 'N/A'}

## Project Structure
- Base Directory: `{directories.get('base', 'N/A')}`
- Configuration: `{directories.get('config', 'N/A')}`
- Scripts: `{directories.get('scripts', 'N/A')}`
- Data: `{directories.get('data', 'N/A')}`
- Logs: `{directories.get('logs', 'N/A')}`

## Command Execution Summary
- Total Commands: {len(command_results)}
- Successful: {successful_commands}
- Failed: {len(command_results) - successful_commands}

### Command Details:
"""
        
        for name, result in command_results.items():
            if isinstance(result, dict):
                status = "✓ PASS" if result.get("success", False) else "✗ FAIL"
                report += f"- **{name}**: {status}\n"
                if not result.get("success", False) and "error" in result:
                    report += f"  Error: {result['error']}\n"
        
        report += f"""
## File Operations Summary
- JSON Files Found: {file_operations.get('json_search', {}).get('matches', 0)}
- Shell Scripts Found: {file_operations.get('script_search', {}).get('matches', 0)}
- Configuration Modified: {'Yes' if 'config_modification' in file_operations else 'No'}
- File Download: {'Success' if 'file_download' in file_operations else 'Failed'}

## Status
Overall Status: {'SUCCESS' if successful_commands > 0 and 'error' not in file_operations else 'PARTIAL SUCCESS'}

---
*Report generated by Daytona SDK Demo Script*
        """.strip()
        
        return report


def main():
    """Main execution function with comprehensive error handling."""
    manager = DaytonaManager()
    
    try:
        with manager.sandbox_context() as sandbox:
            manager.logger.info("Step 1: Creating project structure")
            directories = manager.create_project_structure("~/daytona-demo")
            
            manager.logger.info("Step 2: Uploading project files")
            uploaded_files = manager.upload_project_files(directories)
            
            manager.logger.info("Step 3: Executing demonstration commands")
            command_results = manager.execute_commands(directories)
            
            manager.logger.info("Step 4: Performing file operations")
            file_operations = manager.perform_file_operations(directories)
            
            manager.logger.info("Step 5: Generating comprehensive report")
            report = manager.generate_comprehensive_report(
                directories, command_results, file_operations
            )
            
            report_path = os.path.join(directories["base"], "demo-report.md")
            sandbox.fs.upload_file(report.encode("utf-8"), report_path)
            
            print("\n" + "="*60)
            print("DAYTONA SDK DEMO COMPLETED SUCCESSFULLY")
            print("="*60)
            print(f"Sandbox ID: {sandbox.id}")
            print(f"Files uploaded: {len(uploaded_files)}")
            print(f"Commands executed: {len(command_results)}")
            print(f"Report saved to: {report_path}")
            print("="*60)
            
            print(report)
            
    except Exception as e:
        manager.logger.error(f"Demo failed: {e}")
        raise


if __name__ == "__main__":
    main()
