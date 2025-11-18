#!/usr/bin/env python3
"""
Template Tool - Simple variable substitution using string.Template.
Follows TODO.md specification for Template Tool (#12).
"""
import argparse
import json
import sys
from string import Template


def write_json(obj):
    """Write JSON output to stdout."""
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()


def main():
    parser = argparse.ArgumentParser(description='Template Tool - Variable substitution')
    parser.add_argument('--op', default='render', help='Operation: render')
    parser.add_argument('--template', help='Template string with ${variable} syntax')
    parser.add_argument('--vars', help='JSON object with variable values')
    parser.add_argument('--safe', action='store_true', help='Use safe_substitute (missing keys -> keep ${key})')
    parser.add_argument('--trace-id', help='Trace ID for logging')
    
    args = parser.parse_args()
    
    try:
        if args.op == 'ping':
            write_json({'ok': True, 'data': {'pong': True, 'tool': 'template.py'}})
            return 0
        
        if args.op == 'render':
            if not args.template:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--template required'
                    }
                })
                return 2
            
            # Parse variables
            variables = {}
            if args.vars:
                try:
                    variables = json.loads(args.vars)
                except json.JSONDecodeError as e:
                    write_json({
                        'ok': False,
                        'error': {
                            'code': 'INVALID_VARS',
                            'message': f'Invalid JSON in --vars: {e}'
                        }
                    })
                    return 2
            
            # Create template
            template = Template(args.template)
            
            # Render
            try:
                if args.safe:
                    rendered = template.safe_substitute(variables)
                else:
                    rendered = template.substitute(variables)
                
                write_json({
                    'ok': True,
                    'data': {
                        'rendered': rendered,
                        'vars_used': list(variables.keys())
                    }
                })
                return 0
                
            except KeyError as e:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'MISSING_VAR',
                        'message': f'Missing required variable: {e}'
                    }
                })
                return 4
        
        write_json({
            'ok': False,
            'error': {
                'code': 'USAGE',
                'message': 'Use --op render --template "Hello ${name}" --vars \'{"name":"World"}\''
            }
        })
        return 1
        
    except Exception as e:
        write_json({
            'ok': False,
            'error': {
                'code': 'EXCEPTION',
                'message': str(e)
            }
        })
        return 10


if __name__ == '__main__':
    sys.exit(main())
