// Fork Bomb Attack Test - Java
// Expected: Process limit reached (PidsLimit=50)
// Should return error: "进程数量达到限制(50)，可能触发了 fork 炸弹攻击"

import java.io.IOException;

public class Main {
    public static void main(String[] args) {
        while (true) {
            try {
                // Create new process using ProcessBuilder
                new ProcessBuilder("sh", "-c", "sleep 10").start();
            } catch (IOException e) {
                // Process creation failed, likely hit limit
                System.err.println("Process creation failed: " + e.getMessage());
            }
        }
    }
}
