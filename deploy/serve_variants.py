#!/usr/bin/env python3
"""
agentUsage Multi-Variant UI Runner

Serves 3 completely different glanceable UI designs that display live data:
  - Port 8080: Design 1 (Split: Glanceable Submenu + Deep Inspector)
  - Port 8081: Design 2 (Matrix: Dense Roster Matrix HUD)
  - Port 8082: Design 3 (Bento: Viewport-Fitted Bento Grid / Modular Glance Tiles)

Forwards all /api/* and /healthz calls to the live backend server (default port 8085).
Standard library only.
"""

import argparse
import http.server
import os
import signal
import socket
import sys
import threading
import time
import urllib.error
import urllib.request

DEFAULT_VARIANTS = {
    8080: "split",
    8081: "matrix",
    8082: "bento",
}

VARIANT_DESCRIPTIONS = {
    "split": "Refined Glanceable Submenu + Deep Inspector",
    "matrix": "Dense Roster Matrix HUD",
    "bento": "Viewport-Fitted Bento Grid / Modular Glance Tiles",
}


def is_port_in_use(port, host="127.0.0.1"):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.settimeout(0.5)
        return s.connect_ex((host, port)) == 0


def create_variant_handler(variant_id, ui_dir, backend_host, backend_port):
    class VariantHTTPHandler(http.server.BaseHTTPRequestHandler):
        server_version = "agentUsage-VariantServer/1.0"

        def log_message(self, format, *args):
            # Keep console clean; log errors only or minimal access info
            if args and len(args) >= 2:
                try:
                    status = int(args[1])
                    if status >= 400:
                        sys.stderr.write(f"[{variant_id} :{self.server.server_port}] {format % args}\n")
                except (ValueError, TypeError):
                    pass

        def do_HEAD(self):
            self.handle_request(send_body=False)

        def do_GET(self):
            self.handle_request(send_body=True)

        def do_POST(self):
            self.handle_request(send_body=True)

        def handle_request(self, send_body=True):
            clean_path = self.path.split("?", 1)[0]

            # 1. Forward API and healthz requests to backend
            if clean_path.startswith("/api/") or clean_path == "/healthz":
                self.proxy_to_backend(send_body=send_body)
                return

            # 2. Serve static UI assets
            self.serve_ui_asset(clean_path, send_body=send_body)

        def proxy_to_backend(self, send_body=True):
            backend_url = f"http://{backend_host}:{backend_port}{self.path}"
            body_bytes = None
            if self.command == "POST":
                content_length = int(self.headers.get("Content-Length", 0))
                if content_length > 0:
                    body_bytes = self.rfile.read(content_length)

            req = urllib.request.Request(backend_url, data=body_bytes, method=self.command)

            # Forward relevant client headers
            for header in ["Authorization", "Content-Type", "Accept", "User-Agent", "Sec-Fetch-Site"]:
                val = self.headers.get(header)
                if val:
                    req.add_header(header, val)

            # Forward Origin/Host appropriately (allow loopback origin checks)
            req.add_header("Host", f"{backend_host}:{backend_port}")
            client_origin = self.headers.get("Origin")
            if client_origin:
                req.add_header("Origin", client_origin)
            else:
                req.add_header("Origin", f"http://127.0.0.1:{self.server.server_port}")

            try:
                with urllib.request.urlopen(req, timeout=12) as resp:
                    status_code = resp.status
                    resp_headers = resp.headers
                    resp_body = resp.read()

                    self.send_response(status_code)
                    for k, v in resp_headers.items():
                        if k.lower() not in ["transfer-encoding", "connection"]:
                            self.send_header(k, v)
                    self.end_headers()
                    if send_body:
                        self.wfile.write(resp_body)
            except urllib.error.HTTPError as e:
                err_body = e.read()
                self.send_response(e.code)
                for k, v in e.headers.items():
                    if k.lower() not in ["transfer-encoding", "connection"]:
                        self.send_header(k, v)
                self.end_headers()
                if send_body:
                    self.wfile.write(err_body)
            except Exception as e:
                err_msg = (
                    f'{{"error":"backend unavailable","detail":"{str(e)}",'
                    f'"backend":"http://{backend_host}:{backend_port}"}}\n'
                ).encode("utf-8")
                self.send_response(503)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(err_msg)))
                self.end_headers()
                if send_body:
                    self.wfile.write(err_msg)

        def serve_ui_asset(self, clean_path, send_body=True):
            if clean_path in ["", "/", "/index.html"]:
                file_path = os.path.join(ui_dir, "index.html")
                content_type = "text/html; charset=utf-8"
                inject_layout = True
            elif clean_path == "/app.css":
                file_path = os.path.join(ui_dir, "app.css")
                content_type = "text/css; charset=utf-8"
                inject_layout = False
            elif clean_path == "/app.js":
                file_path = os.path.join(ui_dir, "app.js")
                content_type = "application/javascript; charset=utf-8"
                inject_layout = False
            else:
                rel_path = clean_path.lstrip("/")
                file_path = os.path.join(ui_dir, rel_path)
                if not os.path.exists(file_path) or not os.path.isfile(file_path):
                    self.send_error(404, f"Not found: {clean_path}")
                    return
                content_type = "application/octet-stream"
                inject_layout = False

            try:
                with open(file_path, "rb") as f:
                    content = f.read()

                if inject_layout:
                    html_str = content.decode("utf-8")
                    snippet = f'<script>window.__DEFAULT_LAYOUT__="{variant_id}";</script>'
                    if "</head>" in html_str:
                        html_str = html_str.replace("</head>", f"  {snippet}\n</head>", 1)
                    else:
                        html_str = snippet + html_str
                    content = html_str.encode("utf-8")

                self.send_response(200)
                self.send_header("Content-Type", content_type)
                self.send_header("Content-Length", str(len(content)))
                self.send_header("Cache-Control", "no-cache, no-store, must-revalidate")
                self.end_headers()
                if send_body:
                    self.wfile.write(content)
            except Exception as e:
                self.send_error(500, f"Error reading file: {e}")

    return VariantHTTPHandler


class ThreadingHTTPServer(http.server.ThreadingHTTPServer):
    allow_reuse_address = True
    daemon_threads = True


def start_server(port, variant_id, ui_dir, backend_host, backend_port):
    handler_class = create_variant_handler(variant_id, ui_dir, backend_host, backend_port)
    server = ThreadingHTTPServer(("0.0.0.0", port), handler_class)
    return server


def main():
    parser = argparse.ArgumentParser(description="agentUsage Multi-Variant UI Runner")
    parser.add_argument("--backend-host", default="127.0.0.1", help="Backend daemon host (default: 127.0.0.1)")
    parser.add_argument("--backend-port", type=int, default=8085, help="Backend daemon port (default: 8085)")
    parser.add_argument("--ui-dir", default=None, help="Path to internal/webserve/ui (default: auto-detected)")
    parser.add_argument("--p1", type=int, default=8080, help="Port for Design 1 (split) [default: 8080]")
    parser.add_argument("--p2", type=int, default=8081, help="Port for Design 2 (matrix) [default: 8081]")
    parser.add_argument("--p3", type=int, default=8082, help="Port for Design 3 (bento) [default: 8082]")

    args = parser.parse_args()

    # Resolve UI directory
    if args.ui_dir:
        ui_dir = os.path.abspath(args.ui_dir)
    else:
        repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
        ui_dir = os.path.join(repo_root, "internal", "webserve", "ui")

    if not os.path.isdir(ui_dir):
        sys.stderr.write(f"Error: UI directory not found at {ui_dir}\n")
        sys.exit(1)

    ports_map = {
        args.p1: "split",
        args.p2: "matrix",
        args.p3: "bento",
    }

    # Verify backend availability (warn if not yet running)
    backend_up = is_port_in_use(args.backend_port, args.backend_host)
    if not backend_up:
        sys.stderr.write(
            f"\n[NOTE] Backend server not detected on http://{args.backend_host}:{args.backend_port}.\n"
            f"       Start it with:\n"
            f"       go run ./cmd/agentusage serve --listen {args.backend_host}:{args.backend_port} --source direct --no-open\n\n"
        )

    servers = []
    threads = []
    stop_event = threading.Event()

    for port, variant in ports_map.items():
        try:
            srv = start_server(port, variant, ui_dir, args.backend_host, args.backend_port)
            servers.append(srv)
            t = threading.Thread(target=srv.serve_forever, daemon=True)
            t.start()
            threads.append(t)
        except OSError as e:
            sys.stderr.write(f"Error binding to port {port}: {e}\n")
            for s in servers:
                s.shutdown()
            sys.exit(1)

    print("=" * 68)
    print("  agentUsage Glanceable UI Variants — Live Data Server")
    print("=" * 68)
    for port, variant in ports_map.items():
        desc = VARIANT_DESCRIPTIONS.get(variant, variant)
        print(f"  Port {port}: http://localhost:{port}  →  [{variant.upper()}] {desc}")
    print("-" * 68)
    print(f"  Live Backend Target: http://{args.backend_host}:{args.backend_port}")
    backend_status = "ONLINE (ready)" if backend_up else "OFFLINE (waiting for daemon)"
    print(f"  Backend Status:      {backend_status}")
    print("=" * 68)
    print("  Press Ctrl+C to stop all servers.")
    print("")

    def handle_sig(sig, frame):
        stop_event.set()

    signal.signal(signal.SIGINT, handle_sig)
    signal.signal(signal.SIGTERM, handle_sig)

    try:
        while not stop_event.is_set():
            time.sleep(0.5)
    except KeyboardInterrupt:
        pass
    finally:
        print("\nShutting down variant servers...")
        for s in servers:
            s.shutdown()
        for t in threads:
            t.join(timeout=1.0)
        print("Done.")


if __name__ == "__main__":
    main()
