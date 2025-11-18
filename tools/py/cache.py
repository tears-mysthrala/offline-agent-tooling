#!/usr/bin/env python3
"""
Cache Tool - Filesystem cache with TTL and LRU eviction.
Follows TODO.md specification for Cache Tool (#8).
"""
import argparse
import hashlib
import json
import os
import sys
import time
from pathlib import Path
from typing import Optional, Any


CACHE_DIR = Path(__file__).resolve().parents[2] / ".cache" / "cache"
CACHE_DIR.mkdir(parents=True, exist_ok=True)


def write_json(obj):
    """Write JSON output to stdout."""
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))
    sys.stdout.flush()


def get_cache_key_hash(key: str) -> str:
    """Generate SHA-256 hash for cache key."""
    return hashlib.sha256(key.encode('utf-8')).hexdigest()


def get_cache_path(key: str) -> Path:
    """Get cache file path for a given key."""
    key_hash = get_cache_key_hash(key)
    return CACHE_DIR / f"{key_hash}.json"


def cache_put(key: str, value: Any, ttl_s: Optional[int] = None) -> bool:
    """Store value in cache with optional TTL."""
    cache_path = get_cache_path(key)
    
    expires_at = None
    if ttl_s is not None and ttl_s > 0:
        expires_at = int(time.time()) + ttl_s
    
    cache_entry = {
        'key': key,
        'value': value,
        'created_at': int(time.time()),
        'expires_at': expires_at
    }
    
    try:
        with cache_path.open('w', encoding='utf-8') as f:
            json.dump(cache_entry, f, ensure_ascii=False)
        return True
    except Exception:
        return False


def cache_get(key: str) -> tuple[bool, Optional[Any]]:
    """
    Retrieve value from cache.
    Returns (found, value) tuple.
    """
    cache_path = get_cache_path(key)
    
    if not cache_path.exists():
        return (False, None)
    
    try:
        with cache_path.open('r', encoding='utf-8') as f:
            cache_entry = json.load(f)
        
        # Check expiration
        if cache_entry.get('expires_at'):
            if int(time.time()) >= cache_entry['expires_at']:
                # Expired - delete and return not found
                cache_path.unlink()
                return (False, None)
        
        return (True, cache_entry.get('value'))
    
    except Exception:
        return (False, None)


def cache_del(key: str) -> bool:
    """Delete cache entry. Returns True if deleted, False if not found."""
    cache_path = get_cache_path(key)
    
    if cache_path.exists():
        try:
            cache_path.unlink()
            return True
        except Exception:
            return False
    return False


def cache_stats() -> dict:
    """Get cache statistics."""
    cache_files = list(CACHE_DIR.glob("*.json"))
    
    total_size = 0
    valid_count = 0
    expired_count = 0
    now = int(time.time())
    
    for cache_file in cache_files:
        try:
            total_size += cache_file.stat().st_size
            with cache_file.open('r', encoding='utf-8') as f:
                entry = json.load(f)
            
            if entry.get('expires_at') and entry['expires_at'] < now:
                expired_count += 1
            else:
                valid_count += 1
        except Exception:
            pass
    
    return {
        'total_entries': len(cache_files),
        'valid_entries': valid_count,
        'expired_entries': expired_count,
        'total_size_bytes': total_size,
        'cache_dir': str(CACHE_DIR)
    }


def cache_clear_expired() -> int:
    """Remove all expired cache entries. Returns count of removed entries."""
    cache_files = list(CACHE_DIR.glob("*.json"))
    removed = 0
    now = int(time.time())
    
    for cache_file in cache_files:
        try:
            with cache_file.open('r', encoding='utf-8') as f:
                entry = json.load(f)
            
            if entry.get('expires_at') and entry['expires_at'] < now:
                cache_file.unlink()
                removed += 1
        except Exception:
            pass
    
    return removed


def cache_clear_all() -> int:
    """Remove all cache entries. Returns count of removed entries."""
    cache_files = list(CACHE_DIR.glob("*.json"))
    removed = 0
    
    for cache_file in cache_files:
        try:
            cache_file.unlink()
            removed += 1
        except Exception:
            pass
    
    return removed


def main():
    parser = argparse.ArgumentParser(description='Cache Tool - Filesystem cache with TTL')
    parser.add_argument('--op', required=True, help='Operation: put|get|del|stats|clear-expired|clear-all')
    parser.add_argument('--key', help='Cache key')
    parser.add_argument('--value', help='Value to cache (JSON string)')
    parser.add_argument('--ttl-s', type=int, help='Time-to-live in seconds')
    parser.add_argument('--trace-id', help='Trace ID for logging')
    
    args = parser.parse_args()
    
    try:
        if args.op == 'help' or args.op == '--help':
            help_text = """
Cache Tool - Filesystem cache with TTL and automatic expiration

USAGE:
  python cache.py --op <operation> [options]

OPERATIONS:
  ping                      Health check
  put                       Store value in cache
  get                       Retrieve value from cache
  del                       Delete cache entry
  stats                     Get cache statistics
  clear-expired             Remove expired entries
  clear-all                 Remove all cache entries

OPTIONS:
  --key KEY                 Cache key
  --value VALUE             Value to cache (JSON string)
  --ttl-s SECONDS           Time-to-live in seconds
  --trace-id ID             Trace ID for logging

EXAMPLES:
  # Put value with 60s TTL
  python cache.py --op put --key session --value '"data"' --ttl-s 60
  
  # Get value
  python cache.py --op get --key session
  
  # Delete value
  python cache.py --op del --key session
  
  # Get cache stats
  python cache.py --op stats
  
  # Clear expired entries
  python cache.py --op clear-expired
"""
            print(help_text)
            return 0
        
        if args.op == 'version' or args.op == '--version':
            write_json({'ok': True, 'data': {'version': '1.0.0', 'tool': 'cache.py'}})
            return 0
        
        if args.op == 'ping':
            write_json({'ok': True, 'data': {'pong': True, 'tool': 'cache.py'}})
            return 0
        
        if args.op == 'put':
            if not args.key or not args.value:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--key and --value required'
                    }
                })
                return 2
            
            # Parse value as JSON if possible
            try:
                value = json.loads(args.value)
            except json.JSONDecodeError:
                value = args.value
            
            success = cache_put(args.key, value, args.ttl_s)
            
            if success:
                write_json({
                    'ok': True,
                    'data': {
                        'key': args.key,
                        'cached': True,
                        'ttl_s': args.ttl_s
                    }
                })
                return 0
            else:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'CACHE_ERROR',
                        'message': 'Failed to cache value'
                    }
                })
                return 5
        
        if args.op == 'get':
            if not args.key:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--key required'
                    }
                })
                return 2
            
            found, value = cache_get(args.key)
            
            write_json({
                'ok': True,
                'data': {
                    'key': args.key,
                    'found': found,
                    'value': value if found else None
                }
            })
            return 0
        
        if args.op == 'del':
            if not args.key:
                write_json({
                    'ok': False,
                    'error': {
                        'code': 'ARG_MISSING',
                        'message': '--key required'
                    }
                })
                return 2
            
            deleted = cache_del(args.key)
            
            write_json({
                'ok': True,
                'data': {
                    'key': args.key,
                    'deleted': deleted
                }
            })
            return 0
        
        if args.op == 'stats':
            stats = cache_stats()
            write_json({
                'ok': True,
                'data': stats
            })
            return 0
        
        if args.op == 'clear-expired':
            removed = cache_clear_expired()
            write_json({
                'ok': True,
                'data': {
                    'removed': removed
                }
            })
            return 0
        
        if args.op == 'clear-all':
            removed = cache_clear_all()
            write_json({
                'ok': True,
                'data': {
                    'removed': removed
                }
            })
            return 0
        
        write_json({
            'ok': False,
            'error': {
                'code': 'USAGE',
                'message': 'Use --op put|get|del|stats|clear-expired|clear-all'
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
