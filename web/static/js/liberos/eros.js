(function (global) {
    "use strict";


    var Eros = function () {
        // We don't want these to be modifiable from the outside world.
        var socket = undefined,
            users = [],
            chatRooms = [],
            mapPool = [],
            requests = {},
            txBase = 0,
            connected = false,
            authenticated = false,
            r = starbow.ErosRequests;

        this.users = [];
        this.chatRooms = [];
        this.divisions = [];
        this.state = 'unconnected';
        this.matchmakingState = 'idle';
        this.activeRegions = [];
        this.mapPool = [];


        function sendRequest(request) {
            if (!connected) {
                return;
            }
            var tx = ++txBase;

            requests[String(tx)] = request;

            socket.send(request.command + ' ' + tx + ' ' + request.payload.length + '\n' + request.payload);
        }

        function processServerMessage(command, payload) {
            if (command == "PNG") {
                console.log('has ping');
                sendRequest(new r.PingRequest(payload));
            } else if (command == "SSU") {
                var stats = protobufs.ServerStats.decode64(payload);
                console.log('There are ' + stats.active_users + ' users online.');
            } else if (command == "CHJ") {
                var stats = protobufs.ChatRoomUser.decode(payload);
                console.log('Joined channel ' + stats.room.name);
            }
        }

        function processMessage(command, tx, payload) {
            if (tx in requests) {
                var result = requests[tx].processPayload(command, payload);

                if (!result) {
                    console.log(tx + ': ' + command + ' command not handled by ' + requests[tx].command + ' request.')
                }
            } else {
                console.log('No request found for transaction ' + tx + '. ' + command + ' command.');
            }
        }

        function handshakeRequestComplete(request) {
            authenticated = true;
            console.log('authenticated woot');
            console.log(request.result);
        };

        this.users = function () {
            return users.slice(0);
        };

        this.chatRooms = function () {
            return chatRooms.slice(0);
        };

        this.mapPool = function () {
            return mapPool.slice(0);
        };

        this.disconnect = function () {
            if (typeof (socket) === 'undefined') {
                return;
            }

            if (socket.readyState !== 3) {
                socket.close();
            };

            users = [];
            chatRooms = [];
            mapPool = [],
                requests = [],
                connected = false,
                authenticated = false;
        };

        this.connect = function (server, username, password, callback) {
            this.disconnect();

            txBase = 0;
            socket = new WebSocket('ws://' + server + '/ws');

            socket.onopen = function () {
                connected = true;
                sendRequest(new r.HandshakeRequest(username, password, handshakeRequestComplete));
            };
            socket.onmessage = function (e) {
                var buffer = dcodeIO.ByteBuffer.wrap(e.data);
                var data = e.data.split('\n');
                var header = data[0];

                if (data[0] != '') {
                    header = data[0].split(' ');
                    console.log(header[0]);
                    if (header.length == 2) {
                        processServerMessage(header[0], data[1], Number(header[1]))
                    } else {
                        processMessage(header[0], Number(header[1]), data[1], Number(header[2]))
                    }
                }


            };
            socket.onclose = function (e) {
                console.log(e);
            };
        };
    };

    if (!global["starbow"]) {
        global["starbow"] = {};
    }
    global["starbow"]["Eros"] = Eros;
})(this);