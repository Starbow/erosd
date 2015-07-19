erosd
==========
erosd powers Eros, the Starbow matchmaking service.

www.starbowmod.com

Installing the erosd backend
==========
- Install [go](http://golang.org/doc/install) for your platform.
- Install [goprotobuf](https://code.google.com/p/goprotobuf/) if you need to modify the protocol buffers.
- Install python and sc2reader (`pip install sc2reader`). (Not needed if you don't intend to upload replays.)
- Set up your PATH and GOPATH, for example:
````
echo "export GOPATH=~/.go" >> ~/.bashrc
echo "export PATH=$GOPATH/bin:$PATH" >> ~/.bashrc
````
- Then:

````
$ go get github.com/Sikian/oauth2
$ cd $GOPATH/src/github.com/Sikian/oauth2
$ git checkout authenticatedExchange
$ go install github.com/Starbow/erosd
$ $GOPATH/bin/erosd
````

`erosd` will proceed to write a default configuration file `erosd.cfg`
in the current directory, then croak, complaining about https
certificates.  Open `erosd.cfg` and set testmode=true.  This will
enable continuing without certificates as well as a number of
development-specific features.

Installing the web client
==========

The `web` directory contains the erosjs web client.

````
$ cd web
$ sudo npm install -g grunt
$ sudo npm install -g bower
$ bower update
$ npm install
$ grunt
````

This will put the web client's processed files in `dist`.

To actually access the web client from the web (or locally) you'll
need to set up your server using the following nginx config or its
equivalent for other webserver software.

TODO: explain how to generate a self-signed ssl certificate (useable
for development).

````
server {
    listen 80;

    # Comment out if not interested in ssl (won't be able to add
    # characters through Blizzard's API).
    listen 443 ssl;
    ssl_certificate PATH/TO/server.crt;
    ssl_certificate_key PATH/TO/server.key;

    root PATH_TO_EROSD/web/dist;
    index index.html;

    server_name YOUR_DOMAIN; # same domain that erosd is running on

    location / {
        try_files $uri $uri/ =404;
    }
    location /static {
        alias PATH_TO_EROSD/web/dist;
    }

    # proxying websocket connections
    location /ws {
        proxy_pass http://localhost:9090/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # redirecting bnet API callback to erosd
    location /login/battlenet {
        proxy_pass http://localhost:9090/login/battlenet;
    }

}
````

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
