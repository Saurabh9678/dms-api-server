#!/usr/bin/env python3
"""Sync docs/api/*.postman_collection.json into postman/collections/DMS API/ YAML layout."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DOCS_API = ROOT / "docs" / "api"
POSTMAN_ROOT = ROOT / "postman" / "collections" / "DMS API"

MODULE_ORDER = {
    "auth": 1000,
    "user": 2000,
    "dashboard": 3000,
    "vehicle": 4000,
}

JSON_TO_MODULE = {
    "auth.postman_collection.json": "auth",
    "user.postman_collection.json": "user",
    "dashboard.postman_collection.json": "dashboard",
    "vehicle.postman_collection.json": "vehicle",
}


def yaml_quote(value: str) -> str:
    if not value:
        return '""'
    if any(c in value for c in ':"\'\n#{}[]&*!|>%@`'):
        escaped = value.replace("\\", "\\\\").replace('"', '\\"')
        return f'"{escaped}"'
    return value


def indent_block(text: str, spaces: int = 4) -> str:
    pad = " " * spaces
    return "\n".join(pad + line if line else "" for line in text.splitlines())


def build_url(url_obj: dict) -> str:
    raw = url_obj.get("raw", "")
    if "?" in raw:
        return raw
    query = url_obj.get("query") or []
    enabled = [q for q in query if not q.get("disabled")]
    if not enabled:
        return raw
    params = "&".join(f"{q['key']}={q['value']}" for q in enabled if q.get("key"))
    if not params:
        return raw
    return f"{raw}?{params}"


def headers_dict(header_list: list[dict] | None) -> dict[str, str]:
    result: dict[str, str] = {}
    for header in header_list or []:
        key = header.get("key")
        if key:
            result[key] = header.get("value", "")
    return result


def body_content(request: dict) -> str | None:
    body = request.get("body") or {}
    raw = body.get("raw")
    if raw:
        return raw
    return None


def render_headers(headers: dict[str, str], indent: int = 2) -> str:
    if not headers:
        return ""
    pad = " " * indent
    lines = [f"{pad}{key}: {headers[key]}" for key in headers]
    return "\n".join(lines)


def render_request_section(request: dict, indent: int = 2) -> str:
    pad = " " * indent
    url = build_url(request.get("url", {}))
    method = request.get("method", "GET")
    headers = headers_dict(request.get("header"))
    body = body_content(request)

    parts = [
        f"{pad}url: {yaml_quote(url)}",
        f"{pad}method: {method}",
    ]
    if headers:
        parts.append(f"{pad}headers:")
        parts.append(render_headers(headers, indent + 2))
    if body is not None:
        parts.append(f"{pad}body:")
        parts.append(f"{pad}  type: text")
        parts.append(f"{pad}  content: |-")
        parts.append(indent_block(body, indent + 4))
    return "\n".join(parts)


def render_request_file(item: dict, order: int) -> str:
    request = item["request"]
    description = request.get("description", "")
    lines = [
        "$kind: http-request",
        f"description: {yaml_quote(description)}",
        f"url: {yaml_quote(build_url(request.get('url', {})))}",
        f"method: {request.get('method', 'GET')}",
    ]
    headers = headers_dict(request.get("header"))
    if headers:
        lines.append("headers:")
        for key, value in headers.items():
            lines.append(f"  {key}: {value}")
    body = body_content(request)
    if body is not None:
        lines.append("body:")
        lines.append("  type: text")
        lines.append("  content: |-")
        lines.extend("    " + line for line in body.splitlines())
    lines.append(f"examples: ./.resources/{item['name']}.resources/examples")
    lines.append(f"order: {order}")
    return "\n".join(lines) + "\n"


def render_example_file(
    example: dict,
    fallback_request: dict,
    order: int,
) -> str:
    req = example.get("originalRequest") or fallback_request
    status_code = example.get("code", 200)
    status_text = example.get("status", "OK")
    body = example.get("body", "")

    lines = [
        "$kind: http-example",
        "request:",
    ]
    request_block = render_request_section(req, indent=2)
    lines.append(request_block)
    lines.append("response:")
    lines.append(f"  statusCode: {status_code}")
    lines.append(f"  statusText: {status_text}")
    lines.append("  body:")
    lines.append("    type: text")
    lines.append("    content: |-")
    lines.extend("      " + line for line in body.splitlines())
    lines.append(f"order: {order}")
    return "\n".join(lines) + "\n"


def sanitize_filename(name: str) -> str:
    return name.replace("/", "-")


def extract_items(payload: dict) -> list[dict]:
    items = payload.get("item", [])
    if not items:
        return []
    if "request" in items[0]:
        return items
    nested: list[dict] = []
    for folder in items:
        nested.extend(folder.get("item", []))
    return nested


def sync_module(module: str, json_path: Path) -> int:
    payload = json.loads(json_path.read_text(encoding="utf-8"))
    items = extract_items(payload)
    module_dir = POSTMAN_ROOT / module
    resources_dir = module_dir / ".resources"
    module_dir.mkdir(parents=True, exist_ok=True)
    resources_dir.mkdir(parents=True, exist_ok=True)

    (resources_dir / "definition.yaml").write_text(
        f"$kind: collection\norder: {MODULE_ORDER[module]}\n",
        encoding="utf-8",
    )

    count = 0
    for idx, item in enumerate(items, start=1):
        request_name = item["name"]
        request_order = idx * 1000
        request_path = module_dir / f"{request_name}.request.yaml"
        request_path.write_text(
            render_request_file(item, request_order),
            encoding="utf-8",
        )

        examples_dir = resources_dir / f"{request_name}.resources" / "examples"
        examples_dir.mkdir(parents=True, exist_ok=True)

        for ex_idx, example in enumerate(item.get("response") or [], start=1):
            example_name = sanitize_filename(example["name"])
            example_path = examples_dir / f"{example_name}.example.yaml"
            example_path.write_text(
                render_example_file(
                    example,
                    item["request"],
                    ex_idx * 1000,
                ),
                encoding="utf-8",
            )
        count += 1
    return count


def main() -> int:
    total = 0
    for json_name, module in JSON_TO_MODULE.items():
        json_path = DOCS_API / json_name
        if not json_path.exists():
            print(f"skip missing: {json_path}", file=sys.stderr)
            continue
        count = sync_module(module, json_path)
        print(f"synced {module}: {count} requests")
        total += count

    description = (
        "DMS API collection. Modules: auth, user, dashboard, vehicle. "
        "Kept in sync with docs/api/*.postman_collection.json."
    )
    (POSTMAN_ROOT / ".resources" / "definition.yaml").write_text(
        f"$kind: collection\n"
        f"description: {description}\n"
        f"variables:\n"
        f"  base_url: http://localhost:8080\n"
        f'  access_token: ""\n',
        encoding="utf-8",
    )
    print(f"done: {total} requests synced")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
