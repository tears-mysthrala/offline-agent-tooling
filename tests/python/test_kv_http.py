import unittest
import sys
import subprocess
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PY = sys.executable


def run_script(argv):
    # argv: list for the runpy wrapper like ['kv.py','--op','get',...]
    cmd = [
        PY,
        "-c",
        "import runpy,sys; sys.argv=%r; runpy.run_path('tools/py/' + sys.argv[0], run_name='__main__')"
        % argv,
    ]
    proc = subprocess.run(cmd, capture_output=True, text=True, shell=False)
    out = proc.stdout.strip()
    if not out:
        # sometimes printed to stderr
        out = proc.stderr.strip()
    # Expect JSON on a single line
    try:
        return json.loads(out)
    except Exception:
        raise AssertionError(
            f"Invalid JSON from script. exit={proc.returncode} stdout={proc.stdout!r} stderr={proc.stderr!r}"
        )


class KVHTTPTests(unittest.TestCase):
    def test_kv_set_get_del_keys(self):
        # set
        r = run_script(
            ["kv.py", "--op", "set", "--ns", "unittest", "--key", "k1", "--value", "v1"]
        )
        self.assertTrue(r.get("ok"))
        # get
        r = run_script(["kv.py", "--op", "get", "--ns", "unittest", "--key", "k1"])
        self.assertTrue(r.get("ok"))
        self.assertTrue(r["data"].get("found"))
        self.assertEqual(r["data"].get("value"), "v1")
        # keys
        r = run_script(["kv.py", "--op", "keys", "--ns", "unittest"])
        self.assertTrue(r.get("ok"))
        self.assertIn("k1", r["data"].get("keys", []))
        # del
        r = run_script(["kv.py", "--op", "del", "--ns", "unittest", "--key", "k1"])
        self.assertTrue(r.get("ok"))
        # get again
        r = run_script(["kv.py", "--op", "get", "--ns", "unittest", "--key", "k1"])
        self.assertTrue(r.get("ok"))
        self.assertFalse(r["data"].get("found", False))

    def test_http_offline_fixture(self):
        # uses fixture example-users.json
        r = run_script(
            [
                "http_tool.py",
                "--op",
                "get",
                "--offline",
                "--fixture-key",
                "example-users",
            ]
        )
        self.assertTrue(r.get("ok"))
        data = r["data"]
        self.assertIn("status", data)
        self.assertEqual(data.get("status"), 200)
        self.assertTrue(isinstance(data.get("body"), list))


if __name__ == "__main__":
    unittest.main()
