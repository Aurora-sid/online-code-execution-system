# Filesystem Write Attack - Python
# Expected: Permission denied when writing to system directories

try:
    with open('/etc/malicious_file', 'w') as f:
        f.write('Attemping to overwrite system configuration')
    print('SECURITY FAILURE: Successfully wrote to /etc/')
except Exception as e:
    print(f'Write blocked (Expected): {e}')

try:
    with open('/usr/bin/malicious', 'w') as f:
        f.write('Attemping to inject binary')
    print('SECURITY FAILURE: Successfully wrote to /usr/bin/')
except Exception as e:
    print(f'Write blocked (Expected): {e}')
