(function (global) {
    "use strict";

    var HandshakeRequest = function (username, password, callback) {
        var handshake = new protobufs.Handshake(username, password);
        var request = new starbow.ErosRequests.Request("HSH", handshake.toBase64(), function (command, payload) {
            if (command === "HSH") {
                request.result = protobufs.HandshakeResponse.decode64(payload);
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
            }

            request.complete = true;
            return false;
        });



        return request;
    }


    var PingRequest = function (data, callback) {

        var start = +new Date();

        var request = new starbow.ErosRequests.Request("PNR", data, function (command, payload) {
            if (command === "PNR") {
                var end = +new Date();
                request.result = end - start;
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

    global["starbow"]["ErosRequests"]["HandshakeRequest"] = HandshakeRequest;
    global["starbow"]["ErosRequests"]["PingRequest"] = PingRequest;
})(this);