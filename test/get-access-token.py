#!/usr/bin/python3
"""
Simple example how to utilize OAuth2 with Kopano.
"""

import codecs
import http
import httplib2
import json
import os
import urllib.parse
import queue
import threading
import webbrowser
import sys

from http.server import BaseHTTPRequestHandler, HTTPServer

import requests
from requests_oauthlib import OAuth2Session

INSECURE = os.environ.get("INSECURE", None) and True
HTTPD_HOST = os.environ.get("HTTPD_HOST", "localhost")
HTTPD_PORT = int(os.environ.get("HTTPD_PORT", "8080"), 10)


class Client:
    insecure = INSECURE
    httpd_host = HTTPD_HOST
    httpd_port = HTTPD_PORT

    prompt = "select_account"

    def __init__(self, iss, scope, client_id, client_secret=""):
        self.iss = iss
        self.scope = scope
        self.client_id = client_id
        self.client_secret = client_secret

        self.q = queue.Queue()

        self.httpd = HTTPServer((self.httpd_host, self.httpd_port), Handler)
        self.httpd.q = self.q

    def start(self):
        oidc = json.loads(requests.get(
            iss+"/.well-known/openid-configuration",
            verify=not self.insecure,
        ).text)

        oauth = OAuth2Session(
            client_id,
            redirect_uri="http://%s:%d/" % (
                self.httpd_host,
                self.httpd_port,
            ),
            scope=scope)

        authorization_url, state = oauth.authorization_url(
            oidc['authorization_endpoint'],
            state=codecs.encode(os.urandom(32), 'hex').decode('utf-8'),
            prompt=self.prompt,
        )

        serve = threading.Thread(target=self.serve)
        serve.start()

        print(authorization_url)
        webbrowser.open(authorization_url)

        authorization_response = self.q.get()

        if self.httpd is not None:
            self.httpd.shutdown()
        serve.join()

        if authorization_response is None:
            return

        token = oauth.fetch_token(
            oidc['token_endpoint'],
            client_id=self.client_id,
            client_secret=self.client_secret,
            authorization_response=authorization_response,
            verify=not self.insecure
        )

        return token

    def serve(self):
        self.httpd.serve_forever()

    def shutdown(self):
        if self.httpd is not None:
            self.httpd.shutdown()
            self.httpd = None
        self.q.put(None)


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)

        self.send_header('Content-type','text/html')
        self.end_headers()

        message = "Done - you can close this window now."
        self.wfile.write(bytes(message, "UTF-8"))

        self.server.q.put(self.path)

    def log_request(self, *args, **kwargs):
        pass

if __name__ == "__main__":
    iss = os.environ.get("ISS", None)
    if not iss:
        print("No ISS in environment - abort", file=sys.stderr)
        sys.exit(1)

    scope = os.environ.get("SCOPE", "openid")
    prompt = os.environ.get("PROMPT", None)

    client_id = os.environ.get("CLIENT_ID")
    if not client_id:
        client_id = "get-access-token.py"

    if INSECURE:
        os.environ.setdefault("OAUTHLIB_INSECURE_TRANSPORT", "1")
        import urllib3
        urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

    print("ISS                 : {}".format(iss))
    print("CLIENT_ID           : {}".format(client_id))
    client_secret = os.environ.get("CLIENT_SECRET")
    try:
        client = Client(iss, scope, client_id, client_secret)
        if prompt is not None:
            client.prompt = prompt
        token = client.start()
        at = token['access_token']
        print("ACCESS_TOKEN_VALUE  : {}".format(at))
        rt = token.get('refresh_token')
        if rt:
            print("REFRESH_TOKEN_VALUE : {}".format(rt))

    except KeyboardInterrupt:
        print("Interrupted", file=sys.stderr)
        client.shutdown()
        sys.exit(1)
        pass
