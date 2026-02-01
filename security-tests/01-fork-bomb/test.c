// Fork Bomb Attack Test - C Language
// Expected: Process limit reached (PidsLimit=50)
// Should return error: "进程数量达到限制(50)，可能触发了 fork 炸弹攻击"

#include <unistd.h>

int main() {
    while(1) {
        fork();
    }
    return 0;
}
