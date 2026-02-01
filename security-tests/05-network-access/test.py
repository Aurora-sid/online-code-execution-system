# Network Access Test - Python
# Expected: Network is unreachable
# Should return error containing "Network is unreachable" or fail to connect

import socket
import sys

try:
    print('Attempting to connect to google.com:80...')
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(3) # 3 seconds timeout
    s.connect(('google.com', 80))
    print('SECURITY FAILURE: Successfully connected to google.com')
    s.close()
except Exception as e:
    print(f'Network blocked (Expected): {e}')

try:
    print('Attempting to resolve domain example.com...')
    socket.gethostbyname('example.com')
    print('DNS resolution successful (Might be allowed depending on config)')
except Exception as e:
    print(f'DNS resolution failed: {e}')
