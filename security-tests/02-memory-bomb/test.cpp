// Memory Bomb Attack Test - C++ Language
// Expected: Memory limit reached (Memory=512MB) or OOM Killed
// Should return error: "内存使用过高" or "内存耗尽被系统终止(OOM Killed)"

#include <vector>
#include <cstring>

int main() {
    std::vector<char*> allocations;
    
    while(true) {
        try {
            // Allocate 100MB chunks
            char* ptr = new char[100 * 1024 * 1024];
            memset(ptr, 'A', 100 * 1024 * 1024);
            allocations.push_back(ptr);
        } catch (std::bad_alloc&) {
            // Continue trying even if allocation fails
        }
    }
    
    return 0;
}
