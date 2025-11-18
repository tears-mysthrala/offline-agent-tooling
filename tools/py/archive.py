#!/usr/bin/env python3
"""
Archive Tool - Compress/decompress using zipfile.
Follows TODO.md specification for Archive Tool (#15).
"""
import argparse
import json
import sys
import zipfile
from pathlib import Path


def write_json(obj):
    """Write JSON output to stdout."""
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()


def archive_zip(source: Path, dest: Path) -> bool:
    """Create a zip archive from source path."""
    try:
        with zipfile.ZipFile(dest, 'w', zipfile.ZIP_DEFLATED) as zf:
            if source.is_file():
                zf.write(source, source.name)
            elif source.is_dir():
                for file_path in source.rglob('*'):
                    if file_path.is_file():
                        zf.write(file_path, file_path.relative_to(source.parent))
        return True
    except Exception:
        return False


def archive_unzip(source: Path, dest: Path) -> bool:
    """Extract a zip archive to dest path."""
    try:
        dest.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(source, 'r') as zf:
            zf.extractall(dest)
        return True
    except Exception:
        return False


def archive_list(source: Path) -> list:
    """List contents of a zip archive."""
    try:
        with zipfile.ZipFile(source, 'r') as zf:
            files = []
            for info in zf.infolist():
                files.append({
                    'filename': info.filename,
                    'size': info.file_size,
                    'compressed_size': info.compress_size,
                    'is_dir': info.is_dir()
                })
            return files
    except Exception:
        return []


def main():
    parser = argparse.ArgumentParser(description='Archive Tool - Zip/Unzip operations')
    parser.add_argument('--op', required=True, help='Operation: zip|unzip|list')
    parser.add_argument('--source', help='Source file or directory')
    parser.add_argument('--dest', help='Destination file or directory')
    parser.add_argument('--compact', action='store_true', help='Minimal output (for LLMs)')
    parser.add_argument('--trace-id', help='Trace ID for logging')
    
    args = parser.parse_args()
    
    try:
        if args.op == 'help' or args.op == '--help':
            help_text = """
Archive Tool - Zip/unzip operations using stdlib zipfile

USAGE:
  python archive.py --op <operation> [options]

OPERATIONS:
  ping                      Health check
  zip                       Create zip archive
  unzip                     Extract zip archive
  list                      List archive contents

OPTIONS:
  --source PATH             Source file or directory
  --dest PATH               Destination file or directory
  --trace-id ID             Trace ID for logging

EXAMPLES:
  # Create zip archive from directory
  python archive.py --op zip --source mydir --dest archive.zip
  
  # Extract zip archive
  python archive.py --op unzip --source archive.zip --dest output/
  
  # List archive contents
  python archive.py --op list --source archive.zip
"""
            print(help_text)
            return 0
        
        if args.op == 'version' or args.op == '--version':
            write_json({'ok': True, 'data': {'version': '1.0.0', 'tool': 'archive.py'}})
            return 0
        
        if args.op == 'ping':
            write_json({'ok': True, 'data': {'pong': True, 'tool': 'archive.py'}})
            return 0
        
        if args.op == 'zip':
            if not args.source or not args.dest:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--source and --dest required'
                    }
                })
                return 2
            
            source = Path(args.source).resolve()
            dest = Path(args.dest).resolve()
            
            if not source.exists():
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'NOT_FOUND',
                        'message': f'Source not found: {source}'
                    }
                })
                return 3
            
            success = archive_zip(source, dest)
            
            if success:
                if args.compact:
                    write_json({'ok': True, 'data': str(dest)})
                else:
                    write_json({
                        'ok': True,
                        'data': {
                            'source': str(source),
                            'archive': str(dest),
                            'size': dest.stat().st_size if dest.exists() else 0
                        }
                    })
                return 0
            else:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARCHIVE_ERROR',
                        'message': 'Failed to create archive'
                    }
                })
                return 5
        
        if args.op == 'unzip':
            if not args.source or not args.dest:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--source and --dest required'
                    }
                })
                return 2
            
            source = Path(args.source).resolve()
            dest = Path(args.dest).resolve()
            
            if not source.exists():
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'NOT_FOUND',
                        'message': f'Archive not found: {source}'
                    }
                })
                return 3
            
            success = archive_unzip(source, dest)
            
            if success:
                write_json({
                    'ok': True,
                    'data': {
                        'archive': str(source),
                        'extracted_to': str(dest)
                    }
                })
                return 0
            else:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'EXTRACT_ERROR',
                        'message': 'Failed to extract archive'
                    }
                })
                return 5
        
        if args.op == 'list':
            if not args.source:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--source required'
                    }
                })
                return 2
            
            source = Path(args.source).resolve()
            
            if not source.exists():
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'NOT_FOUND',
                        'message': f'Archive not found: {source}'
                    }
                })
                return 3
            
            files = archive_list(source)
            
            if files is not None:
                write_json({
                    'ok': True,
                    'data': {
                        'archive': str(source),
                        'files': files,
                        'count': len(files)
                    }
                })
                return 0
            else:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'LIST_ERROR',
                        'message': 'Failed to list archive contents'
                    }
                })
                return 5
        
        write_json({
            'ok': False,
            'error': {
                'code': 'USAGE',
                'message': 'Use --op zip|unzip|list --source <path> [--dest <path>]'
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
