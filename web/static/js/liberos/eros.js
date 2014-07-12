(function (global) {
    "use strict";


    var Eros = function (options) {
    	if (typeof (options) !== "object") {
    		options = {};
    	}
        // We don't want these to be modifiable from the outside world.
        var eros = this,
        	socket = undefined,
            users = [],
            mapPool = [],
            requests = {},
            txBase = 0,
            connected = false,
            authenticated = false,
            r = starbow.ErosRequests,
            commandHandlers = {},

            //Distinction: modules is our internal, last-loaded list of modules
            //this.modules is the public facing list that we load.
            modules = {};

        this.users = [];
        this.divisions = [];
        this.matchmakingState = 'idle';
        this.activeRegions = [];
        this.mapPool = [];
        this.stats = {
        	users: {
        		active: 0,
        		searching: 0
        	}
        };
        this.regions = {};


        function loadModules() {
            eros.commandHandlers = {};

            // Remove all our existing modules.
            for (var module in modules) {
                delete eros[module];
            }

            modules = {};
            for (var module in eros.modules) {
                // If we have options for this module, pass 'em on.
                var moduleOptions = undefined;
                if (typeof (options) === "object") {
                    if (typeof (options[module]) === "object") {
                        moduleOptions = options["module"]
                    }
                }

                // Init the module
                eros[module] = new eros.modules[module](eros, moduleOptions);

                // Add to our internal list.
                modules[module] = eros[module];

                // Register server command handlers
                if (typeof (eros[module].commandHandlers) === "object") {
                    for (var i in eros[module].commandHandlers) {
                        eros.commandHandlers[i] = eros[module].commandHandlers[i];
                    }    
                }
            }
        }

        loadModules();

        function regionFromCode(code) {
        	for (var i in protobufs.Region) {
        		if (protobufs.Region[i] == code) {
        			return i;
        		}
        	}

        	return null;
        }

        this.regionFromCode = regionFromCode;

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
                sendRequest(new r.PingRequest(payload, function(pong) {
                    eros.latency = pong.result;
                    if (typeof (options.latencyUpdate) === "function") {
                        options.latencyUpdate(eros);
                    }
                }));
            } else if (command == "SSU") {
            	// Server stats update
                var stats = protobufs.ServerStats.decode64(payload);
                
                var update = false;
                if (eros.stats.users.active != stats.active_users.low) {
                	eros.stats.users.active = stats.active_users.low;
                	update = true;
                }

                if (eros.stats.users.searching != stats.searching_users.low) {
                	eros.stats.users.searching = stats.searching_users.low;
                	update = true;
                }

                if (update && typeof (options.statsUpdate) === "function") {
                	options.statsUpdate(eros);
                }

                
                for (var i = 0; i < stats.region.length; i++) {
                	update = false;
                	var name = regionFromCode(stats.region[i].region);

                	if (!(name in eros.regions)) {
                		eros.regions[name] = {
                			active: false,
                			users: {
                				searching: 0
                			}
                		};
                		update = true;
                	}

                	if (eros.regions[name].users.searching != stats.region[i].searching_users.low) {
	                	eros.regions[name].users.searching = stats.region[i].searching_users.low;
	                	update = true;
	                }

	                if (update && typeof (options.regionUpdate) === "function") {
	                	options.regionUpdate(eros, name);
	                }
                }
            } else {
                if (command in eros.commandHandlers) {
                    if(!eros.commandHandlers[command](command, payload)) {
                        console.log(command + ': registered handler returned false.')
                    }
                } else {
                    console.log(command + ': no handler registered.')
                }
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
        	var callback = undefined
        	console.log(request);
        	if ((request.status === 0) && (request.result.status === 1)) {
        		callback = options.loggedIn;
        		authenticated = true;
        	} else {
        		callback = options.loginFailed;
        	}

        	if (typeof (callback) === "function") {
        		callback(this, request.result.status);
        	}

        	if (authenticated) {
                for (var i = 0; i < request.result.active_region.length; i++) {
                	var name = regionFromCode(request.result.active_region[i]);

            		eros.regions[name] = {
            			active: true,
            			users: {
            				searching: 0
            			}
            		};
                	
	                if (typeof (options.regionUpdate) === "function") {
	                	options.regionUpdate(eros, name);
	                }
                }
        	}
        };

        this.users = function () {
            return users.slice(0);
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
            mapPool = [];
            requests = [];
            connected = false;
            authenticated = false;
        };

        this.connect = function (username, password) {
            var server = window.location.host;

            if (typeof (options.server) === 'string') {
                server = options.server;
            }

            eros.disconnect();

            txBase = 0;
            socket = new WebSocket('ws://' + server + '/ws');

            socket.onopen = function () {
                loadModules();
                connected = true;
                eros.latency = 0;
                eros.regions = {};
                if (typeof (options.connected) === "function") {
                	options.connected(eros);
                }
                sendRequest(new r.HandshakeRequest(username, password, handshakeRequestComplete));
            };

            socket.onmessage = function (e) {
                var data = e.data.split('\n');
                var header = data[0];

                if (data[0] != '') {
                    header = data[0].split(' ');
                    if (header.length == 2) {
                        processServerMessage(header[0], data[1], Number(header[1]))
                    } else {
                        processMessage(header[0], Number(header[1]), data[1], Number(header[2]))
                    }
                }
            };

            socket.onclose = function (e) {
                connected = false;
                authenticated = false;

                if (typeof (options.disconnected) === "function") {
                	options.disconnected(eros);
                }
            };
        };

        this.isConnected = function() {
        	return connected && authenticated;
        }
    };

    Eros.prototype.modules = {};


    if (!global["starbow"]) {
        global["starbow"] = {};
    }
    global["starbow"]["Eros"] = Eros;
})(this);