# Disk Space Bomb - Python
# Expected: Container storage limit reached

import os

try:
    with open('bigfile.dat', 'wb') as f:
        # Try to write 1GB
        for i in range(1024):
            f.write(b'A' * 1024 * 1024) # 1MB chunks
    print('Successfully wrote 1GB file')
    
    file_size = os.path.getsize('bigfile.dat')
    print(f'File size: {file_size} bytes')
    
except Exception as e:
    print(f'Write failed: {e}')
