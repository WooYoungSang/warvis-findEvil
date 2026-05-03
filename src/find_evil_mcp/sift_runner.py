# SPDX-License-Identifier: MIT

import subprocess
import shutil
import time
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Mapping, Optional


# Binary whitelist: logical name -> binary name (as found in PATH or via full path)
BINARIES = {
    "plaso": "log2timeline.py",
    "log2timeline": "log2timeline.py",
    "yara": "yara",
    "sigma": "sigma",
    "volatility3": "vol",
    "zeek": "zeek",
    "suricata": "suricata",
}


@dataclass
class SiftResult:
    """Result of a SIFT subprocess execution."""

    tool: str
    returncode: int
    stdout: bytes
    stderr: bytes
    duration_s: float


class SiftRunner:
    """Secure wrapper for SIFT tool subprocess calls."""

    def __init__(
        self,
        *,
        root: Optional[Path] = None,
        env: Optional[Mapping[str, str]] = None,
        timeout: float = 120.0,
    ):
        """
        Initialize SiftRunner.

        Args:
            root: Case root directory (defaults to FIND_EVIL_CASES_ROOT env var or ./.cases)
            env: Environment variables (defaults to os.environ)
            timeout: Subprocess timeout in seconds (default 120s)
        """
        if root is None:
            root_str = os.environ.get("FIND_EVIL_CASES_ROOT", "./.cases")
            root = Path(root_str)
        self.root = root

        self.env = dict(env) if env else dict(os.environ)
        self.timeout = timeout

    def is_available(self, tool: str) -> bool:
        """
        Check if a tool is available in the system.

        Args:
            tool: Logical tool name (e.g., "yara", "volatility3")

        Returns:
            True if the binary is found via shutil.which(), False otherwise.
        """
        if tool not in BINARIES:
            return False

        binary_name = BINARIES[tool]
        return shutil.which(binary_name) is not None

    def run(
        self,
        tool: str,
        args: list[str],
        *,
        cwd: Optional[Path] = None,
        stdin: Optional[bytes] = None,
    ) -> SiftResult:
        """
        Run a whitelisted SIFT tool subprocess.

        Args:
            tool: Logical tool name from BINARIES whitelist
            args: Tool arguments (metacharacters will be rejected)
            cwd: Working directory (defaults to self.root)
            stdin: Optional stdin bytes

        Returns:
            SiftResult with returncode, stdout, stderr, duration_s

        Raises:
            ValueError: If tool not in whitelist or args contain forbidden chars
            subprocess.TimeoutExpired: If timeout exceeded
        """
        # Validate tool is whitelisted
        if tool not in BINARIES:
            raise ValueError(f"Tool not whitelisted: {tool}")

        # Validate args for shell metacharacters
        forbidden_chars = set(";|`$&<>\\n\\r\0")
        for arg in args:
            if any(c in arg for c in forbidden_chars):
                raise ValueError(f"Forbidden character in argument: {arg}")

        # Set working directory
        if cwd is None:
            cwd = self.root

        # Build command: binary name + args (no shell)
        binary_name = BINARIES[tool]
        cmd = [binary_name] + args

        # Execute subprocess
        start_time = time.time()
        try:
            result = subprocess.run(
                cmd,
                cwd=str(cwd),
                capture_output=True,
                timeout=self.timeout,
                env=self.env,
                shell=False,  # Explicit: no shell
                check=False,  # Don't raise on non-zero exit
            )
            duration_s = time.time() - start_time

            return SiftResult(
                tool=tool,
                returncode=result.returncode,
                stdout=result.stdout,
                stderr=result.stderr,
                duration_s=duration_s,
            )

        except subprocess.TimeoutExpired:
            duration_s = time.time() - start_time
            # Return partial result with timeout indicator in stderr
            return SiftResult(
                tool=tool,
                returncode=-1,
                stdout=b"",
                stderr=f"TIMEOUT after {self.timeout}s".encode("utf-8"),
                duration_s=duration_s,
            )
