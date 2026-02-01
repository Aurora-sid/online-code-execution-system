import requests
import time

BASE_URL = "http://localhost:8080/api"
AUTH_DATA = {"username": "debug_user", "password": "Password123!"}

def debug_execution():
    # 1. Login
    print("1. Logging in...")
    try:
        requests.post(f"{BASE_URL}/register", json=AUTH_DATA) # Ensure user exists
        resp = requests.post(f"{BASE_URL}/login", json=AUTH_DATA)
        if resp.status_code != 200:
            print(f"Login failed: {resp.text}")
            return
        token = resp.json()['token']
        print("   Login successful.")
    except Exception as e:
        print(f"Auth error: {e}")
        return

    headers = {'Authorization': f'Bearer {token}'}

    # 2. Test Normal Execution
    print("\n2. Testing Normal Execution (Python)...")
    payload = {
        "language": "python",
        "code": "print('Hello Security')"
    }
    try:
        resp = requests.post(f"{BASE_URL}/run", json=payload, headers=headers)
        print(f"   Status: {resp.status_code}")
        print(f"   Response: {resp.text}")
    except Exception as e:
        print(f"   Request error: {e}")

    # 3. Test Malicious Execution (Timeout)
    print("\n3. Testing Timeout (Python)...")
    payload = {
        "language": "python",
        "code": "import time\ntime.sleep(2)\nprint('Done')"
    }
    start = time.time()
    try:
        resp = requests.post(f"{BASE_URL}/run", json=payload, headers=headers)
        duration = time.time() - start
        print(f"   Status: {resp.status_code}")
        print(f"   Duration: {duration:.2f}s")
        print(f"   Response: {resp.text}")
    except Exception as e:
        print(f"   Request error: {e}")

if __name__ == "__main__":
    debug_execution()
