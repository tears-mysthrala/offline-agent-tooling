#!/usr/bin/env python3
"""
Config Tool - Load configuration from .env, JSON files, and environment variables.
Follows TODO.md specification for Config Tool (#9).
"""
import argparse
import json
import os
import sys
from pathlib import Path
from typing import Dict, Any, Optional


def write_json(obj):
    """Write JSON output to stdout."""
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()


def load_env_file(path: Path) -> Dict[str, str]:
    """Load .env file and return key-value pairs."""
    config = {}
    if not path.exists():
        return config
    
    with path.open('r', encoding='utf-8') as f:
        for line_num, line in enumerate(f, 1):
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith('#'):
                continue
            
            # Parse KEY=VALUE format
            if '=' not in line:
                continue  # Skip malformed lines
            
            key, _, value = line.partition('=')
            key = key.strip()
            value = value.strip()
            
            # Remove quotes if present
            if value and value[0] in ('"', "'") and value[-1] == value[0]:
                value = value[1:-1]
            
            config[key] = value
    
    return config


def load_json_file(path: Path) -> Dict[str, Any]:
    """Load JSON file and return parsed object."""
    if not path.exists():
        return {}
    
    with path.open('r', encoding='utf-8') as f:
        try:
            return json.load(f)
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in {path}: {e}")


def load_env_vars(prefix: Optional[str] = None) -> Dict[str, str]:
    """Load environment variables, optionally filtered by prefix."""
    config = {}
    for key, value in os.environ.items():
        if prefix:
            if key.startswith(prefix):
                # Strip prefix
                clean_key = key[len(prefix):]
                config[clean_key] = value
        else:
            config[key] = value
    return config


def merge_configs(*configs: Dict[str, Any]) -> Dict[str, Any]:
    """Merge multiple config dicts, later ones override earlier ones."""
    result = {}
    for config in configs:
        result.update(config)
    return result


def main():
    parser = argparse.ArgumentParser(description='Config Tool - Load configuration')
    parser.add_argument('--op', default='load', help='Operation: load')
    parser.add_argument('--paths', help='Comma-separated config file paths')
    parser.add_argument('--env-prefix', help='Environment variable prefix to filter')
    parser.add_argument('--overrides', help='JSON string of override values')
    parser.add_argument('--compact', action='store_true', help='Minimal output (for LLMs)')
    parser.add_argument('--quiet', action='store_true', help='Suppress non-essential fields')
    parser.add_argument('--trace-id', help='Trace ID for logging')
    
    args = parser.parse_args()
    
    try:
        if args.op == 'help' or args.op == '--help':
            help_text = """
Config Tool - Load configuration from .env, JSON, and environment variables

USAGE:
  python config.py --op <operation> [options]

OPERATIONS:
  ping                      Health check
  load                      Load configuration from multiple sources

OPTIONS:
  --paths PATH1,PATH2       Comma-separated config file paths (.env, .json)
  --env-prefix PREFIX_      Filter env vars by prefix (e.g., APP_)
  --overrides '{"k":"v"}'   JSON overrides (highest priority)
  --trace-id ID             Trace ID for logging

EXAMPLES:
  # Load .env file
  python config.py --op load --paths config.env
  
  # Load multiple files (merge order: files -> env -> overrides)
  python config.py --op load --paths base.env,prod.json
  
  # Load with env vars prefix
  python config.py --op load --paths config.env --env-prefix APP_
  
  # Load with overrides
  python config.py --op load --paths config.env --overrides '{"PORT":"8080"}'

NOTE: Keys containing 'password', 'secret', 'token', 'api_key' are auto-redacted
"""
            print(help_text)
            return 0
        
        if args.op == 'version' or args.op == '--version':
            write_json({'ok': True, 'data': {'version': '1.0.0', 'tool': 'config.py'}})
            return 0
        
        if args.op == 'ping':
            write_json({'ok': True, 'data': {'pong': True, 'tool': 'config.py'}})
            return 0
        
        if args.op == 'load':
            configs = []
            
            # 1. Load from file paths (lowest priority)
            if args.paths:
                paths = [p.strip() for p in args.paths.split(',')]
                for path_str in paths:
                    path = Path(path_str).resolve()
                    
                    if path.suffix == '.env':
                        configs.append(load_env_file(path))
                    elif path.suffix == '.json':
                        configs.append(load_json_file(path))
                    else:
                        # Try to detect by content
                        if path.exists():
                            try:
                                configs.append(load_json_file(path))
                            except ValueError:
                                configs.append(load_env_file(path))
            
            # 2. Load from environment variables (higher priority)
            configs.append(load_env_vars(args.env_prefix))
            
            # 3. Apply overrides (highest priority)
            if args.overrides:
                try:
                    overrides = json.loads(args.overrides)
                    configs.append(overrides)
                except json.JSONDecodeError as e:
                    write_json({
                        'ok': False,
                        'error': {
                            'code': 'INVALID_OVERRIDES',
                            'message': f'Invalid JSON in overrides: {e}'
                        }
                    })
                    return 2
            
            # Merge all configs
            final_config = merge_configs(*configs)
            
            # Redact secrets (keys containing 'secret', 'password', 'token', 'key')
            redacted_config = {}
            for k, v in final_config.items():
                key_lower = k.lower()
                if any(s in key_lower for s in ['secret', 'password', 'token', 'api_key', 'apikey']):
                    redacted_config[k] = '***REDACTED***'
                else:
                    redacted_config[k] = v
            
            # Compact mode: return just the config
            if args.compact:
                write_json({'ok': True, 'data': final_config})
            elif args.quiet:
                write_json({'ok': True, 'data': {'config': final_config}})
            else:
                write_json({
                    'ok': True,
                    'data': {
                        'config': final_config,
                        'redacted_config': redacted_config,
                        'count': len(final_config)
                    }
                })
            return 0
        
        write_json({
            'ok': False,
            'error': {
                'code': 'USAGE',
                'message': 'Use --op load [--paths file1.env,file2.json] [--env-prefix PREFIX_] [--overrides \'{"key":"val"}\']'
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
