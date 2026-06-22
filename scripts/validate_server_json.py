#!/usr/bin/env python3
"""Validate icuvisor MCP Registry metadata without publishing anything."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path
from urllib.parse import urlparse

DESCRIPTION_MAX = 100
NAME_RE = re.compile(r"^[a-zA-Z0-9.-]+/[a-zA-Z0-9._-]+$")
SHA256_RE = re.compile(r"^[a-f0-9]{64}$")
RANGE_TOKENS = ("^", "~", ">", "<", "*", "x", "X")
TRANSPORTS_WITH_URL = {"streamable-http", "sse"}
VALID_TRANSPORTS = {"stdio", *TRANSPORTS_WITH_URL}


def main() -> int:
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("server.json")
    errors = validate(path)
    if errors:
        for error in errors:
            print(f"server.json: {error}", file=sys.stderr)
        return 1
    print(f"server.json: ok ({path})")
    return 0


def validate(path: Path) -> list[str]:
    errors: list[str] = []
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        return [f"{path} does not exist"]
    except json.JSONDecodeError as exc:
        return [f"invalid JSON at line {exc.lineno}, column {exc.colno}: {exc.msg}"]

    if not isinstance(data, dict):
        return ["top-level value must be an object"]

    require_string(data, "$schema", errors)
    require_string(data, "name", errors)
    require_string(data, "description", errors)
    require_string(data, "version", errors)

    name = data.get("name")
    if isinstance(name, str) and not NAME_RE.fullmatch(name):
        errors.append("name must match reverse-DNS namespace/server format, e.g. io.github.owner/name")

    description = data.get("description")
    if isinstance(description, str) and len(description) > DESCRIPTION_MAX:
        errors.append(f"description is {len(description)} characters; registry limit is {DESCRIPTION_MAX}")

    version = data.get("version")
    if isinstance(version, str):
        validate_version("version", version, errors)

    for field in ("$schema", "websiteUrl"):
        value = data.get(field)
        if isinstance(value, str) and field.endswith("Url") and not is_url(value):
            errors.append(f"{field} must be an absolute URL")

    repository = data.get("repository")
    if repository is not None:
        if not isinstance(repository, dict):
            errors.append("repository must be an object")
        else:
            require_string(repository, "url", errors, "repository")
            require_string(repository, "source", errors, "repository")
            url = repository.get("url")
            if isinstance(url, str) and not is_url(url):
                errors.append("repository.url must be an absolute URL")

    packages = data.get("packages")
    remotes = data.get("remotes")
    if not packages and not remotes:
        errors.append("at least one package or remote entry is required")
    if packages is not None:
        if not isinstance(packages, list):
            errors.append("packages must be an array")
        else:
            for index, package in enumerate(packages):
                validate_package(package, index, errors)
    if remotes is not None:
        if not isinstance(remotes, list):
            errors.append("remotes must be an array")
        else:
            for index, remote in enumerate(remotes):
                validate_transport(remote, f"remotes[{index}].transport", errors)

    return errors


def validate_package(package: object, index: int, errors: list[str]) -> None:
    prefix = f"packages[{index}]"
    if not isinstance(package, dict):
        errors.append(f"{prefix} must be an object")
        return

    for field in ("registryType", "identifier", "version"):
        require_string(package, field, errors, prefix)
    validate_transport(package.get("transport"), f"{prefix}.transport", errors)

    version = package.get("version")
    if isinstance(version, str):
        validate_version(f"{prefix}.version", version, errors)

    registry_base_url = package.get("registryBaseUrl")
    if registry_base_url is not None:
        if not isinstance(registry_base_url, str):
            errors.append(f"{prefix}.registryBaseUrl must be a string")
        elif not is_url(registry_base_url):
            errors.append(f"{prefix}.registryBaseUrl must be an absolute URL")

    registry_type = package.get("registryType")
    identifier = package.get("identifier")
    if registry_type == "mcpb":
        file_sha = package.get("fileSha256")
        if not isinstance(file_sha, str) or not SHA256_RE.fullmatch(file_sha):
            errors.append(f"{prefix}.fileSha256 is required for mcpb packages and must be lowercase SHA-256")
        if isinstance(identifier, str) and not is_url(identifier):
            errors.append(f"{prefix}.identifier must be a release asset URL for mcpb packages")

    for field in ("runtimeArguments", "packageArguments", "environmentVariables"):
        value = package.get(field)
        if value is not None and not isinstance(value, list):
            errors.append(f"{prefix}.{field} must be an array")


def validate_transport(transport: object, prefix: str, errors: list[str]) -> None:
    if not isinstance(transport, dict):
        errors.append(f"{prefix} must be an object")
        return
    transport_type = transport.get("type")
    if not isinstance(transport_type, str):
        errors.append(f"{prefix}.type is required")
        return
    if transport_type not in VALID_TRANSPORTS:
        errors.append(f"{prefix}.type must be one of: {', '.join(sorted(VALID_TRANSPORTS))}")
    if transport_type in TRANSPORTS_WITH_URL:
        url = transport.get("url")
        if not isinstance(url, str) or not is_url(url):
            errors.append(f"{prefix}.url must be an absolute URL for {transport_type}")


def validate_version(field: str, version: str, errors: list[str]) -> None:
    if not version:
        errors.append(f"{field} must not be empty")
    if version == "latest" or any(token in version for token in RANGE_TOKENS):
        errors.append(f"{field} must be a specific version, not a range or latest")


def require_string(data: dict[str, object], field: str, errors: list[str], prefix: str = "") -> None:
    value = data.get(field)
    label = f"{prefix}.{field}" if prefix else field
    if not isinstance(value, str) or not value:
        errors.append(f"{label} is required and must be a non-empty string")


def is_url(value: str) -> bool:
    parsed = urlparse(value)
    return parsed.scheme in {"http", "https"} and bool(parsed.netloc)


if __name__ == "__main__":
    raise SystemExit(main())
