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
    }

    if (!global["starbow"]) {
        global["starbow"] = {};
    }

    if (!global["starbow"]["ErosRequests"]) {
        global["starbow"]["ErosRequests"] = {};
    }

    global["starbow"]["ErosRequests"]["ChatMessageRequest"] = ChatMessageRequest;
})(this);