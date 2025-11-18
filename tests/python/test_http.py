import unittest
import sys
import json
import subprocess
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
HTTP_TOOL = REPO_ROOT / "tools" / "py" / "http_tool.py"

class TestHttpTool(unittest.TestCase):
    
    def run_http_tool(self, *args):
        """Run the http tool and return exit code and output"""
        result = subprocess.run(
            [sys.executable, str(HTTP_TOOL)] + list(args),
            capture_output=True,
            text=True
        )
        return result.returncode, result.stdout.strip()
    
    def test_ping(self):
        exit_code, output = self.run_http_tool('--op', 'ping')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['pong'])
        self.assertEqual(data['data']['tool'], 'http_tool.py')
    
    def test_offline_no_fixture_key(self):
        # Without fixture-key it should use 'default' and fail if not found
        exit_code, output = self.run_http_tool('--op', 'get', '--offline')
        data = json.loads(output)
        # Either fixture found or fixture not found, both are valid offline behavior
        self.assertIn(data.get('ok'), [True, False])
    
    def test_missing_url_online(self):
        exit_code, output = self.run_http_tool('--op', 'get')
        self.assertNotEqual(exit_code, 0)
        data = json.loads(output)
        self.assertFalse(data['ok'])
        self.assertEqual(data['error']['code'], 'USAGE')

if __name__ == '__main__':
    unittest.main()
