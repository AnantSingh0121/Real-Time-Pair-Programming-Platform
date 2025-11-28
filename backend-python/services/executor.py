import subprocess
import sys
import tempfile
import os
from typing import Dict, Any
import time

class CodeExecutor:
    def __init__(self):
        self.timeout = 10 
        self.max_output_size = 10000  

    def execute_python(self, code: str, stdin: str = "") -> Dict[str, Any]:
        try:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
                f.write(code)
                temp_file = f.name

            start_time = time.time()
            
            result = subprocess.run(
                [sys.executable, temp_file],
                input=stdin,
                capture_output=True,
                text=True,
                timeout=self.timeout,
                env={**os.environ, 'PYTHONIOENCODING': 'utf-8'}
            )
            
            execution_time = time.time() - start_time
            
            os.unlink(temp_file)
            
            output = result.stdout
            error = result.stderr
            
            if len(output) > self.max_output_size:
                output = output[:self.max_output_size] + "\n... (output truncated)"
            
            if len(error) > self.max_output_size:
                error = error[:self.max_output_size] + "\n... (error truncated)"
            
            return {
                "success": result.returncode == 0,
                "output": output,
                "error": error,
                "executionTime": round(execution_time, 3),
                "returnCode": result.returncode
            }
            
        except subprocess.TimeoutExpired:
            os.unlink(temp_file)
            return {
                "success": False,
                "output": "",
                "error": f"Execution timed out after {self.timeout} seconds",
                "executionTime": self.timeout,
                "returnCode": -1
            }
        except Exception as e:
            return {
                "success": False,
                "output": "",
                "error": f"Execution error: {str(e)}",
                "executionTime": 0,
                "returnCode": -1
            }

    def execute_javascript(self, code: str, stdin: str = "") -> Dict[str, Any]:
        try:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.js', delete=False) as f:
                f.write(code)
                temp_file = f.name

            start_time = time.time()
            
            result = subprocess.run(
                ['node', temp_file],
                input=stdin,
                capture_output=True,
                text=True,
                timeout=self.timeout
            )
            
            execution_time = time.time() - start_time
            os.unlink(temp_file)
            
            return {
                "success": result.returncode == 0,
                "output": result.stdout,
                "error": result.stderr,
                "executionTime": round(execution_time, 3),
                "returnCode": result.returncode
            }
            
        except FileNotFoundError:
            return {
                "success": False,
                "output": "",
                "error": "Node.js is not installed. Please install Node.js to run JavaScript code.",
                "executionTime": 0,
                "returnCode": -1
            }
        except subprocess.TimeoutExpired:
            os.unlink(temp_file)
            return {
                "success": False,
                "output": "",
                "error": f"Execution timed out after {self.timeout} seconds",
                "executionTime": self.timeout,
                "returnCode": -1
            }
        except Exception as e:
            return {
                "success": False,
                "output": "",
                "error": f"Execution error: {str(e)}",
                "executionTime": 0,
                "returnCode": -1
            }

    def execute_go(self, code: str, stdin: str = "") -> Dict[str, Any]:
        try:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.go', delete=False) as f:
                f.write(code)
                temp_file = f.name

            start_time = time.time()
            
            result = subprocess.run(
                ['go', 'run', temp_file],
                input=stdin,
                capture_output=True,
                text=True,
                timeout=self.timeout
            )
            
            execution_time = time.time() - start_time
            os.unlink(temp_file)
            
            return {
                "success": result.returncode == 0,
                "output": result.stdout,
                "error": result.stderr,
                "executionTime": round(execution_time, 3),
                "returnCode": result.returncode
            }
        except FileNotFoundError:
             return {
                "success": False,
                "output": "",
                "error": "Go is not installed.",
                "executionTime": 0,
                "returnCode": -1
            }
        except subprocess.TimeoutExpired:
            os.unlink(temp_file)
            return {
                "success": False,
                "output": "",
                "error": f"Execution timed out after {self.timeout} seconds",
                "executionTime": self.timeout,
                "returnCode": -1
            }
        except Exception as e:
            if os.path.exists(temp_file):
                os.unlink(temp_file)
            return {
                "success": False,
                "output": "",
                "error": f"Execution error: {str(e)}",
                "executionTime": 0,
                "returnCode": -1
            }

    def execute_cpp(self, code: str, stdin: str = "") -> Dict[str, Any]:
        try:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.cpp', delete=False) as f:
                f.write(code)
                source_file = f.name
            
            exe_file = source_file.replace('.cpp', '.exe')
            start_time = time.time()           
            compile_result = subprocess.run(
                ['g++', source_file, '-o', exe_file],
                capture_output=True,
                text=True,
                timeout=self.timeout
            )

            if compile_result.returncode != 0:
                os.unlink(source_file)
                return {
                    "success": False,
                    "output": "",
                    "error": f"Compilation Error:\n{compile_result.stderr}",
                    "executionTime": 0,
                    "returnCode": compile_result.returncode
                }

            result = subprocess.run(
                [exe_file],
                input=stdin,
                capture_output=True,
                text=True,
                timeout=self.timeout
            )
            
            execution_time = time.time() - start_time
            
            os.unlink(source_file)
            if os.path.exists(exe_file):
                os.unlink(exe_file)
            
            return {
                "success": result.returncode == 0,
                "output": result.stdout,
                "error": result.stderr,
                "executionTime": round(execution_time, 3),
                "returnCode": result.returncode
            }
        except FileNotFoundError:
             return {
                "success": False,
                "output": "",
                "error": "G++ is not installed.",
                "executionTime": 0,
                "returnCode": -1
            }
        except subprocess.TimeoutExpired:
            if os.path.exists(source_file):
                os.unlink(source_file)
            if os.path.exists(exe_file):
                os.unlink(exe_file)
            return {
                "success": False,
                "output": "",
                "error": f"Execution timed out after {self.timeout} seconds",
                "executionTime": self.timeout,
                "returnCode": -1
            }
        except Exception as e:
            return {
                "success": False,
                "output": "",
                "error": f"Execution error: {str(e)}",
                "executionTime": 0,
                "returnCode": -1
            }

    def execute(self, code: str, language: str, stdin: str = "") -> Dict[str, Any]:
        language = language.lower()
        
        if language in ['python', 'py']:
            return self.execute_python(code, stdin)
        elif language in ['javascript', 'js', 'node']:
            return self.execute_javascript(code, stdin)
        elif language in ['go', 'golang']:
            return self.execute_go(code, stdin)
        elif language in ['cpp', 'c++', 'c']:
            return self.execute_cpp(code, stdin)
        else:
            return {
                "success": False,
                "output": "",
                "error": f"Unsupported language: {language}. Supported: Python, JS, Go, C++",
                "executionTime": 0,
                "returnCode": -1
            }
