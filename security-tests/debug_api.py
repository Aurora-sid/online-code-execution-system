import requests

url = "http://localhost:8080/api/run"
data = {
    "language": "python",
    "code": "print('hello')"
}

try:
    print(f"Sending request to {url}...")
    resp = requests.post(url, json=data)
    print(f"Status Code: {resp.status_code}")
    print("--- Raw Response Content ---")
    print(resp.content)
    print("--- End Raw Response ---")
    print("Attempting to parse JSON:")
    print(resp.json())
except Exception as e:
    print(f"Error: {e}")
