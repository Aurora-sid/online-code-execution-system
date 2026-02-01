// Memory Bomb Attack Test - C Language
// Expected: Memory limit reached (Memory=512MB) or OOM Killed
// Should return error: "内存使用过高" or "内存耗尽被系统终止(OOM Killed)"

#include <stdlib.h>
#include <string.h>

int main() {
    while(1) {
        // Try to allocate 1GB of memory
        void *ptr = malloc(1024 * 1024 * 1024);
        if (ptr != NULL) {
            // Actually use the memory to trigger OOM
            memset(ptr, 0, 1024 * 1024 * 1024);
        }
    }
    return 0;
}
