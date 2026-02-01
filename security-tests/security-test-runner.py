#!/usr/bin/env python3
import os
import sys
import json
import time
import requests
import argparse
import glob
from concurrent.futures import ThreadPoolExecutor
from typing import Dict, List, Any

# Configuration
DEFAULT_API_URL = "http://localhost:8080/execute"
REPORT_FILE_JSON = "security-test-report.json"
REPORT_FILE_HTML = "security-test-report.html"

# Test Suites Mapping
TEST_SUITES = {
    '01-fork-bomb': {
        'name': 'Fork Bomb Attack',
        'expected_error': ['进程数量达到限制', 'fork', 'resource temporarily unavailable'],
        'criticality': 'HIGH'
    },
    '02-memory-bomb': {
        'name': 'Memory Exhaustion',
        'expected_error': ['内存使用过高', 'OOM Killed', 'signal: killed', 'memory limit'],
        'criticality': 'HIGH'
    },
    '03-cpu-exhaustion': {
        'name': 'CPU Exhaustion',
        'expected_error': ['执行超时', 'context deadline exceeded'],
        'criticality': 'MEDIUM'
    },
    '04-filesystem-attacks': {
        'name': 'Filesystem Isolation',
        'expected_error': ['Permission denied', 'read-only', 'no such file'],
        # Note: Some filesystem tests like read-passwd expect success but specific content
        'special_check': True,
        'criticality': 'HIGH'
    },
    '05-network-access': {
        'name': 'Network Isolation',
        'expected_error': ['Network is unreachable', 'Name or service not known', 'Temporary failure in name resolution', 'Connection refused'],
        'criticality': 'HIGH'
    },
    '06-info-leakage': {
        'name': 'Info Leakage',
        'expected_error': [], # These usually run successfully but we check output
        'special_check': True,
        'criticality': 'MEDIUM'
    }
}

LANGUAGE_MAP = {
    '.py': 'python',
    '.c': 'c',
    '.cpp': 'cpp',
    '.java': 'java',
    '.go': 'go',
    '.js': 'javascript'
}

class SecurityRunner:
    def __init__(self, api_url, output_format='html'):
        self.api_url = api_url
        self.base_url = api_url.rsplit('/', 1)[0] # e.g. http://localhost:8080/api
        self.output_format = output_format
        self.token = None
        self.results = []
        self.stats = {'total': 0, 'passed': 0, 'failed': 0, 'warning': 0}

    def authenticate(self):
        print("🔐 Authenticating...")
        user_data = {"username": "security_test", "password": "SafePassword123!"}
        
        # Try Register
        try:
            reg_resp = requests.post(f"{self.base_url}/register", json=user_data)
            if reg_resp.status_code == 200:
                print("   Created test user.")
            elif "User already exists" in reg_resp.text:
                print("   Test user exists.")
        except Exception as e:
            print(f"   Registration warning: {e}")

        # Login
        try:
            login_resp = requests.post(f"{self.base_url}/login", json=user_data)
            if login_resp.status_code == 200:
                self.token = login_resp.json().get('token')
                print("   Logged in successfully.")
            else:
                print(f"❌ Login failed: {login_resp.text}")
        except Exception as e:
            print(f"❌ Login error: {e}")

    def run_tests(self):
        self.authenticate()
        if not self.token:
            print("❌ Cannot run tests without authentication token")
            return

        print(f"🚀 Starting Security Test Suite against {self.api_url}")
        
        script_dir = os.path.dirname(os.path.abspath(__file__))
        
        # Helper to get files for an extension
        def get_files(ext):
            return glob.glob(os.path.join(script_dir, f"**/*{ext}"), recursive=True)

        test_files = sorted(get_files(".py") + 
                          get_files(".c") +
                          get_files(".cpp") +
                          get_files(".java") +
                          get_files(".go") +
                          get_files(".js"))
        
        # Filter out the runner script itself if picked up
        runner_script = os.path.abspath(__file__)
        test_files = [f for f in test_files if os.path.abspath(f) != runner_script and "debug" not in f]

        if not test_files:
            print("❌ No test files found in security-tests/ directory")
            return

        print(f"📝 Found {len(test_files)} test cases")
        
        for test_file in test_files:
            self.run_single_test(test_file)

        self.generate_report()
        self.print_summary()

    def run_single_test(self, file_path):
        self.stats['total'] += 1
        
        # Determine suite and language
        parts = file_path.split(os.sep)
        suite_dir = parts[-2] if len(parts) > 1 else "unknown"
        ext = os.path.splitext(file_path)[1]
        language = LANGUAGE_MAP.get(ext)
        
        if not language:
            print(f"⚠️ Skipping {file_path}: Unsupported language")
            return

        print(f"running: {file_path} ({language})...", end="", flush=True)

        # Read code
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                code = f.read()
        except Exception as e:
            self.record_result(file_path, suite_dir, language, False, f"Read error: {str(e)}")
            return

        # Execute
        start_time = time.time()
        try:
            headers = {'Authorization': f'Bearer {self.token}'}
            response = requests.post(self.api_url, json={
                "language": language,
                "code": code
            }, headers=headers, timeout=15) # Long timeout for exhaustion tests
            
            duration = time.time() - start_time
            
            try:
                result_data = response.json()
            except:
                print(f" ❌ INVALID JSON: {response.text[:50]}...")
                self.record_result(file_path, suite_dir, language, False, f"Invalid JSON response: {response.status_code}")
                return

            if response.status_code != 200:
                 # Check if the error is essentially the security block
                 error_msg = result_data.get('error', '')
                 success, message = self.verify_security_response(suite_dir, '', error_msg)
                 if success:
                     # It was a 400/500 but it was the EXPECTED security block
                     pass
                 else:
                     print(f" ❌ API ERROR {response.status_code}")
                     self.record_result(file_path, suite_dir, language, False, f"API Error {response.status_code}: {error_msg}")
                     return
            
            # Extract task info
            task_id = result_data.get('taskId')
            
            # Poll for result
            output = ""
            error_msg = ""
            status = "unknown"
            
            print(" ⏳", end="", flush=True)
            for _ in range(20): # Poll for up to 20 seconds
                time.sleep(1)
                try:
                    hist_resp = requests.get(f"{self.base_url}/submissions", headers=headers, timeout=5)
                    if hist_resp.status_code == 200:
                        submissions = hist_resp.json().get('submissions', [])
                        if submissions:
                            latest = submissions[0]
                            # Simple heuristic: assuming no parallel tests running
                            status = latest.get('status')
                            if status in ['Completed', 'Failed', 'Error']:
                                output = latest.get('output', '')
                                # Sometimes error is in output, sometimes separate? 
                                # Model seems to only have Output field. 
                                # So we treat Output as the Result.
                                if status != 'Completed':
                                    error_msg = output # Use output as error if status is failed
                                break
                            elif status == 'Pending' or status == 'Processing':
                                continue
                except:
                    pass
            else:
                 print(" ⚠️ TIMEOUT (Polling)")
                 self.record_result(file_path, suite_dir, language, False, "Polling Timeout")
                 return

            if status not in ['Completed', 'Failed', 'Error']:
                print(f" ⚠️ STUCK ({status})")
                self.record_result(file_path, suite_dir, language, False, f"Task Stuck in {status}")
                return
            
            # Verify Result
            success, message = self.verify_security_response(suite_dir, output, error_msg)

            
            # Special check for Java (output often mixed)
            if not success and language == 'java' and 'Error' in output:
                 success, message = self.verify_security_response(suite_dir, output, output)

            print(f" {'✅ PASS' if success else '❌ FAIL'} ({duration:.2f}s)")
            if not success:
                print(f"    Expected: {self.get_expected_for_suite(suite_dir)}")
                print(f"    Got: {error_msg if error_msg else output[:100].strip()}")

            self.record_result(file_path, suite_dir, language, success, message, output, error_msg, duration)

        except requests.exceptions.Timeout:
            print(" ⚠️ TIMEOUT (API)")
            self.record_result(file_path, suite_dir, language, False, "API Request Timed Out")
        except Exception as e:
            print(f" ❌ ERROR: {e}")
            self.record_result(file_path, suite_dir, language, False, f"Exception: {str(e)}")

    def get_expected_for_suite(self, suite_dir):
        config = TEST_SUITES.get(suite_dir)
        if config:
            return config.get('expected_error', 'N/A')
        return 'N/A'

    def verify_security_response(self, suite_dir, output, error_msg):
        config = TEST_SUITES.get(suite_dir)
        if not config:
            return True, "Unknown Suite"

        # Special logic for info leakage and filesystem read/write check scripts
        if config.get('special_check'):
            # These scripts print their own pass/fail status
            combined_out = (output + " " + error_msg).lower()
            if "security failure" in combined_out or "warning" in combined_out:
                return False, "Script detected security failure"
            if "security warning" in combined_out:
                return False, "Script detected potential leak"
            return True, "Script execution verification passed"

        # Standard error matching
        expected_errors = config.get('expected_error', [])
        
        # We expect a failure (error_msg should not be empty)
        # However, for our API, "Run Error" is returned in the 'output' or 'error' field depending on implementation
        # The Current Go implementation returns run errors in the 'error' field of the JSON or 'output' 
        
        combined_out = (output + " " + error_msg).lower()
        
        for expected in expected_errors:
            if expected.lower() in combined_out:
                return True, f"Blocked with expected error: {expected}"
        
        # If no expected error found, checking if it ran successfully (which is bad for bombs)
        if not error_msg and "process exited" not in output.lower():
             return False, "Malicious code executed successfully without being blocked"
             
        return False, f"Did not receive expected error. Got: {error_msg or output[:50]}"

    def record_result(self, file, suite, lang, passed, msg, out='', err='', dur=0):
        self.results.append({
            'file': file,
            'suite': suite,
            'language': lang,
            'passed': passed,
            'message': msg,
            'output': out,
            'error': err,
            'duration': dur,
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S')
        })
        if passed:
            self.stats['passed'] += 1
        else:
            self.stats['failed'] += 1

    def generate_report(self):
        with open(REPORT_FILE_JSON, 'w', encoding='utf-8') as f:
            json.dump({'stats': self.stats, 'results': self.results}, f, indent=2)
            
        if self.output_format == 'html':
            self.generate_html_report()

    def generate_html_report(self):
        # Determine the absolute path for the template
        script_dir = os.path.dirname(os.path.abspath(__file__))
        template_path = os.path.join(script_dir, "test-report-template.html")
        if not os.path.exists(template_path):
            template_path = os.path.abspath("test-report-template.html") # Fallback to cwd
        
        # Read template or use default
        try:
            with open(template_path, 'r', encoding='utf-8') as f:
                template = f.read()
        except:
            print("Warning: HTML template not found, using simple fallback.")
            template = "<html><body><h1>Security Test Report</h1><pre>{{RESULTS_JSON}}</pre></body></html>"

        # Build HTML content
        rows = ""
        for r in self.results:
            status_class = "pass" if r['passed'] else "fail"
            status_icon = "✅" if r['passed'] else "❌"
            rows += f"""
            <tr class="{status_class}">
                <td>{r['suite']}</td>
                <td>{os.path.basename(r['file'])}</td>
                <td>{r['language']}</td>
                <td class="status">{status_icon} {'PASS' if r['passed'] else 'FAIL'}</td>
                <td>{r['message']}</td>
                <td>{r['duration']:.2f}s</td>
            </tr>
            <tr class="details">
                <td colspan="6">
                    <details>
                        <summary>View Output</summary>
                        <strong>Output:</strong><br><pre>{r['output']}</pre>
                        <strong>Error:</strong><br><pre>{r['error']}</pre>
                    </details>
                </td>
            </tr>
            """
            
        html = template.replace('{{DATE}}', time.strftime('%Y-%m-%d %H:%M:%S'))
        html = template.replace('{{TOTAL}}', str(self.stats['total']))
        html = template.replace('{{PASSED}}', str(self.stats['passed']))
        html = template.replace('{{FAILED}}', str(self.stats['failed']))
        pass_rate = (self.stats['passed'] / self.stats['total'] * 100) if self.stats['total'] > 0 else 0
        html = template.replace('{{PASS_RATE}}', f"{pass_rate:.1f}%")
        html = template.replace('{{TABLE_ROWS}}', rows)
        
        with open(REPORT_FILE_HTML, 'w', encoding='utf-8') as f:
            f.write(html)
        print(f"\n📊 Report generated: {os.path.abspath(REPORT_FILE_HTML)}")

    def print_summary(self):
        print("\n" + "="*30)
        print("TEST SUMMARY")
        print("="*30)
        print(f"Total Tests: {self.stats['total']}")
        print(f"Passed:      {self.stats['passed']} ✅")
        print(f"Failed:      {self.stats['failed']} ❌")
        print("="*30)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Run Security Test Suite')
    parser.add_argument('--url', default=DEFAULT_API_URL, help='API Endpoint URL')
    args = parser.parse_args()
    
    runner = SecurityRunner(args.url)
    runner.run_tests()
