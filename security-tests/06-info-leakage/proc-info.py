# Process Info Leakage Test - Python
# Expected: Should only see container processes (few PIDs)
# Should NOT see host processes

import os

print('--- Process List (via /proc) ---')
try:
    pids = [pid for pid in os.listdir('/proc') if pid.isdigit()]
    print(f'Total visible processes: {len(pids)}')
    print(f'PIDs: {pids[:20]} ...' if len(pids) > 20 else f'PIDs: {pids}')
    
    if len(pids) > 20:
        print('SECURITY WARNING: Seeing too many processes, possible PID namespace leak?')
    else:
        print('Process isolation seems OK (Low PID count).')
        
    # Check specific host processes (common ones)
    # This is a heuristic check
    if '1' in pids:
       # PID 1 in container should be the startup script or shell, not systemd/init of host
       pass
       
except Exception as e:
    print(f'Error reading /proc: {e}')
