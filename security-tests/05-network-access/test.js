// Network Access Test - Node.js
// Expected: Network unreachable

const http = require('http');

console.log('Attempting HTTP request to http://example.com...');

const req = http.get('http://example.com', (res) => {
    console.log('SECURITY FAILURE: Got response: ' + res.statusCode);
    res.resume();
}).on('error', (e) => {
    console.log('Network blocked (Expected): ' + e.message);
});
