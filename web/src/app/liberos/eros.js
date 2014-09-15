(function (global) {
    "use strict";

    var mock = {
        getRandom: function(value_array){
            if(value_array.length == 0){
                return 
            }
            var rand = Math.random();
            rand = rand * value_array.length;
            rand = Math.floor(rand);

            return value_array;
        },

        getRandomInt: function(min, max){
            if(min >= max){
                return
            }
            var rand = Math.random();
            rand = Math.floor(rand*(max-min))

            return min+rand
        }

    }


    var ErosUserStats = function(u) {
        var stats = this;

        stats.division = 0;
        stats.divisionRank = 0;
        stats.forfeits = 0;
        stats.losses = 0;
        stats.mmr = 0;
        stats.placementsRemaining = 0;
        stats.points = 0;
        stats.wins = 0;
        stats.walkovers = 0;

        function update(u) {
            var mocker = eros.isTest()
            if(mocker == true){
                stats.division = getDivision(mock.getRandomInt(0,6));
                stats.divisionRank = getRank(mock.getRandomInt(0,30));
            }else{
                stats.division = getDivision(u.division.low);
                stats.divisionRank = getRank(u.division_rank.low);
            }
            
            stats.forfeits = u.forfeits.low;
            stats.losses = u.losses.low;
            stats.mmr = u.mmr;
            stats.placementsRemaining = u.placements_remaining.low;
            stats.points = u.points.low;
            stats.wins = u.wins.low;
            stats.walkovers = u.walkovers.low;
        }

        function getDivision(division){
            var divisions = ["P", "E", "D", "C", "B", "A"];

            return divisions[division]
        }

        function getRank(rank){
            return rank;
        }

        this.update = update;

        if (typeof (u) === 'object') {
            update(u);
        }
    }

    var ErosUser = function(eros, u) {  
        var user = this;    

        if (typeof (u) === 'string') {
            user.username = u;
            user.stats = new ErosUserStats(u);
        } else {
            user.username = u.username;
            user.stats = new ErosUserStats(u);
        }


        this.regions = {};

        function update(u) {
            user.stats.update(u);
            user.id = u.id.low;

            for (var i = 0; i < u.region.length; i++) {
                var region = u.region[i];
                var name = eros.regionFromCode(region.region);

                if (!(name in user.regions)) {
                    user.regions[name] = new ErosUserStats(region);
                } else {
                    user.regions[name].update(region);
                }
            }
        }

        this.update = update;
        if (typeof (u) === 'object') {
            update(u);
        }
    };

    var Eros = function (options) {
    	if (typeof (options) !== "object") {
    		options = {};
    	}
        // We don't want these to be modifiable from the outside world.
        var eros = this,
        	socket = undefined,
            users = {},
            requests = {},
            txBase = 0,
            connected = false,
            authenticated = false,
            r = starbow.ErosRequests,
            commandHandlers = {},

            //Distinction: modules is our internal, last-loaded list of modules
            //this.modules is the public facing list that we load.
            modules = {};

        this.isTest=function(){
            return window.document.baseURI == 'http://localhost:9090/'
        }

        
        function reset() {
            eros.matchmakingState = 'idle';
            eros.activeRegions = [];
            eros.mapPool = [];
            eros.stats = {
                users: {
                    active: 0,
                    searching: 0
                }
            };
            eros.localUser = {};
            eros.regions = {};
            eros.ladder = {
                maps: [],
                divisions: [],
                vetoes: 0,
            };
            eros.latency = 0;
            users = {};
        };


        function sendRequest(request) {
            if (!connected) {
                return;
            }
            var tx = ++txBase;

            requests[String(tx)] = request;

            socket.send(request.command + ' ' + tx + ' ' + request.payload.length + '\n' + request.payload);

            return request
        }

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
                        moduleOptions = options[module];
                    }
                }

                console.log(moduleOptions);
                // Init the module
                eros[module] = new eros.modules[module](eros, sendRequest, moduleOptions);

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
                	var name = eros.regionFromCode(stats.region[i].region);

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
            } else if (command == "USU") {
                processUserStats(protobufs.UserStats.decode64(payload));
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

        function processUserStats(stats) {
            var key = stats.username.toLowerCase();

            if (key in users) {
                users[key].update(stats);

                var callback;
                if (users[key] === localUser) {
                    callback = options.localUserStatsUpdate;
                } else {
                    callback = options.userStatsUpdate;
                }

                if (typeof (callback) === "function") {
                    callback(eros, users[key]);
                }
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

            if (authenticated) {
                eros.localUser = new ErosUser(eros, request.result.user);
                eros.localUser.local = true;
                users[eros.localUser.username.toLowerCase()] = eros.localUser;
            }

        	if (typeof (callback) === "function") {
        		callback(eros, request.result.status);
        	}

        	if (authenticated) {
                for (var i = 0; i < request.result.active_region.length; i++) {
                	var name = eros.regionFromCode(request.result.active_region[i]);

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
            var copy = {}
            for (var x in users) {
                copy[x] = users[x]
            }
            return copy;
        };

        this.user = function(username) {
            var key = username.toLowerCase();

            if (key in users) {
                return users[key];
            } else {
                var user = new ErosUser(eros, username.trim());
                user.local = false;
                users[key] = user;

                return user;
            }
        }

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

            users = {};
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
                reset();
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

    Eros.prototype.regionFromCode = function (code) {
        for (var i in protobufs.Region) {
            if (protobufs.Region[i] == code) {
                return i;
            }
        }

        return null;
    }

    Eros.prototype.modules = {};


    if (!global["starbow"]) {
        global["starbow"] = {};
    }
    global["starbow"]["Eros"] = Eros;
})(this);