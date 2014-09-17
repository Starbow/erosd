(function (global) {
    "use strict";

    var Match = function(eros){
    	// var opponent,
    	// 	match_time,
    	// 	battle_net_channel,
    	// 	map,
    	// 	region

    	// From protobufs
    	var timespan,
    		quality,
    		opponent,
    		opponent_latency,
    		channel,
    		chat_room,
    		map,
    		long_unlock_time,
    		long_response_time

    	var match = this;
    }

    var MatchmakingModule = function(eros, options){
    	var queued = false,
    		matched = false,
    		match

    	function processServerMessage(command, payload) {
    		if (command == "MMQ" || command == "MMD"){
    			matched = false;
    			match = undefined;
    			queued = command == "MMQ" ? true : false;
    			options.queued(queued)

    			return true;
    		}else if(command == "MMR"){
    			queued = false
    			matched = true
    			match = protobufs.MatchmakingResult.decode64(payload)

    			return true;
    		}else if(command == "MMF"){
    			queued = false;
    			matched = false;
    			match = undefined;

    		}else{
    			return false;
    		}
    	}

    	this.commandHandlers = {
    		"MMQ": processServerMessage, // Queued
    		"MMD": processServerMessage, // Dequed
    		"MMR": processServerMessage, // Matched
    		"MMF": processServerMessage, // Forfeited
    	}

    	this.queue = function(regions, search_range){
    		sendRequest(new starbow.ErosRequests.MatchmakingQueueRequest)
    	}
    }

    global.starbow.Eros.prototype.modules.matchmaking = MatchmakingModule;

})(this);