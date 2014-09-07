(function (global) {
    "use strict";

    var ChatRoom = function(eros, chat, registerJoin, registerLeave, joinCallback, leaveCallback) {
    	var users = {},
    	room = this;

    	if (typeof(chat) === 'string') {
    		this.name = chat;
    		this.key = chat.toLowerCase().trim();
    	}

        registerJoin(function(user) {
            var key = user.username.toLowerCase();

            if (!(key in users)) {
                users[key] = user;
                return true;
            }

            return false;
        });

        registerLeave(function(user) {
            var key = user.username.toLowerCase();

            if ((key in users)) {
                delete users[key];
                return true;
            }

            return false;
        });

    	this.fixed = false;
    	this.joinable = false;
    	this.forced = false;
    	this.passworded = false;


        function update(r) {
            room.fixed = r.fixed;
            room.forced = r.forced;
            room.joinable = r.joinable;
            room.passworded = r.passworded;
            room.name = r.name;
            room.key = r.key;

            if (r.participant.length > 0) {
            	// Add users we don't have
            	for (var i = 0; i < r.participant.length; i++) {
            		var key = r.participant[i].username.toLowerCase();
            		var user = eros.user(key);
            		user.update(r.participant[i]);

            		if (!(key in users)) {
            			users[key] = user
            		}
            	}


            	// Remove users we don't have locally.
            	var remove = [];
            	for (var user in users) {
            		var found = false;
            		for (var i = 0; i < r.participant.length; i++) {
            			var key = r.participant[i].username.toLowerCase();
            			if (key == user) {
            				found = true;
            				break;
            			}
            		}

            		if (!found) {
            			remove.push(user);
            		}
            	}

            	for (var i = 0; i < remove.length; i++) {
            		delete users[i];
            	}
            }
        }

        this.update = update;

        if (typeof(chat) === "object") {
        	update(chat);
        }

        this.users = function () {
            var copy = {}
            for (var x in users) {
                copy[x] = users[x]
            }
            return copy;
        };

        this.join = function(password) {
            joinCallback(password);
        }

        this.leave = function() {
            leaveCallback();
        }

    };

    var ChatPrivate = function(eros, user, leaveCallback){
        // var user,
        var priv = this;

        if(typeof(user) === 'object'){
            var user = user
        }else if (typeof(user) === 'string') {
            var user = eros.user(user)
        }

        this.name = user.username;
        this.key = user.username.toLowerCase().trim();

        this.user = function() {
            var copy = user;
            return copy;
        }

        this.join = function(password) {
            joinCallback();
        }

        this.leave = function() {
            leaveCallback();
        }
    }

    var ChatModule = function (eros, sendRequest, options) {
    	if (typeof (options) !== "object") {
    		options = {};
    	}

    	var chat = this,
    	rooms = {},
        roomJoinedHandlers = {},
        roomLeftHandlers = {},
        selected = "",
        privs = {};

    	function processServerMessage(command, payload) {
    		if (command == "CHJ" || command == "CHL") {
    			var roomUser = protobufs.ChatRoomUser.decode64(payload);

				var user = eros.user(roomUser.user.username);
				user.update(roomUser.user);

				var room = chat.room(roomUser.room.key)
				room.update(roomUser.room);

				var callback = undefined;
				if (command == "CHJ") {
					if (!(room.key in rooms)) {
						rooms[room.key] = room;
                    }

                    if (eros.localUser == user) {
						if (typeof (options.joined) === "function") {
							options.joined(eros, room);
						}
                    } else {
						callback = options.userJoined;
                        roomJoinedHandlers[room.key](user);
					}
				} else if (command == "CHL") {
                    // Doesn't get sent when local user leaves.
                    roomLeftHandlers[room.key](user);

					if (eros.localUser == user) {
						delete rooms[room.key];
                        delete roomJoinedHandlers[room.key];
                        delete roomLeftHandlers[room.key];
						if (typeof (options.left) === "function") {
							options.left(eros, room);
						}
					} else {
                        callback = options.userLeft;
                    }
				}

				if (typeof (callback) === "function") {
					callback(eros, room, user);
				}

    			return true;
    		} else if (command == "CHM") { // Chat message
    			var message = protobufs.ChatRoomMessage.decode64(payload);
    			var user = eros.user(message.sender.username);
				user.update(message.sender);

				var room = chat.room(message.room.key)
				room.update(message.room);

				if (typeof (options.message) === "function") {
					options.message(eros, room, user, message.message);
				}

				return true;
    		} else if (command == "CHP"){ // Private message
                var message = protobufs.ChatPrivateMessage.decode64(payload);
                var senderUser = eros.user(message.sender.username);
                // user.update(message.sender);

                // var targetUser = eros.user(message.target.username); 
                var targetUser = eros.localUser
                var priv_return = chat.priv(message.sender.username)
                var priv = priv_return[0]
                var joined = priv_return[1]

                if(!joined){
                    if (typeof (options.privjoined) === "function") {
                        options.privjoined(eros, priv);
                    } 
                }

                // Display message
                if (typeof (options.privmessage) === "function") {
                    options.privmessage(eros, priv, senderUser, message.message);
                }
            } else{
    			return false;
    		}
    	}

    	this.commandHandlers = {
    		"CHJ": processServerMessage, // User joined chat
    		"CHL": processServerMessage, // User left chat
    		"CHM": processServerMessage, // Incoming chat room messae
    		"CHP": processServerMessage  // Incoming private message
    	}

        this.rooms = function () {
            var copy = {}
            for (var x in rooms) {
                copy[x] = rooms[x]
            }
            return copy;
        };

        this.room = function(name) {
            var key = name.toLowerCase().trim();

            if (key in rooms) {
                return rooms[key];
            } else {
                var chat = this;

                var room = new ChatRoom(
                    eros,
                    name.trim(),
                    function(joinHandler) { // user joined room callback
                        roomJoinedHandlers[key] = joinHandler;
                    },
                    function(leaveHandler) { // user left room callback
                        roomLeftHandlers[key] = leaveHandler;
                    },
                    function(password) { // room.join() proxy handler
                        chat.joinRoom(this, password);
                    },
                    function() { // room.leave() proxy handler
                        // chat.leaveRoom(this);
                        chat.leaveRoom(room)
                    }
                );
                return room;
            }
        };

        this.joinRoom = function(room, password) {
            if(typeof "room" == "string"){
                room = this.room(room)
            }
            sendRequest(new starbow.ErosRequests.ChatJoinRequest(room, password));
        };

        this.leaveRoom = function(room, password) {
            var chat = this;
            sendRequest(new starbow.ErosRequests.ChatLeaveRequest(room, function() {
                delete rooms[room.key];
                // delete roomJoinedHandlers[room.key];
                // delete roomLeftHandlers[room.key];
                if (typeof (options.left) === "function") {
                    options.left(eros, room);
                }
            }));
        };

        this.sendToRoom = function(room, message) {
        	message = message.trim();
        	if (message == '') {
        		return;
        	}

        	sendRequest(new starbow.ErosRequests.ChatMessageRequest(room, message));
        };

        // Private messages

        this.privs = function(){
            var copy = {}
            for (var x in privs) {
                copy[x] = privs[x]
            }
            return copy;
        };

        this.priv = function(user){
            var key = user.toLowerCase().trim();

            if (key in privs) {
                return [privs[key], true];
            } else {
                var chat = this;

                var priv = new ChatPrivate(
                    eros,
                    user,
                    function() { // priv.leave() proxy handler
                        chat.leavePriv(priv)
                    }
                );
                return [priv, false];
            }
        };

        this.leavePriv = function(room, password) {
            var chat = this;
            delete privs[room.key];
            // delete roomJoinedHandlers[room.key];
            // delete roomLeftHandlers[room.key];
            if (typeof (options.privleave) === "function") {
                options.privleave(eros, room);
            }
        };

        this.sendToPriv = function(user, message, ensure_response){
            message = message.trim();
            if (message == '') {
                return;
            }
            if (user == eros.localUser.username){
                return;
            }

            var writeToChannel = function(){
                // Only if request is successful
                var priv_return = chat.priv(user)
                var priv = priv_return[0]
                var joined = priv_return[1]

                var senderUser = eros.user(user)

                if(!joined){
                    if (typeof (options.privjoined) === "function") {
                        options.privjoined(eros, priv);
                    } 
                }

                // Display message
                if (typeof (options.privmessage) === "function") {
                    options.privmessage(eros, priv, eros.localUser, message);
                } 

            }

            sendRequest(new starbow.ErosRequests.PrivateMessageRequest(user, message, function(result){
                if(ensure_response){
                    writeToChannel()
                }
            }))

            if(!ensure_response && eros.users()[user]){
                writeToChannel()
            }


        }
    };

    global.starbow.Eros.prototype.modules.chat = ChatModule;
})(this);