"""
Common validation and error utilities for tools.
Provides consistent error handling across all tools.
"""
import sys
import json


def write_json(obj):
    """Write JSON output to stdout."""
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()


def error_missing_arg(arg_name: str, example: str = None) -> dict:
    """
    Generate consistent error for missing required argument.
    
    Args:
        arg_name: Name of the missing argument
        example: Optional usage example
    
    Returns:
        Error dict with code, message, and suggestion
    """
    msg = f"Missing required argument: {arg_name}"
    suggestion = f"Use --{arg_name} <value>"
    if example:
        suggestion += f"\nExample: {example}"
    
    return {
        'ok': False,
        'error': {
            'code': 'ARG_MISSING',
            'message': msg,
            'suggestion': suggestion
        }
    }


def error_invalid_value(arg_name: str, value: any, reason: str, valid_values: list = None) -> dict:
    """
    Generate error for invalid argument value.
    
    Args:
        arg_name: Name of the argument
        value: The invalid value
        reason: Why it's invalid
        valid_values: List of valid values (optional)
    
    Returns:
        Error dict with code, message, and valid options
    """
    msg = f"Invalid value for --{arg_name}: {value}"
    details = {'reason': reason}
    if valid_values:
        details['valid_values'] = valid_values
    
    return {
        'ok': False,
        'error': {
            'code': 'INVALID_VALUE',
            'message': msg,
            'details': details
        }
    }


def error_invalid_op(op: str, valid_ops: list, tool_name: str) -> dict:
    """
    Generate error for invalid operation.
    
    Args:
        op: The invalid operation name
        valid_ops: List of valid operations
        tool_name: Name of the tool
    
    Returns:
        Error dict with valid operations and help suggestion
    """
    return {
        'ok': False,
        'error': {
            'code': 'INVALID_OPERATION',
            'message': f"Unknown operation: {op}",
            'valid_operations': valid_ops,
            'suggestion': f"Use: python {tool_name} --op help"
        }
    }


def validate_file_path(path: str, must_exist: bool = False) -> tuple:
    """
    Validate file path.
    
    Args:
        path: Path to validate
        must_exist: Whether file must already exist
    
    Returns:
        (is_valid, error_dict or None)
    """
    from pathlib import Path
    
    if not path or not path.strip():
        return (False, error_missing_arg('path', 'python tool.py --op operation --path /path/to/file'))
    
    try:
        p = Path(path)
        if must_exist and not p.exists():
            return (False, {
                'ok': False,
                'error': {
                    'code': 'NOT_FOUND',
                    'message': f'File or directory not found: {path}',
                    'suggestion': 'Check the path is correct and accessible'
                }
            })
        return (True, None)
    except Exception as e:
        return (False, {
            'ok': False,
            'error': {
                'code': 'INVALID_PATH',
                'message': f'Invalid path: {path}',
                'details': str(e)
            }
        })


def validate_json_string(json_str: str, arg_name: str) -> tuple:
    """
    Validate and parse JSON string.
    
    Args:
        json_str: JSON string to validate
        arg_name: Name of the argument (for error messages)
    
    Returns:
        (parsed_value or None, error_dict or None)
    """
    if not json_str or not json_str.strip():
        return (None, error_missing_arg(arg_name, f'--{arg_name} \'{{"key":"value"}}\''))
    
    try:
        parsed = json.loads(json_str)
        return (parsed, None)
    except json.JSONDecodeError as e:
        return (None, {
            'ok': False,
            'error': {
                'code': 'INVALID_JSON',
                'message': f'Invalid JSON in --{arg_name}',
                'details': str(e),
                'suggestion': 'Ensure JSON is properly formatted with quotes'
            }
        })
