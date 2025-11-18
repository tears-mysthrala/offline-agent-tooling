import unittest
import subprocess
import sys
import json
import os
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
CONFIG_TOOL = REPO_ROOT / "tools" / "py" / "config.py"
FIXTURES = REPO_ROOT / "fixtures" / "config"


class TestConfigTool(unittest.TestCase):
    
    def run_config_tool(self, *args, **env_vars):
        """Run the config tool and return exit code and output"""
        env = os.environ.copy()
        env.update(env_vars)
        
        result = subprocess.run(
            [sys.executable, str(CONFIG_TOOL)] + list(args),
            capture_output=True,
            text=True,
            env=env
        )
        return result.returncode, result.stdout.strip()
    
    def test_ping(self):
        exit_code, output = self.run_config_tool('--op', 'ping')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['pong'])
    
    def test_load_env_file(self):
        env_file = FIXTURES / "test.env"
        exit_code, output = self.run_config_tool('--op', 'load', '--paths', str(env_file))
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        config = data['data']['config']
        
        self.assertEqual(config['DATABASE_URL'], 'postgresql://localhost/testdb')
        self.assertEqual(config['DEBUG'], 'true')
        self.assertEqual(config['PORT'], '8080')
        self.assertEqual(config['LOG_LEVEL'], 'info')
        
        # Check that API_KEY is redacted
        redacted = data['data']['redacted_config']
        self.assertEqual(redacted['API_KEY'], '***REDACTED***')
    
    def test_load_json_file(self):
        json_file = FIXTURES / "test.json"
        exit_code, output = self.run_config_tool('--op', 'load', '--paths', str(json_file))
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        config = data['data']['config']
        
        self.assertEqual(config['app_name'], 'test-app')
        self.assertEqual(config['version'], '1.0.0')
        self.assertTrue(config['features']['enable_cache'])
    
    def test_load_env_vars_with_prefix(self):
        exit_code, output = self.run_config_tool(
            '--op', 'load',
            '--env-prefix', 'TEST_',
            TEST_VAR1='value1',
            TEST_VAR2='value2',
            OTHER_VAR='ignored'
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        config = data['data']['config']
        
        # Should have TEST_ vars with prefix stripped
        self.assertEqual(config.get('VAR1'), 'value1')
        self.assertEqual(config.get('VAR2'), 'value2')
        # Should not have OTHER_VAR (doesn't match prefix)
        self.assertNotIn('OTHER_VAR', config)
    
    def test_merge_with_overrides(self):
        env_file = FIXTURES / "test.env"
        overrides = json.dumps({'PORT': '9090', 'NEW_KEY': 'new_value'})
        
        exit_code, output = self.run_config_tool(
            '--op', 'load',
            '--paths', str(env_file),
            '--overrides', overrides
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        config = data['data']['config']
        
        # Override should take precedence
        self.assertEqual(config['PORT'], '9090')
        # Original values should still be there
        self.assertEqual(config['DEBUG'], 'true')
        # New key should be added
        self.assertEqual(config['NEW_KEY'], 'new_value')
    
    def test_multiple_files(self):
        env_file = FIXTURES / "test.env"
        json_file = FIXTURES / "test.json"
        
        exit_code, output = self.run_config_tool(
            '--op', 'load',
            '--paths', f'{env_file},{json_file}'
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        config = data['data']['config']
        
        # Should have keys from both files
        self.assertIn('DATABASE_URL', config)  # from .env
        self.assertIn('app_name', config)      # from .json


if __name__ == '__main__':
    unittest.main()
