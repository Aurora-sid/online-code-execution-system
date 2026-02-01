// Memory Bomb Attack Test - Java
// Expected: Memory limit reached (Memory=512MB) or OOM Killed
// Should return error: "内存使用过高" or "内存耗尽被系统终止(OOM Killed)"

import java.util.ArrayList;
import java.util.List;

public class Main {
    public static void main(String[] args) {
        List<byte[]> allocations = new ArrayList<>();
        
        while (true) {
            try {
                // Allocate 100MB chunks
                byte[] chunk = new byte[100 * 1024 * 1024];
                // Fill the memory to ensure it's actually used
                for (int i = 0; i < chunk.length; i++) {
                    chunk[i] = (byte) 'A';
                }
                allocations.add(chunk);
            } catch (OutOfMemoryError e) {
                // Continue trying even on OOM
                System.err.println("OOM: " + e.getMessage());
            }
        }
    }
}
