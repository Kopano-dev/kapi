#!/usr/bin/python3
"""
Simple interactive Python Kopano console which automatically signs into
Kopano with OIDC.

Usage: KOPANO_SOCKET=default: TOKEN_VALUE=$TOKEN_VALUE ./oidc-pyko-console.py"

  KOPANO_SOCKET : Socket to Kopano server, can be unix, http or https.
  TOKEN_VALUE   : Kopano Konnect Access Token.

"""

import base64
import json
import os
try:
    import readline
    import rlcompleter
except ImportError:
    readline = None
    pass
import traceback
from code import InteractiveConsole

import kopano


def main(server_socket=None, access_token=None):
    print("KOPANO_SOCKET : {}".format(server_socket))
    print("TOKEN_VALUE   : {}".format(access_token))
    helper = Helper(server_socket, access_token)

    server = None
    if access_token:
        try:
            server = helper.getServer()
            helper.defines.set("server", server)
        except:
            traceback.print_exc()

    if readline:
        readline.set_completer(rlcompleter.Completer(helper.defines).complete)
        readline.parse_and_bind("tab: complete")

    console = InteractiveConsole(locals=helper.defines)
    console.interact(banner="")


class Helper:
    def __init__(self, server_socket, access_token):
        self.defines = {
            "server_socket": server_socket,
            "access_token": access_token,
            "server": None,
            "kopano": kopano,
            "helper": self,
        }

    def getServer(self, server_socket=None, access_token=None):
        if server_socket is None:
            server_socket = self.defines.get("server_socket")
        if access_token is None:
            access_token = self.defines.get("access_token")

        payload = self.decodeAccessToken(access_token)

        auth_user = payload["kc.identity"]["kc.i.id"]
        server = kopano.Server(server_socket=server_socket, auth_user=auth_user, auth_pass=access_token, oidc=True)
        return server

    def decodeAccessToken(self, access_token=None):
        if access_token is None:
            access_token = self.defines.get("access_token")

        # Poor mans JWT parse.
        header, body, signature = access_token.split(".", 3)
        # Poor mans base64URL decode - this works in Python since it only complains
        # when padding is missing.
        return json.loads(base64.decodestring(body.encode('ascii') + b"===").decode("UTF-8"))

if __name__ == "__main__":
    server_socket = os.environ.get("KOPANO_SOCKET", None)
    access_token = os.environ.get("TOKEN_VALUE", None)
    main(server_socket, access_token)
