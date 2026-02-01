# Filesystem Read Attack - Python
# Expected: Read only container files or Permission denied
# Should NOT be able to read host files

try:
    with open('/etc/passwd', 'r') as f:
        content = f.read()
        print('File content length:', len(content))
        print('First 100 chars:', content[:100])
        
        # Check if it looks like a full linux passwd file
        if 'root:x:0:0' in content:
            print('Successfully read /etc/passwd')
        else:
            print('Read content, but does not match expected format')
            
except Exception as e:
    print(f'Error reading file: {e}')
