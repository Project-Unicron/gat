#!/usr/bin/env python3
import os
from pathlib import Path
import sys
from typing import Generator, Dict, Optional, List
import re

EXCLUDE_PATTERNS = {
    'node_modules',
    '.git',
    'dist',
    'build',
    '__pycache__',
    '.pytest_cache',
    'venv',
    'env',
    'backups'   # Add this
}

BINARY_EXTENSIONS = {
    # Archives
    '.zip', '.gz', '.tar', '.rar', '.7z',
    # Databases
    '.db', '.db.bak', '.sqlite', '.sqlite3',
    # Images
    '.jpg', '.jpeg', '.png', '.gif',
    # Compiled
    '.pyc', '.pyo', '.pyd',
    # Libraries
    '.so', '.dll', '.dylib',
    # Executables
    '.exe', '.bin',
    # Backups and logs
    '.bak', '.log'
}

def find_files(directory: str) -> Generator[str, None, None]:
    for root, dirs, files in os.walk(directory):
        # Skip excluded directories
        dirs[:] = [d for d in dirs if d not in EXCLUDE_PATTERNS]
        
        for file in files:
            if Path(file).suffix.lower() not in BINARY_EXTENSIONS:
                yield os.path.join(root, file)

def sanitize_search_term(term: str) -> str:
    """Convert raw input into safe search pattern"""
    return re.escape(term)

def find_matches(content: str, query: str) -> List[int]:
    """Find all occurrences of query in content"""
    pattern = sanitize_search_term(query)
    return [m.start() for m in re.finditer(pattern, content)]

def format_snippet(content: str, index: int, context: int = 50) -> str:
    """Format a snippet with consistent context"""
    start = max(0, index - context)
    end = min(len(content), index + context)
    return f"...{content[start:end]}..."

def search_file(path: str, query: str) -> Optional[Dict]:
    try:
        with open(path, 'r', encoding='utf-8') as f:
            content = f.read()
            matches = find_matches(content, query)
            
            if matches:
                return {
                    'path': os.path.relpath(path),
                    'snippets': [format_snippet(content, index) for index in matches]
                }
    except Exception as e:
        print(f"Error processing {path}: {e}")
        return None

def main():
    if len(sys.argv) != 2:
        print("Usage: python worm.py '<search_term>'")
        print("Example: python worm.py 'with open(path, \"r\") as f:'")
        return

    query = sys.argv[1]
    print(f"\nSearching for: {query}\n")

    results_found = False
    for file in find_files('.'):
        result = search_file(file, query)
        if result:
            results_found = True
            print(f"\nFound in: {result['path']}")
            for snippet in result['snippets']:
                print(f"Context: {snippet}")
            print("-" * 80)

    if not results_found:
        print("No matches found.")

if __name__ == '__main__':
    main()