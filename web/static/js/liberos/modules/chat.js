(function (global) {
    "use strict";

    var ChatRoom = function(eros, chat, r) {
    	
    	var users = {},
    	room = this;

    	if (typeof(r) === 'string') {
    		this.name = r;
    		this.key = r.toLowerCase().trim();
    	}

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

        if (typeof(r) === "object") {
        	update(r);
        }

        this.users = function () {
            var copy = {}
            for (var x in users) {
                copy[x] = users[x]
            }
            return copy;
        };

    };

    var ChatModule = function (eros, sendRequest, options) {
    	if (typeof (options) !== "object") {
    		options = {};
    	}

    	var chat = this,
    	rooms = {};

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

						if (typeof (options.joined) === "function") {
							options.joined(eros, room);
						}

						callback = options.userJoined;
					}
				} else if (command == "CHL") {
					if (eros.localUser == user) {
						delete rooms[room.key];
						if (typeof (options.left) === "function") {
							options.left(eros, room);
						}

						callback = options.userLeft;
					}
				}

				if (typeof (callback) === "function") {
					callback(eros, room, user);
				}

    			return true;
    		} else if (command == "CHM") {
    			var message = protobufs.ChatRoomMessage.decode64(payload);
    			var user = eros.user(message.sender.username);
				user.update(message.sender);

				var room = chat.room(message.room.key)
				room.update(message.room);

				if (typeof (options.message) === "function") {
					options.message(eros, room, user, message.message);
				}

				return true;
    		} else {
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
                var room = new ChatRoom(eros, name.trim());
                return room;
            }
        }

        this.sendToRoom = function(room, message) {
        	message = message.trim();
        	if (message == '') {
        		return;
        	}

        	sendRequest(new starbow.ErosRequests.ChatMessageRequest(room, message));
        }
    };

    global.starbow.Eros.prototype.modules.chat = ChatModule;
})(this);