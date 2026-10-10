"""A fake runtime for the tunnel spike, bound to 127.0.0.1 only (as a runtime behind Bridge would
be): /v1/models, a streamed chat reply with a few chunks, and /bulk?mb=N for throughput.
Usage: fake_runtime.py PORT"""
import json
import sys
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

CHUNK = b"x" * (1 << 20)


class H(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def do_GET(self):
        u = urlparse(self.path)
        if u.path == "/bulk":
            mb = int(parse_qs(u.query).get("mb", ["100"])[0])
            self.send_response(200)
            self.send_header("Content-Length", str(mb << 20))
            self.end_headers()
            for _ in range(mb):
                self.wfile.write(CHUNK)
            return
        body = json.dumps({"data": [{"id": "qwen3:4b", "context_length": 8192}]}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_POST(self):
        self.rfile.read(int(self.headers.get("Content-Length") or 0))
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Connection", "close")
        self.end_headers()
        for i in range(5):
            self.wfile.write(f'data: {{"choices":[{{"delta":{{"content":"t{i}"}}}}]}}\n\n'.encode())
            self.wfile.flush()
            time.sleep(0.05)
        self.wfile.write(b"data: [DONE]\n\n")
        self.close_connection = True

    def log_message(self, *a):
        pass


ThreadingHTTPServer(("127.0.0.1", int(sys.argv[1])), H).serve_forever()
