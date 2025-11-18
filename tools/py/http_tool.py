#!/usr/bin/env python3
import argparse
import hashlib
import json
import sys
import urllib.request
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
FIXTURES = REPO_ROOT / "fixtures" / "http"
CACHE_DIR = REPO_ROOT / ".cache" / "http"


def write_json(obj):
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))


def load_fixture(key: str):
    p = FIXTURES / f"{key}.json"
    if not p.exists():
        return None
    with p.open("r", encoding="utf-8") as f:
        try:
            return json.load(f)
        except Exception:
            f.seek(0)
            return {"status": 200, "body": f.read(), "fixture": True}


def cache_get(url: str):
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    h = hashlib.sha256(url.encode("utf-8")).hexdigest()
    p = CACHE_DIR / f"{h}.json"
    if p.exists():
        with p.open("r", encoding="utf-8") as f:
            try:
                return json.load(f)
            except Exception:
                return None
    return None


def cache_put(url: str, data):
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    h = hashlib.sha256(url.encode("utf-8")).hexdigest()
    p = CACHE_DIR / f"{h}.json"
    with p.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False)


def http_fetch(url: str, timeout_ms: int = 0):
    req = urllib.request.Request(url, headers={})
    timeout = (timeout_ms / 1000.0) if timeout_ms and timeout_ms > 0 else None
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        body = resp.read()
        headers = dict(resp.getheaders())
        return {
            "status": resp.getcode(),
            "headers": headers,
            "body": body.decode("utf-8", errors="replace"),
        }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--op", default="help")
    parser.add_argument("--url")
    parser.add_argument("--headers")
    parser.add_argument("--body")
    parser.add_argument("--timeout-ms", type=int, default=0)
    parser.add_argument("--offline", action="store_true")
    parser.add_argument("--fixture-key")
    parser.add_argument("--cache", action="store_true")
    parser.add_argument("--compact", action="store_true", help="Minimal output")
    parser.add_argument("--trace-id")
    args = parser.parse_args()

    try:
        if args.op == "ping":
            write_json({"ok": True, "data": {"pong": True, "tool": "http_tool.py"}})
            return 0

        if args.op in ("get", "post", "put", "delete", "head"):
            # Offline via fixtures
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
                out = {
                    "status": data.get("status", 200),
                    "headers": data.get("headers", {}),
                    "body": data.get("body", data),
                    "fixture": True,
                }
                write_json({"ok": True, "data": out})
                return 0

            # Cache lookup
            if args.url and args.cache:
                cached = cache_get(args.url)
                if cached is not None:
                    write_json({"ok": True, "data": cached, "cached": True})
                    return 0

            # Online fetch (requires network)
            if args.url:
                try:
                    fetched = http_fetch(args.url, timeout_ms=args.timeout_ms)
                    if args.cache:
                        cache_put(args.url, fetched)
                    write_json({"ok": True, "data": fetched})
                    return 0
                except Exception as e:
                    write_json(
                        {
                            "ok": False,
                            "error": {"code": "FETCH_ERROR", "message": str(e)},
                        }
                    )
                    return 6

            write_json(
                {
                    "ok": False,
                    "error": {
                        "code": "USAGE",
                        "message": "--url requerido para modo online o usar --offline",
                    },
                }
            )
            return 2

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
