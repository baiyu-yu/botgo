# -*- coding: utf-8 -*-
import os
import re
import sys

COMMON_SUFFIX_MAP = {
    '鎷?': '拉',
    '鎷夊彇': '拉取',
    '鐨?': '的',
    '褰?': '当',
    '璇?': '误',
    '涓?': '为',
    '浜?': '了',
    '鍜?': '和',
    '鎴?': '或',
    '鏈?': '有',
    '鏃?': '无',
    '鏄?': '是',
    '鍙?': '可',
    '鎺?': '接',
    '鐢?': '用',
    '鍝?': '哈',
    '浠?': '件',
    '鎭?': '息',
    '璁?': '设',
    '绗?': '第',
    '绛?': '等',
    '琛ㄦ€佷簨': '表态事',
    '鐘舵€佸彉鏇翠簨': '状态变更事',
}

def is_cjk_or_ascii(char: str) -> bool:
    code = ord(char)
    if code <= 0x7F:
        return True
    if 0x2000 <= code <= 0x20BF:
        return True
    if 0x3040 <= code <= 0x30FF:
        return True
    if 0x3100 <= code <= 0x31BF:
        return True
    if 0x4E00 <= code <= 0x9FFF or 0x3400 <= code <= 0x4DBF:
        return True
    if 0x3000 <= code <= 0x303F:
        return True
    if 0xFF00 <= code <= 0xFFEF:
        return True
    if 0xF900 <= code <= 0xFAFF or 0x20000 <= code <= 0x2E7FF:
        return True
    return False

def recover_chunk(chunk: str) -> str:
    if not chunk:
        return ""
    try:
        b = chunk.encode('gb18030')
        decoded = b.decode('utf-8')
        if decoded != chunk and all(is_cjk_or_ascii(c) for c in decoded):
            if any(0x4E00 <= ord(c) <= 0x9FFF for c in decoded):
                return decoded
    except Exception:
        pass
    
    for i in range(len(chunk) - 1, 0, -1):
        prefix = chunk[:i]
        suffix = chunk[i:]
        try:
            b = prefix.encode('gb18030')
            decoded = b.decode('utf-8')
            if decoded != prefix and all(is_cjk_or_ascii(c) for c in decoded):
                if any(0x4E00 <= ord(c) <= 0x9FFF for c in decoded):
                    return decoded + recover_chunk(suffix)
        except Exception:
            pass
            
    return chunk

def recover_mojibake(text: str) -> tuple[str, bool]:
    pattern = re.compile(r'[^\x00-\x7F]+')
    
    def replace_func(match):
        chunk = match.group(0)
        return recover_chunk(chunk)
    
    new_text = pattern.sub(replace_func, text)
    
    for garbled, fixed in COMMON_SUFFIX_MAP.items():
        if garbled in new_text:
            new_text = new_text.replace(garbled, fixed)
            
    return new_text, (new_text != text)

def process_file(filepath: str, dry_run=True):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except UnicodeDecodeError:
        try:
            with open(filepath, 'r', encoding='gb18030') as f:
                content = f.read()
            print(f"[RECODE] File is GBK/GB18030, converting to UTF-8: {filepath}")
            if not dry_run:
                with open(filepath, 'w', encoding='utf-8') as f:
                    f.write(content)
            return True
        except Exception as e:
            print(f"[ERROR] Could not read file {filepath} as UTF-8 or GB18030: {e}")
            return False
    except Exception as e:
        print(f"[ERROR] Could not read file {filepath}: {e}")
        return False
        
    new_content, changed = recover_mojibake(content)
    if changed:
        print(f"[FIX] Found mojibake in UTF-8 file, fixing: {filepath}")
        if not dry_run:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(new_content)
        return True
    return False

def main():
    dry_run = '--write' not in sys.argv
    root_dir = "."
    files_to_process = []
    
    allowed_extensions = ('.go', '.md', '.json', '.txt', '.yml', '.yaml', '.properties', '.ini', '.sh', '.bat')
    
    for dirpath, dirnames, filenames in os.walk(root_dir):
        if '.git' in dirnames:
            dirnames.remove('.git')
        
        for filename in filenames:
            ext = os.path.splitext(filename)[1].lower()
            if ext in allowed_extensions or filename in ('LICENSE', 'Dockerfile', 'Makefile'):
                filepath = os.path.join(dirpath, filename)
                files_to_process.append(filepath)
                
    print(f"Scanning {len(files_to_process)} files (dry_run={dry_run})...")
    modified_count = 0
    for filepath in files_to_process:
        if any(x in filepath for x in ['fix_mojibake.py', 'detect.ps1']):
            continue
        if process_file(filepath, dry_run=dry_run):
            modified_count += 1
            
    if dry_run:
        print(f"\nDry-run completed. Found {modified_count} files to modify. Run with --write to apply changes.")
    else:
        print(f"\nCompleted. Successfully modified {modified_count} files.")

if __name__ == '__main__':
    main()
