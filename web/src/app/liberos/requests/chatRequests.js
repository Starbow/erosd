(function (global) {
    "use strict";

    var ChatMessageRequest = function (room, message, callback) {
        var chatMessage = new protobufs.ChatMessage("", room.key, message);

        var request = new starbow.ErosRequests.Request("UCM", chatMessage.toBase64(), function (command, payload) {
            if (command === "UCM") {
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            return false;
        });

        return request;
    };

    var PrivateMessageRequest = function (user, message, callback) {
        if(typeof user === "object"){
            user = user.username
        }
        // var privateMessage = new protobufs.ChatPrivateMessage(user, message);
        var privateMessage = new protobufs.ChatMessage("", user, message);

        var request = new starbow.ErosRequests.Request("UPM", privateMessage.toBase64(), function (command, payload) {
            if (command === "UPM") {
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            return false;
        });

        return request;
    };

    var ChatJoinRequest = function (room, password, callback) {
        var chatMessage = new protobufs.ChatRoomRequest(room.key, password);

        var request = new starbow.ErosRequests.Request("UCJ", chatMessage.toBase64(), function (command, payload) {
            if (command === "UCJ") {
                var info = protobufs.ChatRoomInfo.decode64(payload);
                room.update(info);

                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            return false;
        }, function(command){
            if(command == 505){
                console.log("Received 505: Can't join any more channels.")
                return {error: 505, message: "Can't join any more channels."}
            }
        });

        return request;
    };

    var ChatLeaveRequest = function (room, callback) {
        var chatMessage = new protobufs.ChatRoomRequest(room.key, '');

        var request = new starbow.ErosRequests.Request("UCL", chatMessage.toBase64(), function (command, payload) {
            if (command === "UCL") {
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            return false;
        });

        return request;
    };

    var ChatIndexRequest = function(callback){
        var request = new starbow.ErosRequests.Request("UCI", '', function(command, payload){
            if (command === "UCI"){
                request.result = protobufs.ChatRoomIndex.decode64(payload);
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            return false
        })

        return request;
    }

    if (!global["starbow"]) {
        global["starbow"] = {};
    }

    if (!global["starbow"]["ErosRequests"]) {
        global["starbow"]["ErosRequests"] = {};
    }

    global["starbow"]["ErosRequests"]["ChatMessageRequest"] = ChatMessageRequest;
    global["starbow"]["ErosRequests"]["PrivateMessageRequest"] = PrivateMessageRequest;
    global["starbow"]["ErosRequests"]["ChatJoinRequest"] = ChatJoinRequest;
    global["starbow"]["ErosRequests"]["ChatLeaveRequest"] = ChatLeaveRequest;
    global["starbow"]["ErosRequests"]["ChatIndexRequest"] = ChatIndexRequest;
})(this);