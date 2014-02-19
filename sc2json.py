#!/usr/bin/env python
# -*- coding: utf-8 -*-
from __future__ import absolute_import, print_function, unicode_literals, division

# Set up sc2reader
import sc2reader
from sc2reader.factories.plugins.replay import toJSON
sc2reader.register_plugin("Replay", toJSON(encoding='UTF-8', indent=None))

# Set up logging
import logging
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter(
    fmt='%(asctime)s - %(name)s [%(levelname)s] - %(message)s',
    datefmt='%Y%m%dT%H%M%S'
))
logger = logging.getLogger('sc2json')
logger.setLevel(logging.INFO)
logger.addHandler(handler)

# Create our socket server
import SocketServer


class ReplayParser(SocketServer.StreamRequestHandler):
    def handle(self):
        path = self.rfile.readline().strip()
        logger.info("Parsing replay file: "+path)
        json = sc2reader.load_replay(path, load_level=2)
        self.wfile.write(json+"\n\n")


def main():
    # Handle basic commandline args
    import argparse
    parser = argparse.ArgumentParser(description="Listens on .")
    parser.add_argument('PORT', metavar='port', type=int, nargs=1, help="The port to listen on.")
    args = parser.parse_args()

    # Spawn our server
    HOST, PORT = 'localhost', args.PORT[0]
    logger.info("Listening on: {}:{}".format(HOST, PORT))
    server = SocketServer.TCPServer((HOST, PORT), ReplayParser)

    # And shut down gracefully
    try:
        server.serve_forever()
    except:
        server.shutdown()


if __name__ == '__main__':
    main()
