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
- Commentify everything.
- Veto management.
- Make use of database transactions.

Todo (Long Term)
==========
- Modify the ladder code to support more game types.
- Reloadable config.
- Logging.

Tidbits
=========
Need to regenerate the protocol buffers?

`protoc --go_out=. buffers/eros.proto --plugin=$GOPATH/bin/protoc-gen-go`
