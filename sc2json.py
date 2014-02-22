#!/usr/bin/env python
# -*- coding: utf-8 -*-
from __future__ import absolute_import, print_function, unicode_literals, division

import logging
import os
import argparse
import shutil
import SocketServer

# Set up sc2reader
import sc2reader
from sc2reader.factories.plugins.replay import toJSON
sc2reader.register_plugin("Replay", toJSON(encoding='UTF-8', indent=None))

# Set up logging
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter(
    fmt='%(asctime)s - %(name)s [%(levelname)s] - %(message)s',
    datefmt='%Y%m%dT%H%M%S'
))
logger = logging.getLogger('sc2json')
logger.setLevel(logging.INFO)
logger.addHandler(handler)


# Create our socket server
class ReplayParser(SocketServer.StreamRequestHandler):
    def handle(self):
        path = self.rfile.readline().strip()
        logger.info("Parsing replay file: {}".format(path))
        try:
            json = sc2reader.load_replay(path, load_level=2)
            self.wfile.write(json+"\n\n")
        except Exception as e:
            logger.exception("Error parsing {}".format(path))
            try:
                shutil.copy(path, self.server.replaydir)
            except:
                logger.exception("Error saving {} to {}. Aborting.".format(path, self.server.replaydir))


def main():
    # Handle basic commandline args
    parser = argparse.ArgumentParser(description="Listens on .")
    parser.add_argument('PORT', metavar='port', type=int, nargs=1,
                        help="The port to listen on.")
    parser.add_argument('REPLAYDIR', metavar='replaydir', type=str, nargs=1,
                        help="The directory to save failed replays to.")
    args = parser.parse_args()

    # Don't start if the replaydir is bad
    if not os.path.isdir(args.REPLAYDIR[0]):
        logger.fatal("{} is not a directory or does not exist".format(args.REPLAYDIR[0]))
        return
    else:
        try:
            f = open(os.path.join(args.REPLAYDIR[0], '12l3kn421l3i480o2i324n23lk4'), 'w')
            f.close()
        except:
            logger.fatal("{} is not writable. Check your permissions.".format(args.REPLAYDIR[0]))
            return

    # Spawn our server
    HOST, PORT = 'localhost', args.PORT[0]
    logger.info("Listening on: {}:{}".format(HOST, PORT))
    logger.info("Saving replay failures to: {}".format(args.REPLAYDIR[0]))
    server = SocketServer.TCPServer((HOST, PORT), ReplayParser)
    server.replaydir = args.REPLAYDIR[0]

    # And shut down gracefully
    try:
        server.serve_forever()
    except:
        server.shutdown()


if __name__ == '__main__':
    main()
