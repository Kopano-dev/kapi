#!/usr/bin/python3
"""
Simple example how to get OAuth2 credentials from the commandline.

For this script to work, the OpenID Connect Identifier must accept the CLIENT_ID
and CLIENT_SECRET. This script must be registered as a `native` appplication to
allow the local redirect callback to work.

Usage: ISS=https://your-kopano.local SCOPE="openid kopano/gc" \
         CLIENT_ID=my-client-id CLIENT_SECRET=my-client-secret \
         ./get-access-token.py --format env

  See `./get-access-token.py --help for further usage infos.

Environment variables supported:

  ISS           : Issuer identifier (required).
  CLIENT_ID     : Client ID to use for OAuth2 requests.
  CLIENT_SECRET : Client secret to use for OAuth token request.
  SCOPE         : Scope to use for OAuth2 authorize request.
  PROMPT        : Prompt value to use for OAuth2 authorize request.

  INSECURE=1    : Disables TLS validation.

Runtime Python dependencies:

 - Python 3
 - requests
 - requests_oauthlib

"""

import codecs
import json
import os
import queue
import threading
import sys

from http.server import BaseHTTPRequestHandler, HTTPServer

import requests
from requests_oauthlib import OAuth2Session

INSECURE = os.environ.get("INSECURE", None) and True
HTTPD_HOST = os.environ.get("HTTPD_HOST", "localhost")
HTTPD_PORT = int(os.environ.get("HTTPD_PORT", "8080"), 10)


class RedirectAuthenticator(object):
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
        self.httpd = None

    def start(self, noauth_local_webserver=False):
        if self.httpd is not None and not noauth_local_webserver:
            raise RuntimeError("already started")

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

        if not noauth_local_webserver:
            import webbrowser
            webbrowser.open(authorization_url)
            print('Your browser has been opened to visit:')
            print()
            print('    ' + authorization_url)
            print('If your browser is on a different machine then '
                  'exit and re-run this')
            print('application with the command-line parameter ')
            print()
            print('  --noauth_local_webserver')
            print()
            httpd = HTTPServer(
                (self.httpd_host, self.httpd_port),
                handlerFactory(self),
            )
            self.httpd = httpd
            serve = threading.Thread(target=lambda: httpd.serve_forever())
            serve.start()

            authorization_response = self.q.get()

            if self.httpd is httpd:
                httpd.shutdown()
                self.httpd = None
            serve.join()

        else:
            print('Go to the following link in your browser:')
            print()
            print('    ' + authorization_url)
            print()

            authorization_response = input("Enter full callback URL: ").strip()
            if not authorization_response:
                raise ValueError("no callback URL given")

        if authorization_response is None:
            return

        print("Authentication successful.")

        token = oauth.fetch_token(
            oidc['token_endpoint'],
            client_id=self.client_id,
            client_secret=self.client_secret,
            authorization_response=authorization_response,
            verify=not self.insecure
        )

        return token

    def shutdown(self):
        if self.httpd is not None:
            self.httpd.shutdown()
            self.httpd = None
            self.q.put(None)


def handlerFactory(app):
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self):
            self.send_response(200)

            self.send_header('Content-type', 'text/html')
            self.end_headers()

            self.wfile.write(okPageHTML)

            app.q.put(self.path)

        def log_request(self, *args, **kwargs):
            pass

    return Handler

okPageHTML = b"""
<!DOCTYPE html>
<html>
<head>
<title>OAuth2 callback</title>
<style>
body {
    font-family: sans-serif;
}
</style>
</head>
<body>
Done - you can close this window now.
</body>
</html>
"""

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description='OAuth2 with Kopano.')
    parser.add_argument('--auth_host_name', default=HTTPD_HOST,
                        help='Hostname when running a local web server.')
    parser.add_argument('--auth_host_port', default=[HTTPD_PORT], type=int,
                        nargs='*', help='Port web server should listen on.')
    parser.add_argument('--noauth_local_webserver', action='store_true',
                        default=False, help='Do not run a local web server.')
    parser.add_argument('--output', help='Write auth result to file.')
    parser.add_argument('--force', action='store_true',
                        default=False, help='Allow overwrite of output file.')
    parser.add_argument('--format', default="env", choices=["env", "json"],
                        help='Selects the output format.')

    flags = parser.parse_args()
    print(flags)

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
        app = RedirectAuthenticator(iss, scope, client_id, client_secret)
        if prompt is not None:
            app.prompt = prompt
        token = app.start(noauth_local_webserver=flags.noauth_local_webserver)
        output = sys.stdout
        try:
            if flags.output:
                output = open(flags.output, flags.force and "w" or "x")
            if flags.format == "json":
                print(json.dumps(token, indent=2, sort_keys=True), file=output)
            elif flags.format == "env":
                print("TOKEN_VALUE={}".format(token["access_token"]), file=output)
                print("EXPIRES_AT={}".format(token["expires_at"]), file=output)
                print("EXPIRES_IN={}".format(token["expires_in"]), file=output)
                print("TOKEN_TYPE={}".format(token["token_type"]), file=output)
                rt = token.get("refresh_token")
                if rt:
                    print("REFRESH_TOKEN_VALUE={}".format(rt), file=output)
                it = token.get("id_token")
                if it:
                    print("ID_TOKEN_VALUE={}".format(it), file=output)
            else:
                raise ValueError("unknown output format", flags.format)
        finally:
            if flags.output:
                output.close()

    except KeyboardInterrupt:
        print("Interrupted", file=sys.stderr)
        app.shutdown()
        sys.exit(1)
        pass
