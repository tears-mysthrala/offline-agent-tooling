#!/usr/bin/env python3
import argparse
import json
import sqlite3
import sys
import time
from pathlib import Path

DB_PATH = Path(__file__).resolve().parents[2] / ".data" / "kv.sqlite"
DB_PATH.parent.mkdir(parents=True, exist_ok=True)

SCHEMA = """
CREATE TABLE IF NOT EXISTS kv (
  ns TEXT NOT NULL,
  key TEXT NOT NULL,
  value TEXT,
  expires_at INTEGER,
  PRIMARY KEY (ns, key)
);
"""


def write_json(obj):
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))


def get_conn():
    conn = sqlite3.connect(str(DB_PATH))
    return conn


def init_db(conn):
    conn.execute(SCHEMA)
    conn.commit()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--op", required=True)
    parser.add_argument("--ns", default="default")
    parser.add_argument("--key")
    parser.add_argument("--value")
    parser.add_argument("--ttl", type=int)
    args = parser.parse_args()

    conn = get_conn()
    init_db(conn)
    cur = conn.cursor()
    now = int(time.time())

    try:
        if args.op == "set":
            if args.key is None or args.value is None:
                write_json(
                    {
                        "ok": False,
                        "error": {
                            "code": "ARG_MISSING",
                            "message": "--key and --value required",
                        },
                    }
                )
                return 2
            expires = None
            if args.ttl:
                expires = now + args.ttl
            cur.execute(
                "REPLACE INTO kv (ns,key,value,expires_at) VALUES (?,?,?,?)",
                (args.ns, args.key, args.value, expires),
            )
            conn.commit()
            write_json({"ok": True, "data": {"ns": args.ns, "key": args.key}})
            return 0
        if args.op == "get":
            if args.key is None:
                write_json(
                    {
                        "ok": False,
                        "error": {"code": "ARG_MISSING", "message": "--key required"},
                    }
                )
                return 2
            cur.execute(
                "SELECT value,expires_at FROM kv WHERE ns=? AND key=?",
                (args.ns, args.key),
            )
            row = cur.fetchone()
            if not row:
                write_json({"ok": True, "data": {"found": False}})
                return 0
            value, expires = row
            if expires and expires < now:
                # expired
                cur.execute("DELETE FROM kv WHERE ns=? AND key=?", (args.ns, args.key))
                conn.commit()
                write_json({"ok": True, "data": {"found": False, "expired": True}})
                return 0
            write_json({"ok": True, "data": {"found": True, "value": value}})
            return 0
        if args.op == "del":
            if args.key is None:
                write_json(
                    {
                        "ok": False,
                        "error": {"code": "ARG_MISSING", "message": "--key required"},
                    }
                )
                return 2
            cur.execute("DELETE FROM kv WHERE ns=? AND key=?", (args.ns, args.key))
            conn.commit()
            write_json({"ok": True, "data": {"deleted": True}})
            return 0
        if args.op == "keys":
            cur.execute("SELECT key FROM kv WHERE ns=?", (args.ns,))
            rows = [r[0] for r in cur.fetchall()]
            write_json({"ok": True, "data": {"keys": rows}})
            return 0
        if args.op == "purge-expired":
            cur.execute(
                "DELETE FROM kv WHERE expires_at IS NOT NULL AND expires_at<?", (now,)
            )
            deleted = conn.total_changes
            conn.commit()
            write_json({"ok": True, "data": {"purged": deleted}})
            return 0
        write_json(
            {
                "ok": False,
                "error": {
                    "code": "USAGE",
                    "message": "op must be set|get|del|keys|purge-expired",
                },
            }
        )
        return 1
    finally:
        conn.close()


if __name__ == "__main__":
    sys.exit(main())
