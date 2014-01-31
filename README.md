erosd
==========
erosd powers Eros, the Starbow matchmaking service.

www.starbowmod.com

Installation
==========
- Install [go1.2](http://golang.org/doc/install) for your platform.
- Install [goprotobuf](https://code.google.com/p/goprotobuf/) if you need to modify the protocol buffers.
- Install python and sc2reader.
- Set up your PATH and GOPATH.
- `go install github.com/Starbow/erosd`
- `erosd`

Todo (Short Term)
==========
- Reduce overall messyness.
- Latency measurement with pings.
- Commentify everything.
- Map pool in the database.
- Battle.net character management.
- Veto management.
- Chat rooms and private chat.
- No-show support and match timeouts.
- Lobbies?

Todo (Long Term)
==========
- Modify the ladder code to support more game types.
- Reloadable config.
- Logging.