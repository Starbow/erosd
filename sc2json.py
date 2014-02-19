#!/usr/bin/env python
# -*- coding: utf-8 -*-
from __future__ import absolute_import, print_function, unicode_literals, division

import SocketServer
import sc2reader
from sc2reader.factories.plugins.replay import toJSON
sc2reader.register_plugin("Replay", toJSON(encoding='UTF-8', indent=None))


class ReplayParser(SocketServer.StreamRequestHandler):
    def handle(self):
        path = self.rfile.readline().strip()
        print("Parsing replay file: "+path)
        json = sc2reader.load_replay(path, load_level=2)
        self.wfile.write(json+"\n\n")


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Listens on .")
    parser.add_argument('PORT', metavar='port', type=int, nargs=1, help="The port to listen on.")
    args = parser.parse_args()

    HOST, PORT = 'localhost', args.PORT[0]
    print("Listening on: {}:{}".format(HOST, PORT))
    server = SocketServer.TCPServer((HOST, PORT), ReplayParser)
    server.serve_forever()

if __name__ == '__main__':
    main()
