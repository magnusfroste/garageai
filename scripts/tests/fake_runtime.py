"""A fake OpenAI-compatible runtime for the connect script tests: serves /v1/models, and
Ollama's /api/ps with the model loaded at a small (default) context window.

Usage: fake_runtime.py <bind address> <port>
"""
import json
import socketserver
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/api/ps":
            body = json.dumps({"models": [{"name": "qwen3:4b", "model": "qwen3:4b", "context_length": 4096}]}).encode()
        else:
            body = json.dumps({"data": [{"id": "qwen3:4b"}, {"id": "nomic-embed-text:latest"}]}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


class Server(HTTPServer):
    def server_bind(self):
        # HTTPServer.server_bind resolves the host name, which takes many seconds on macOS.
        socketserver.TCPServer.server_bind(self)
        self.server_name, self.server_port = self.server_address[:2]


Server((sys.argv[1], int(sys.argv[2])), Handler).serve_forever()
