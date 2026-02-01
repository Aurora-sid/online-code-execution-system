// Network Access Test - Java
// Expected: Network unreachable

import java.net.HttpURLConnection;
import java.net.URL;
import java.io.IOException;

public class Main {
    public static void main(String[] args) {
        try {
            System.out.println("Attempting HTTP request to http://example.com...");
            URL url = new URL("http://example.com");
            HttpURLConnection con = (HttpURLConnection) url.openConnection();
            con.setConnectTimeout(3000);
            con.setRequestMethod("GET");
            int status = con.getResponseCode();
            System.out.println("SECURITY FAILURE: Response Code: " + status);
        } catch (IOException e) {
            System.out.println("Network blocked (Expected): " + e.getMessage());
        }
    }
}
