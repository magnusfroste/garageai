import sys, json
from http.server import BaseHTTPRequestHandler, HTTPServer
class H(BaseHTTPRequestHandler):
    def do_GET(self):
        b = json.dumps({"data":[{"id":"qwen3:4b"},{"id":"nomic-embed-text:latest"}]}).encode()
        self.send_response(200); self.send_header("Content-Type","application/json"); self.end_headers(); self.wfile.write(b)
    def log_message(self,*a): pass
HTTPServer((sys.argv[1], int(sys.argv[2])), H).serve_forever()
