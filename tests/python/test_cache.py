import unittest
import subprocess
import sys
import json
import time
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
CACHE_TOOL = REPO_ROOT / "tools" / "py" / "cache.py"


class TestCacheTool(unittest.TestCase):
    
    def run_cache_tool(self, *args):
        """Run the cache tool and return exit code and output"""
        result = subprocess.run(
            [sys.executable, str(CACHE_TOOL)] + list(args),
            capture_output=True,
            text=True
        )
        return result.returncode, result.stdout.strip()
    
    def test_ping(self):
        exit_code, output = self.run_cache_tool('--op', 'ping')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['pong'])
    
    def test_put_and_get(self):
        # Put a value
        exit_code, output = self.run_cache_tool(
            '--op', 'put',
            '--key', 'test_key',
            '--value', '"test_value"'
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['cached'])
        
        # Get the value
        exit_code, output = self.run_cache_tool('--op', 'get', '--key', 'test_key')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['found'])
        self.assertEqual(data['data']['value'], 'test_value')
    
    def test_put_json_value(self):
        json_value = json.dumps({'nested': 'object', 'count': 42})
        
        exit_code, output = self.run_cache_tool(
            '--op', 'put',
            '--key', 'json_key',
            '--value', json_value
        )
        self.assertEqual(exit_code, 0)
        
        exit_code, output = self.run_cache_tool('--op', 'get', '--key', 'json_key')
        data = json.loads(output)
        self.assertTrue(data['data']['found'])
        self.assertEqual(data['data']['value']['nested'], 'object')
        self.assertEqual(data['data']['value']['count'], 42)
    
    def test_ttl_expiration(self):
        # Put with 1 second TTL
        exit_code, output = self.run_cache_tool(
            '--op', 'put',
            '--key', 'ttl_key',
            '--value', '"expires_soon"',
            '--ttl-s', '1'
        )
        self.assertEqual(exit_code, 0)
        
        # Should be found immediately
        exit_code, output = self.run_cache_tool('--op', 'get', '--key', 'ttl_key')
        data = json.loads(output)
        self.assertTrue(data['data']['found'])
        
        # Wait for expiration
        time.sleep(1.5)
        
        # Should be expired and not found
        exit_code, output = self.run_cache_tool('--op', 'get', '--key', 'ttl_key')
        data = json.loads(output)
        self.assertFalse(data['data']['found'])
    
    def test_delete(self):
        # Put a value
        self.run_cache_tool('--op', 'put', '--key', 'delete_me', '--value', '"value"')
        
        # Delete it
        exit_code, output = self.run_cache_tool('--op', 'del', '--key', 'delete_me')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['deleted'])
        
        # Should not be found
        exit_code, output = self.run_cache_tool('--op', 'get', '--key', 'delete_me')
        data = json.loads(output)
        self.assertFalse(data['data']['found'])
    
    def test_stats(self):
        # Clear all first
        self.run_cache_tool('--op', 'clear-all')
        
        # Add some entries
        self.run_cache_tool('--op', 'put', '--key', 'k1', '--value', '"v1"')
        self.run_cache_tool('--op', 'put', '--key', 'k2', '--value', '"v2"')
        
        exit_code, output = self.run_cache_tool('--op', 'stats')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        stats = data['data']
        
        self.assertGreaterEqual(stats['total_entries'], 2)
        self.assertGreaterEqual(stats['valid_entries'], 2)
        self.assertGreater(stats['total_size_bytes'], 0)
    
    def test_clear_expired(self):
        # Clear all first
        self.run_cache_tool('--op', 'clear-all')
        
        # Add expired entry
        self.run_cache_tool('--op', 'put', '--key', 'exp', '--value', '"old"', '--ttl-s', '1')
        time.sleep(1.5)
        
        # Clear expired
        exit_code, output = self.run_cache_tool('--op', 'clear-expired')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertGreaterEqual(data['data']['removed'], 1)


if __name__ == '__main__':
    unittest.main()
