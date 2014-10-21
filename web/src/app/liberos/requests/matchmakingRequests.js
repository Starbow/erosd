(function (global) {
    "use strict";

    var commands = {
    	queue: "MMQ",
    	forefeit: "MMF",
    	dequeue: "MMD",
    	result: "MMR"
    }

    var MatchmakingQueueRequest = function(regions, search_range, callback){ // MMQ

    	queue_options = new protobufs.MatchmakingQueue(regions, search_range);
    	var request = new starbow.ErosRequests.Request(commands["queue"], queue_options, function(command, payload){
    		if(command === commands["queue"]){

    		}else if(command === commands["result"]){

    		}
    		return false;
    	})

    	return request;
    }

    var MatchmakingDequeueRequest = function(callback){
    	var request = new starbow.ErosRequests.Request(commands["dequeue"], queue_options, function(command, payload){
    		if(command === commands["dequeue"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
    		}
    		return false;
    	}
    	return request;
    }

    var MatchmakingForefeitRequest = function(callback){
    	var request = new starbow.ErosRequests.Request(commands["forefeit"], queue_options, function(command, payload){
    		if(command === commands["forefeit"]){
                request.result = true;
                request.complete = true;

                if (typeof (callback) === "function") {
                    callback(request);
                }

                return true;
    		}
    		return false;
    	}

    	return request;
    }

    global["starbow"]["ErosRequests"]["MatchmakingQueueRequest"] = MatchmakingQueueRequest;
    global["starbow"]["ErosRequests"]["MatchmakingDequeueRequest"] = MatchmakingDequeueRequest;
    global["starbow"]["ErosRequests"]["MatchmakingForefeitRequest"] = MatchmakingForefeitRequest;
})(this);