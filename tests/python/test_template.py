import unittest
import subprocess
import sys
import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
TEMPLATE_TOOL = REPO_ROOT / "tools" / "py" / "template.py"


class TestTemplateTool(unittest.TestCase):
    
    def run_template_tool(self, *args):
        """Run the template tool and return exit code and output"""
        result = subprocess.run(
            [sys.executable, str(TEMPLATE_TOOL)] + list(args),
            capture_output=True,
            text=True
        )
        return result.returncode, result.stdout.strip()
    
    def test_ping(self):
        exit_code, output = self.run_template_tool('--op', 'ping')
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertTrue(data['data']['pong'])
    
    def test_simple_substitution(self):
        exit_code, output = self.run_template_tool(
            '--op', 'render',
            '--template', 'Hello ${name}!',
            '--vars', '{"name":"World"}'
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertTrue(data['ok'])
        self.assertEqual(data['data']['rendered'], 'Hello World!')
    
    def test_multiple_variables(self):
        template = 'User: ${user}, Age: ${age}, City: ${city}'
        vars_json = json.dumps({'user': 'Alice', 'age': '30', 'city': 'NYC'})
        
        exit_code, output = self.run_template_tool(
            '--op', 'render',
            '--template', template,
            '--vars', vars_json
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        self.assertEqual(data['data']['rendered'], 'User: Alice, Age: 30, City: NYC')
    
    def test_missing_variable_error(self):
        exit_code, output = self.run_template_tool(
            '--op', 'render',
            '--template', 'Hello ${name}!',
            '--vars', '{}'
        )
        self.assertNotEqual(exit_code, 0)
        data = json.loads(output)
        self.assertFalse(data['ok'])
        self.assertEqual(data['error']['code'], 'MISSING_VAR')
    
    def test_safe_substitution(self):
        exit_code, output = self.run_template_tool(
            '--op', 'render',
            '--template', 'Hello ${name}, you have ${count} messages',
            '--vars', '{"name":"Bob"}',
            '--safe'
        )
        self.assertEqual(exit_code, 0)
        data = json.loads(output)
        # Safe mode leaves missing vars as ${var}
        self.assertEqual(data['data']['rendered'], 'Hello Bob, you have ${count} messages')


if __name__ == '__main__':
    unittest.main()
