# Memory Bomb Attack Test - Python
# Expected: Memory limit reached (Memory=512MB) or OOM Killed
# Should return error: "内存使用过高" or "内存耗尽被系统终止(OOM Killed)"

data = []
while True:
    # Allocate 100MB each iteration
    data.append('A' * (10 ** 8))
