# SPDX-License-Identifier: MIT

import json
import os
from pathlib import Path
from jsonschema import Draft7Validator, ValidationError

# Cache for loaded schemas and validators
_SCHEMA_CACHE = {}
_VALIDATOR_CACHE = {}


def _get_schema_path(tool_name: str) -> Path:
    """Get absolute path to schema JSON file for a tool."""
    # Walk up to find harness directory (works from any installation location)
    current = Path(__file__).resolve()
    while current != current.parent:
        candidate = current.parent.parent / "harness" / "find-evil" / "mcp-schema" / f"{tool_name}.json"
        if candidate.exists():
            return candidate
        current = current.parent

    # Fallback: relative from package root
    schema_dir = Path(__file__).parent.parent.parent / "harness" / "find-evil" / "mcp-schema"
    schema_file = schema_dir / f"{tool_name}.json"
    return schema_file


def _load_schema(tool_name: str) -> dict:
    """Load JSON schema from disk. Cache result."""
    if tool_name in _SCHEMA_CACHE:
        return _SCHEMA_CACHE[tool_name]

    schema_path = _get_schema_path(tool_name)
    if not schema_path.exists():
        raise FileNotFoundError(f"Schema not found: {schema_path}")

    with open(schema_path) as f:
        schema = json.load(f)

    _SCHEMA_CACHE[tool_name] = schema
    return schema


def _get_validator(tool_name: str, schema_key: str) -> Draft7Validator:
    """Get or create a Draft7Validator for input or output schema."""
    cache_key = f"{tool_name}:{schema_key}"
    if cache_key in _VALIDATOR_CACHE:
        return _VALIDATOR_CACHE[cache_key]

    full_schema = _load_schema(tool_name)
    # Extract inputSchema or outputSchema from properties
    properties = full_schema.get("properties", {})
    schema = properties.get(schema_key)
    if not schema:
        raise ValueError(f"No {schema_key} in schema for {tool_name}")

    validator = Draft7Validator(schema)
    _VALIDATOR_CACHE[cache_key] = validator
    return validator


def validate_input(tool_name: str, payload: dict) -> None:
    """
    Validate input payload against the tool's inputSchema.
    Raises jsonschema.ValidationError if invalid.
    """
    validator = _get_validator(tool_name, "inputSchema")
    validator.validate(payload)


def validate_output(tool_name: str, payload: dict) -> None:
    """
    Validate output payload against the tool's outputSchema.
    Raises jsonschema.ValidationError if invalid.
    """
    validator = _get_validator(tool_name, "outputSchema")
    validator.validate(payload)


def get_output_schema(tool_name: str) -> dict:
    """
    Get the outputSchema dict for a tool.
    Used by server.py to populate Tool.outputSchema field.
    """
    full_schema = _load_schema(tool_name)
    properties = full_schema.get("properties", {})
    schema = properties.get("outputSchema")
    if not schema:
        raise ValueError(f"No outputSchema in schema for {tool_name}")
    return schema
