#!/usr/bin/env python3
import argparse
import json
import os
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
FIXTURES = REPO_ROOT / "fixtures" / "http"


def write_json(obj):
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))


def load_fixture(key: str):
    p = FIXTURES / f"{key}.json"
    if not p.exists():
        return None
    with p.open("r", encoding="utf-8") as f:
        try:
            return json.load(f)
        except Exception as e:
            return {"status": 200, "body": f.read(), "fixture": True}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--op", default="help")
    parser.add_argument("--url")
    parser.add_argument("--headers")
    parser.add_argument("--body")
    parser.add_argument("--timeout-ms", type=int, default=0)
    parser.add_argument("--offline", action="store_true")
    parser.add_argument("--fixture-key")
    parser.add_argument("--trace-id")
    args = parser.parse_args()

    try:
        if args.op == "ping":
            write_json({"ok": True, "data": {"pong": True, "tool": "http.py"}})
            return 0
        if args.op in ("get", "post", "put", "delete", "head"):
            if args.offline:
                key = args.fixture_key or "default"
                data = load_fixture(key)
                if data is None:
                    write_json(
                        {
                            "ok": False,
                            "error": {
                                "code": "FIXTURE_NOT_FOUND",
                                "message": f"No fixture: {key}",
                            },
                        }
                    )
                    return 4
                # Normalize shape
                out = {
                    "status": data.get("status", 200),
                    "headers": data.get("headers", {}),
                    "body": data.get("body", data),
                    "fixture": True,
                }
                write_json({"ok": True, "data": out})
                return 0
            else:
                write_json(
                    {
                        "ok": False,
                        "error": {
                            "code": "OFFLINE_ONLY",
                            "message": "MVP soporta modo --offline con fixtures",
                        },
                    }
                )
                return 3
        else:
            write_json(
                {
                    "ok": False,
                    "error": {
                        "code": "USAGE",
                        "message": "Usa --op get|post|put|delete|head [--offline --fixture-key X]",
                    },
                }
            )
            return 1
    except Exception as e:
        write_json({"ok": False, "error": {"code": "EXCEPTION", "message": str(e)}})
        return 10


if __name__ == "__main__":
    sys.exit(main())
