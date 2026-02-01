# Environment Variable Leakage Test - Python
# Expected: Should NOT see host environment variables (like API keys, DB passwords)
# Should only see container specific vars

import os

print('--- Environment Variables ---')
sensitive_keywords = ['KEY', 'SECRET', 'PASSWORD', 'TOKEN', 'AUTH', 'DB_', 'MYSQL', 'REDIS']
found_sensitive = False

for key, value in os.environ.items():
    print(f'{key}={value}')
    
    # Check for suspicious sensitive info
    for keyword in sensitive_keywords:
        if keyword in key.upper():
            print(f'WARNING: Potential sensitive variable found: {key}')
            found_sensitive = True

if found_sensitive:
    print('SECURITY WARNING: Sensitive environment variables might be exposed!')
else:
    print('No obvious sensitive environment variables found.')
