# Fork Bomb Attack Test - Python
# Expected: Process limit reached (PidsLimit=50)
# Should return error: "进程数量达到限制(50)，可能触发了 fork 炸弹攻击"

import os

while True:
    os.fork()
