"""
mitmproxy addon: logs HTTP request/response pairs to ok's SQLite DB.

Usage:
  mitmdump -s proxy/capture.py --set capture_id=<ID> --set db_path=<PATH>
"""

import json
import time
import sqlite3
import uuid
from mitmproxy import http, ctx

class CaptureAddon:
    def __init__(self):
        self.db = None
        self.seq = 0

    def load(self, loader):
        loader.add_option("capture_id", str, "", "Capture session ID")
        loader.add_option("db_path", str, "", "Path to SQLite database")

    def running(self):
        self.db = sqlite3.connect(ctx.options.db_path)
        self.db.execute("PRAGMA journal_mode=WAL")

    def response(self, flow: http.HTTPFlow):
        # Only log requests with JSON bodies (API calls)
        content_type = flow.response.headers.get("content-type", "")
        if "json" not in content_type and "json" not in flow.request.headers.get("content-type", ""):
            return

        req_body = None
        if flow.request.content:
            try:
                req_body = json.loads(flow.request.content)
            except (json.JSONDecodeError, UnicodeDecodeError):
                req_body = None

        res_body = None
        if flow.response.content:
            try:
                res_body = json.loads(flow.response.content)
            except (json.JSONDecodeError, UnicodeDecodeError):
                res_body = None

        if req_body is None and res_body is None:
            return

        self.seq += 1
        latency_ms = int((flow.response.timestamp_end - flow.request.timestamp_start) * 1000)

        self.db.execute(
            """INSERT INTO capture_turns (id, capture_id, seq, host, path, method, request_json, response_json, status_code, latency_ms, created_at)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))""",
            (
                str(uuid.uuid4()),
                ctx.options.capture_id,
                self.seq,
                flow.request.host,
                flow.request.path,
                flow.request.method,
                json.dumps(req_body) if req_body else None,
                json.dumps(res_body) if res_body else None,
                flow.response.status_code,
                latency_ms,
            ),
        )
        self.db.commit()

    def done(self):
        if self.db:
            self.db.close()

addons = [CaptureAddon()]
