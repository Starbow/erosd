(function (global) {
    "use strict";

    var commands = {
    	queue: "MMQ",
    	forfeit: "MMF",
    	dequeue: "MMD",
    	result: "MMR",
        upload: "REP",
        lp_request: "RLP",
        lp_response: "LPR",
        auth_request: "BNN",
        remove_char: "BNR"
    };

    var MatchmakingQueueRequest = function(regions, search_range, callback){ // MMQ

    	var queue_options = new protobufs.MatchmakingQueue(regions, search_range);
    	var request = new starbow.ErosRequests.Request(commands.queue, queue_options.toBase64(), function(command, payload){
    		if(command === commands.queue){
                request.result = true;
                request.complete = true;

                console.log("Queued.");

                if (typeof (callback) === "function") {
                    callback(true, command);
                }

                return true;
    		}else if(command === commands["result"]){
                request.result = true;
                request.complete = true;

                var match;
                try{
                    match = protobufs.MatchmakingResult.decode64(payload);
                }catch(e){
                    console.error(e);
                }

                if (typeof (callback) === "function") {
                    callback(true, command, match);
                }

                return true;
    		}
    		return false;
    	}, function(command){ // Error handler
            request.result = true;
            request.complete = true;

            callback(false, command);

            return true;
        });

    	return request;
    };

    var MatchmakingDequeueRequest = function(callback){
    	var request = new starbow.ErosRequests.Request(commands["dequeue"],'', function(command, payload){
    		if(command === commands["dequeue"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, command);
                }

                return true;
            }
    	}, function(command){ // Error handler
            callback(false, command);

            request.result = true;
            request.complete = true;

           return true;
        });

    	return request;
    };

    var MatchmakingForfeitRequest = function(callback){
    	var request = new starbow.ErosRequests.Request(commands["forfeit"], '', function(command, payload){
            request.result = true;
            request.complete = true;

            if (typeof (callback) === "function") {
                callback(true, command);
            }
            return true;
    	}, function(command){ // Error handler
            request.result = true;
            request.complete = true;

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            return true;
        });

    	return request;
    };

    // This request will be a part of matchmaking until we get the Replay module up
    var MatchmakingUploadReplayRequest = function(file, callback){

        // var replay_file = window.btoa(unescape(encodeURIComponent(file)))
        var replay_file = file.split('base64,')[1];
        // window.btoa(unescape(encodeURIComponent(file)))

        var request = new starbow.ErosRequests.Request(commands["upload"], replay_file, function(command, payload){
            if(command === commands["replay"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request);
                }

                return true;
            }else if(command === commands["upload"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request);
                }

                return true;
            }
            return false;
        }, function(command){ // Error handler
            console.warn("Replay answer: "+command);

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            request.result = true;
            request.complete = true;

            return true;
        });

        return request;
    };

    var MatchmakingLongProcessRequest = function(process, callback){
        var request = new starbow.ErosRequests.Request(commands["lp_request"], process, function(command, payload){
            if(command === commands["lp_request"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request);
                }
                return true;
            }
        }, function(command){
            request.result = true;
            request.complete = true;
            
            console.warn("Long Process Request error: "+command);

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            return true;
        });

        return request;
    };

    var MatchmakingLongProcessResponse = function(process, callback){
        var request = new starbow.ErosRequests.Request(commands["lp_response"], process, function(command, payload){
            if(command === commands["lp_response"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request);
                }
                return true;
            }
        }, function(command){
            console.warn("Long Process Response error: "+command);
            callback(false, command);

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            return true;
        });

        return request;
    };

    var OAuthVerificationRequest = function(region, callback){
        var oauth_options = new protobufs.OAuthRequest(region);

        var request = new starbow.ErosRequests.Request(commands.auth_request, oauth_options.toBase64(), function(command, payload){
            if(command === commands["auth_request"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request, payload);
                }
                return true;
            }
        }, function(command, payload){
            console.warn("OAuth Request Error "+command+": "+window.atob(payload));
            callback(false, command);

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            return true;
        });

        return request;
    };

    var RemoveCharacterRequest = function(character, callback){
        character = new protobufs.Character(character)

        var request = new starbow.ErosRequests.Request(commands.remove_char, character.toBase64(), function(command, payload){
            if(command === commands["remove_char"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(true, request, payload);
                }
                return true;
            }
        }, function(command, payload){
            console.warn("Character Remove Request Error "+command+": "+window.atob(payload));
            callback(false, command);

            if (typeof (callback) === "function") {
                callback(false, command);
            }

            return true;
        });

        return request;
    }

    global["starbow"]["ErosRequests"]["MatchmakingQueueRequest"] = MatchmakingQueueRequest;
    global["starbow"]["ErosRequests"]["MatchmakingDequeueRequest"] = MatchmakingDequeueRequest;
    global["starbow"]["ErosRequests"]["MatchmakingForfeitRequest"] = MatchmakingForfeitRequest;
    global["starbow"]["ErosRequests"]["MatchmakingUploadReplayRequest"] = MatchmakingUploadReplayRequest;
    global["starbow"]["ErosRequests"]["MatchmakingLongProcessRequest"] = MatchmakingLongProcessRequest;
    global["starbow"]["ErosRequests"]["MatchmakingLongProcessResponse"] = MatchmakingLongProcessResponse;
    global["starbow"]["ErosRequests"]["OAuthVerificationRequest"] = OAuthVerificationRequest;
    global["starbow"]["ErosRequests"]["RemoveCharacterRequest"] = RemoveCharacterRequest;
})(this);